package transaction_attachment

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
)

// extractionHandler runs a draft-only vision extraction for an unattached
// attachment and returns the proposed prefill values. Nothing is persisted;
// kernel.WriteError maps the service's domain errors to 404 (draft
// missing/already linked), 400 (non-image media), or 503 (provider failure).
func extractionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		// A vision extraction may run for minutes (the provider's own timeout is
		// the real bound), so clear the server's short response-write deadline
		// after request validation has passed.
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		start := time.Now()
		prefill, err := deps.TransactionAttachmentExtractionService.Extract(appCtx, attachmentID)
		elapsed := time.Since(start)
		if err != nil {
			logExtractionFailure(appCtx, attachmentID, elapsed, err)
			kernel.WriteError(w, r, err)
			return
		}

		log.WithContext(appCtx).Info("attachment extraction completed",
			"attachment_id", attachmentID.String(),
			"duration_ms", elapsed.Milliseconds(),
			"sub_transaction_count", len(prefill.SubTransactions),
			"warning_count", len(prefill.Warnings),
		)
		kernel.WriteOK(w, prefillToResponse(prefill))
	}
}

type batchExtractionRequest struct {
	AttachmentIDs []uuid.UUID `json:"attachment_ids"`
}

// batchExtractionHandler reads several draft images in one vision-model call.
// The ordered IDs represent pages/views of one logical receipt.
func batchExtractionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)
		request, err := kernel.DecodeJSON[batchExtractionRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		// A vision extraction may run for minutes (the provider's own timeout is
		// the real bound), so clear the server's short response-write deadline
		// after request validation has passed.
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		start := time.Now()
		prefill, err := deps.TransactionAttachmentExtractionService.ExtractMany(appCtx, request.AttachmentIDs)
		elapsed := time.Since(start)
		if err != nil {
			log.WithContext(appCtx).Error("batch attachment extraction failed",
				"attachment_count", len(request.AttachmentIDs),
				"duration_ms", elapsed.Milliseconds(),
				"error", err.Error(),
			)
			kernel.WriteError(w, r, err)
			return
		}
		log.WithContext(appCtx).Info("batch attachment extraction completed",
			"attachment_count", len(prefill.AttachmentIDs),
			"duration_ms", elapsed.Milliseconds(),
			"sub_transaction_count", len(prefill.SubTransactions),
		)
		kernel.WriteOK(w, prefillToResponse(prefill))
	}
}

// logExtractionFailure logs one failed extraction at the boundary, the only
// place that should: the service returns a typed error and stays silent. Level
// follows the error's transport kind so a client-caused rejection (unknown or
// linked draft, non-image media) is a Warn rather than an Error — the kernel's
// RequestLogger already records the resulting status code, and this line adds
// the attachment id, duration, and (for wrapped failures) the underlying
// provider/storage cause that never reaches the client.
func logExtractionFailure(appCtx domain.AppContext, attachmentID uuid.UUID, elapsed time.Duration, err error) {
	attrs := []any{
		"attachment_id", attachmentID.String(),
		"duration_ms", elapsed.Milliseconds(),
		"error", err.Error(),
		"error_type", fmt.Sprintf("%T", err),
	}
	if cause := errors.Unwrap(err); cause != nil {
		attrs = append(attrs, "cause", cause.Error())
	}

	boundary := log.WithContext(appCtx)
	var domainErr domain.DomainError
	if errors.As(err, &domainErr) && isClientCausedKind(domainErr.Kind()) {
		boundary.Warn("attachment extraction rejected", attrs...)
		return
	}
	boundary.Error("attachment extraction failed", attrs...)
}

// isClientCausedKind reports whether an error kind is the caller's problem
// rather than ours, matching the kernel's status mapping (4xx vs 5xx).
func isClientCausedKind(kind domain.ErrorKind) bool {
	switch kind {
	case domain.ErrorKindNotFound, domain.ErrorKindConflict,
		domain.ErrorKindValidation, domain.ErrorKindUnauthorized:
		return true
	default:
		return false
	}
}

// prefillToResponse maps the service-level Prefill into its wire DTO. The
// converter lives here, at the HTTP boundary, rather than in the schema
// package: schema must not import service packages (see adapter/schema's
// README), so the package that already depends on both owns the mapping.
// SubTransactions is built non-nil (empty when nothing was itemized) because
// the DTO field has no omitempty and would otherwise serialize as null.
func prefillToResponse(prefill *transaction_attachment_extraction.Prefill) schema.TransactionPrefillResponse {
	response := schema.TransactionPrefillResponse{
		AttachmentID:    prefill.AttachmentID,
		AttachmentIDs:   append([]uuid.UUID(nil), prefill.AttachmentIDs...),
		Type:            prefillFieldToResponse(prefill.Type),
		Amount:          prefillFieldToResponse(prefill.Amount),
		CurrencyCode:    prefillFieldToResponse(prefill.CurrencyCode),
		OccurredOn:      prefillFieldToResponse(prefill.OccurredOn),
		OccurredAt:      prefillFieldToResponse(prefill.OccurredAt),
		Name:            prefillFieldToResponse(prefill.Name),
		Description:     prefillFieldToResponse(prefill.Description),
		CategoryHint:    prefillFieldToResponse(prefill.CategoryHint),
		SubTransactions: make([]schema.PrefillSubTransactionResponse, 0, len(prefill.SubTransactions)),
		Warnings:        prefill.Warnings,
		DurationMs:      prefill.DurationMs,
	}
	for _, sub := range prefill.SubTransactions {
		response.SubTransactions = append(response.SubTransactions, schema.PrefillSubTransactionResponse{
			Name:       sub.Name,
			Label:      sub.Label,
			Amount:     sub.Amount,
			Confidence: sub.Confidence,
		})
	}
	return response
}

// prefillFieldToResponse maps one proposed scalar, preserving an absent field
// (nil) as nil so the DTO omits it.
func prefillFieldToResponse[T any](field *transaction_attachment_extraction.PrefillField[T]) *schema.PrefillFieldResponse[T] {
	if field == nil {
		return nil
	}
	return &schema.PrefillFieldResponse[T]{
		Value:      field.Value,
		Confidence: field.Confidence,
		Source:     field.Source,
	}
}
