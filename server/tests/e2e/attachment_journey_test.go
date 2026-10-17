//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/tests/utils"
)

// TestAttachmentJourney exercises the whole attachment surface against real
// Postgres and real MinIO: upload-before-create, get/download as a draft,
// atomic create-and-link, post-link standalone scope, attach-later, the direct
// nested upload, and deletion.
func TestAttachmentJourney(t *testing.T) {
	token := apiClient.Login(t, seededUser.Username, seededPassword)

	var currencyResp utils.Envelope[schema.CurrencyResponse]
	status := apiClient.Do(t, http.MethodPost, "/api/v1/currencies", token, schema.CurrencyCreateRequest{
		Code: "ZEA", Name: "E2E Attachment Currency", IsBase: false,
	}, &currencyResp)
	require.Equal(t, http.StatusCreated, status)

	var categoryResp utils.Envelope[schema.TransactionCategoryResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transaction-categories", token, schema.TransactionCategoryCreateRequest{
		Type: domain.TransactionCategoryTypeExpense, Name: "E2E Attachment Category", SortOrder: 1,
	}, &categoryResp)
	require.Equal(t, http.StatusCreated, status)

	// 1. Upload a draft before any transaction exists.
	draftBytes := []byte("%PDF-1.7 e2e attachment bytes")
	resp, raw := apiClient.UploadFile(t, "/api/v1/transaction-attachments", token, "receipt.pdf", "application/pdf", draftBytes)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "upload draft body: %s", raw)
	var draft utils.Envelope[schema.TransactionAttachmentResponse]
	require.NoError(t, json.Unmarshal(raw, &draft))
	draftID := draft.Data.ID

	// 2. Read and download it while it is still a draft.
	var got utils.Envelope[schema.TransactionAttachmentResponse]
	status = apiClient.Do(t, http.MethodGet, "/api/v1/transaction-attachments/"+draftID.String(), token, nil, &got)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, got.Data.TransactionID)
	assert.Equal(t, int64(len(draftBytes)), got.Data.ByteSize)

	status, body, _ := apiClient.Download(t, "/api/v1/transaction-attachments/"+draftID.String()+"/content", token)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, draftBytes, body)

	// 3. Create a transaction that links the draft atomically.
	var txResp utils.Envelope[schema.TransactionResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transactions", token, schema.TransactionCreateRequest{
		Type:          domain.TransactionTypeExpense,
		Amount:        domain.Money(4250),
		CurrencyCode:  "ZEA",
		CategoryID:    categoryResp.Data.ID,
		UserID:        seededUser.ID,
		Source:        domain.TransactionSourceManual,
		Name:          "E2E attachment transaction",
		AttachmentIDs: []uuid.UUID{draftID},
	}, &txResp)
	require.Equal(t, http.StatusCreated, status)
	txID := txResp.Data.ID
	require.Len(t, txResp.Data.Attachments, 1)
	require.NotNil(t, txResp.Data.Attachments[0].TransactionID)
	assert.Equal(t, txID, *txResp.Data.Attachments[0].TransactionID)

	// 4. Standalone access to the now-linked attachment is closed.
	status = apiClient.Do(t, http.MethodGet, "/api/v1/transaction-attachments/"+draftID.String(), token, nil, &utils.Envelope[struct{}]{})
	assert.Equal(t, http.StatusNotFound, status)

	// 5. Upload a second draft and attach it later.
	resp, raw = apiClient.UploadFile(t, "/api/v1/transaction-attachments", token, "note.txt", "text/plain", []byte("later"))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var secondDraft utils.Envelope[schema.TransactionAttachmentResponse]
	require.NoError(t, json.Unmarshal(raw, &secondDraft))

	status = apiClient.Do(t, http.MethodPut,
		fmt.Sprintf("/api/v1/transactions/%d/attachments/%s", txID, secondDraft.Data.ID), token, nil, nil)
	require.Equal(t, http.StatusNoContent, status)

	// 6. The nested list reports both attachments.
	var list utils.Envelope[[]schema.TransactionAttachmentResponse]
	status = apiClient.Do(t, http.MethodGet, fmt.Sprintf("/api/v1/transactions/%d/attachments", txID), token, nil, &list)
	require.Equal(t, http.StatusOK, status)
	assert.Len(t, list.Data, 2)

	// 7. The direct transaction-first upload path still works.
	resp, raw = apiClient.UploadFile(t, fmt.Sprintf("/api/v1/transactions/%d/attachments", txID), token, "direct.png", "image/png", []byte("png"))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var direct utils.Envelope[schema.TransactionAttachmentResponse]
	require.NoError(t, json.Unmarshal(raw, &direct))
	require.NotNil(t, direct.Data.TransactionID)
	assert.Equal(t, txID, *direct.Data.TransactionID)

	// 8. Nested download returns the exact bytes.
	status, body, _ = apiClient.Download(t, fmt.Sprintf("/api/v1/transactions/%d/attachments/%s/content", txID, direct.Data.ID), token)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []byte("png"), body)

	// 9. Deleting the transaction explicitly removes its attachments first.
	status = apiClient.Do(t, http.MethodDelete, fmt.Sprintf("/api/v1/transactions/%d", txID), token, nil, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = apiClient.Do(t, http.MethodGet, fmt.Sprintf("/api/v1/transactions/%d", txID), token, nil, &utils.Envelope[struct{}]{})
	assert.Equal(t, http.StatusNotFound, status)
}
