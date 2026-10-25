package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/openai"
	"jww/internal/domain"
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
)

func appContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

func newClient(baseURL string) *openai.Client {
	return openai.NewClient(openai.Config{
		BaseURL:        baseURL,
		APIKey:         "test-api-key",
		Model:          "gpt-4.1-mini",
		TimeoutSeconds: 5,
	})
}

func TestClient_Check(t *testing.T) {
	var path, authorization string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"gpt-4.1-mini","object":"model","created":0,"owned_by":"test"}]}`))
	}))
	defer srv.Close()

	err := newClient(srv.URL).Check(appContext())

	require.NoError(t, err)
	assert.Equal(t, "/models", path)
	assert.Equal(t, "Bearer test-api-key", authorization)
}

func TestClient_CheckConfiguredModelMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"another-model","object":"model","created":0,"owned_by":"test"}]}`))
	}))
	defer srv.Close()

	err := newClient(srv.URL).Check(appContext())

	require.EqualError(t, err, `openai: configured model "gpt-4.1-mini" is not available`)
}

// wireRequest is the subset of the OpenAI Chat Completions request body these
// tests pin: temperature 0, a single user message carrying one instruction
// text part plus one image_url content part, and json_schema structured output.
type wireRequest struct {
	Model          string   `json:"model"`
	Temperature    *float64 `json:"temperature"`
	ResponseFormat struct {
		Type       string `json:"type"`
		JSONSchema struct {
			Name   string         `json:"name"`
			Schema map[string]any `json:"schema"`
			Strict *bool          `json:"strict"`
		} `json:"json_schema"`
	} `json:"response_format"`
	Messages []struct {
		Role    string `json:"role"`
		Content []struct {
			Type     string `json:"type"`
			Text     string `json:"text"`
			ImageURL *struct {
				URL string `json:"url"`
			} `json:"image_url"`
		} `json:"content"`
	} `json:"messages"`
}

// prefillJSON is a well-formed structured-output payload matching the schema
// the adapter forces (every scalar nullable, sub_transactions an array).
const prefillJSON = `{"type":{"value":"expense","confidence":96},"amount":{"value":"24.90","confidence":92},"currency_code":{"value":"USD","confidence":99},"occurred_on":{"value":"2026-10-12","confidence":88},"occurred_at":{"value":"14:35","confidence":74},"name":{"value":"Corner Market","confidence":90},"description":null,"category_hint":{"value":"Groceries","confidence":81},"sub_transactions":[{"name":"Milk 2%","label":"dairy","amount":"3.49","confidence":93},{"name":"Coupon","label":"discount","amount":"-1.00","confidence":85}]}`

func completionResponse(finishReason, content string) string {
	body, _ := json.Marshal(map[string]any{
		"id":      "chatcmpl-test",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   "gpt-4.1-mini",
		"choices": []map[string]any{
			{
				"index":         0,
				"finish_reason": finishReason,
				"message":       map[string]any{"role": "assistant", "content": content},
			},
		},
	})
	return string(body)
}

func TestClient_ExtractTransactionPrefill_SendsVisionMessageWithStructuredOutput(t *testing.T) {
	var captured wireRequest
	var authHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(completionResponse("stop", prefillJSON)))
	}))
	defer srv.Close()

	const imageURL = "https://api.example.com/v1/attachment-content?grant=abc%2Bdef"
	const secondImageURL = "https://api.example.com/v1/attachment-content?grant=second"
	result, err := newClient(srv.URL).ExtractTransactionPrefill(appContext(), transaction_attachment_extraction.TransactionPrefillInput{
		ImageURLs:         []string{imageURL, secondImageURL},
		ExpenseCategories: []string{"Groceries", "Dining out"},
		IncomeCategories:  []string{"Salary"},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Amount)
	assert.Equal(t, domain.Money(2490), result.Amount.Value)
	require.NotNil(t, result.OccurredAt)
	assert.Equal(t, "14:35", result.OccurredAt.Value)
	require.NotNil(t, result.OccurredAt.Confidence)
	assert.Equal(t, 74.0, *result.OccurredAt.Confidence)
	require.Len(t, result.SubTransactions, 2)
	require.NotNil(t, result.SubTransactions[1].Confidence)
	assert.Equal(t, domain.Money(-100), result.SubTransactions[1].Amount)
	assert.Equal(t, 85.0, *result.SubTransactions[1].Confidence)

	assert.Equal(t, "Bearer test-api-key", authHeader, "the API key is sent as a bearer token")
	assert.Equal(t, "gpt-4.1-mini", captured.Model)
	require.NotNil(t, captured.Temperature)
	assert.Equal(t, float64(0), *captured.Temperature, "determinism: temperature must be 0")

	assert.Equal(t, "json_schema", captured.ResponseFormat.Type)
	assert.NotEmpty(t, captured.ResponseFormat.JSONSchema.Name)
	assert.NotNil(t, captured.ResponseFormat.JSONSchema.Schema)
	require.NotNil(t, captured.ResponseFormat.JSONSchema.Strict)
	assert.True(t, *captured.ResponseFormat.JSONSchema.Strict)

	require.Len(t, captured.Messages, 1, "exactly one user message")
	assert.Equal(t, "user", captured.Messages[0].Role)

	var sawText bool
	var imageURLs []string
	for _, part := range captured.Messages[0].Content {
		switch part.Type {
		case "text":
			sawText = true
			assert.NotEmpty(t, part.Text)
		case "image_url":
			require.NotNil(t, part.ImageURL)
			imageURLs = append(imageURLs, part.ImageURL.URL)
		}
	}
	assert.True(t, sawText, "the message carries an instruction text part")
	assert.Equal(t, []string{imageURL, secondImageURL}, imageURLs, "every supplied image URL is sent once and in order")
	assert.Contains(t, captured.Messages[0].Content[0].Text, `expense_categories: ["Groceries","Dining out"]`)
	assert.Contains(t, captured.Messages[0].Content[0].Text, `income_categories: ["Salary"]`)
	assert.Contains(t, captured.Messages[0].Content[0].Text, "occurred_at")
	assert.Contains(t, captured.Messages[0].Content[0].Text, "confidence integer from 0 to 100")
	assert.Contains(t, captured.Messages[0].Content[0].Text, "Be sincere and truthful")
	assert.Contains(t, captured.Messages[0].Content[0].Text, "honest low score is better than false certainty")
	assert.Contains(t, captured.Messages[0].Content[0].Text, "negative amount")
	assert.Contains(t, captured.Messages[0].Content[0].Text, "suitable for grouping and later analytics")
	assert.Contains(t, captured.Messages[0].Content[0].Text, `chicken or beef as "meat"`)
	assert.Contains(t, captured.Messages[0].Content[0].Text, `apples or bananas as "fruits"`)
	assert.Contains(t, captured.Messages[0].Content[0].Text, `carrots or spinach as "vegetables"`)
	assert.Contains(t, captured.Messages[0].Content[0].Text, `not repeat or lightly reword the item name`)
	properties, ok := captured.ResponseFormat.JSONSchema.Schema["properties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, properties, "occurred_at")
}

func TestClient_ExtractTransactionPrefill_RejectsConfidenceOutsideRange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		outOfRange := strings.Replace(prefillJSON, `"confidence":96`, `"confidence":101`, 1)
		_, _ = w.Write([]byte(completionResponse("stop", outOfRange)))
	}))
	defer srv.Close()

	_, err := newClient(srv.URL).ExtractTransactionPrefill(appContext(), transaction_attachment_extraction.TransactionPrefillInput{
		ImageURLs: []string{"https://example.com/img"},
	})

	require.ErrorContains(t, err, "outside 0-100")
}

func TestClient_ExtractTransactionPrefill_ErrorMapping(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		finishReason string
		content      string
	}{
		{name: "non-2xx", status: http.StatusInternalServerError, finishReason: "stop", content: prefillJSON},
		{name: "rate limited", status: http.StatusTooManyRequests, finishReason: "stop", content: prefillJSON},
		{name: "content filtered", status: http.StatusOK, finishReason: "content_filter", content: ""},
		{name: "empty structured output", status: http.StatusOK, finishReason: "stop", content: ""},
		{name: "malformed structured output", status: http.StatusOK, finishReason: "stop", content: "{not valid json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if tt.status != http.StatusOK {
					w.WriteHeader(tt.status)
					_, _ = w.Write([]byte(`{"error":{"message":"boom"}}`))
					return
				}
				_, _ = w.Write([]byte(completionResponse(tt.finishReason, tt.content)))
			}))
			defer srv.Close()

			_, err := newClient(srv.URL).ExtractTransactionPrefill(appContext(), transaction_attachment_extraction.TransactionPrefillInput{ImageURLs: []string{"https://example.com/img"}})

			require.Error(t, err, "every refusal/garbage/outage case must return an error the service maps to 503")
			if tt.status != http.StatusOK {
				assert.ErrorContains(t, err, `message="boom"`, "provider diagnostics remain available to internal logs")
			}
		})
	}
}

// TestClient_ExtractTransactionPrefill_BaseURLSelectsBackend proves the portability goal: the only
// difference between two "compatible backends" is the configured base URL.
func TestClient_ExtractTransactionPrefill_BaseURLSelectsBackend(t *testing.T) {
	var firstHits, secondHits int
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(completionResponse("stop", prefillJSON)))
	}))
	defer first.Close()
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(completionResponse("stop", prefillJSON)))
	}))
	defer second.Close()

	req := transaction_attachment_extraction.TransactionPrefillInput{ImageURLs: []string{"https://example.com/img"}}
	_, err := newClient(first.URL).ExtractTransactionPrefill(appContext(), req)
	require.NoError(t, err)
	_, err = newClient(second.URL).ExtractTransactionPrefill(appContext(), req)
	require.NoError(t, err)

	assert.Equal(t, 1, firstHits)
	assert.Equal(t, 1, secondHits)
}

// TestClient_ExtractTransactionPrefill_TimeoutIsBounded proves the adapter never hangs: an
// unresponsive backend fails within the configured timeout instead of blocking
// the request forever.
func TestClient_ExtractTransactionPrefill_TimeoutIsBounded(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer srv.Close()
	defer close(release)

	client := openai.NewClient(openai.Config{
		BaseURL:        srv.URL,
		APIKey:         "test-api-key",
		Model:          "gpt-4.1-mini",
		TimeoutSeconds: 1,
	})

	done := make(chan error, 1)
	go func() {
		_, err := client.ExtractTransactionPrefill(appContext(), transaction_attachment_extraction.TransactionPrefillInput{ImageURLs: []string{"https://example.com/img"}})
		done <- err
	}()

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Read must return within its configured timeout")
	}
}
