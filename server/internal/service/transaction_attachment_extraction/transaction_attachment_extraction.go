package transaction_attachment_extraction

import (
	"mime"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/utils/logger"
)

// log carries the use case's step-by-step trace. Every line here is Debug:
// these are diagnostic details of a happy path, off by default and enabled
// with LOG_LEVEL=debug. The boundary (http_server) owns the Info/Warn/Error
// summary of the outcome — see extraction_handler.go.
var log = logger.New().WithComponent("service.transaction_attachment_extraction")

// contentPath is the full public path of the attachment-content grant endpoint,
// including the "/api" mount prefix. It mirrors the route registration in
// http_server.NewRouter: transaction_attachment.Mount registers the public
// `GET /v1/attachment-content` on the sub-router mounted under `r.Route("/api", …)`,
// so the URL a provider must fetch is PUBLIC_API_BASE_URL + this path. Keep the
// two in sync — a mismatch means the provider fetches a non-routed path and
// every extraction fails with a 503.
const (
	contentPath         = "/api/v1/attachment-content"
	maxExtractionImages = 10
)

// Prefill is the service-level extraction result: proposed transaction fields
// and line items plus the round-trip duration. It is never persisted — the
// client decides, per field and per item, what to accept.
type Prefill struct {
	AttachmentID    uuid.UUID
	AttachmentIDs   []uuid.UUID
	Type            *PrefillField[domain.TransactionType]
	Amount          *PrefillField[domain.Money] // integer cents
	CurrencyCode    *PrefillField[string]
	OccurredOn      *PrefillField[string]
	OccurredAt      *PrefillField[string] // HH:MM, optional
	Name            *PrefillField[string]
	Description     *PrefillField[string]
	CategoryHint    *PrefillField[string]
	SubTransactions []PrefillSubTransaction
	Warnings        []string
	DurationMs      int64
}

// Service orchestrates a draft-only extraction: authorize the draft, mint a
// short-lived content grant, hand our own absolute content URL to the
// provider-agnostic Extractor, and map its result. No attachment bytes flow
// through this service.
type Service struct {
	extractor  Extractor
	source     AttachmentSource
	categories CategorySource
	config     Config
}

// NewService takes the extractor, attachment source, and category source. All
// three are injected ports so this package depends on no adapter.
func NewService(extractor Extractor, source AttachmentSource, categories CategorySource, config Config) *Service {
	return &Service{extractor: extractor, source: source, categories: categories, config: config}
}

// Extract reads a draft attachment and returns proposed prefill values. The
// order of checks is deliberate:
//
//  1. load the draft (a missing/linked draft is not found, and then no grant
//     is minted and the extractor is never called),
//  2. reject non-image media BEFORE minting a grant or calling the provider,
//  3. mint a short-lived single-attachment content grant,
//  4. build our absolute content URL and hand it to the extractor,
//  5. map the extractor result, or turn any extractor failure into the 503 kind.
func (s *Service) Extract(ctx domain.AppContext, attachmentID uuid.UUID) (*Prefill, error) {
	return s.ExtractMany(ctx, []uuid.UUID{attachmentID})
}

// ExtractMany reads several draft images as pages/evidence for one logical
// transaction and makes exactly one extractor call containing all image URLs.
func (s *Service) ExtractMany(ctx domain.AppContext, attachmentIDs []uuid.UUID) (*Prefill, error) {
	trace := log.WithContext(ctx)
	if len(attachmentIDs) == 0 {
		return nil, domain.NewValidationError("at least one attachment id is required")
	}
	if len(attachmentIDs) > maxExtractionImages {
		return nil, domain.NewValidationError("at most 10 attachment ids are allowed")
	}
	trace.Debug("extraction started", "attachment_count", len(attachmentIDs))
	seen := make(map[uuid.UUID]struct{}, len(attachmentIDs))

	// Validate the list itself before touching storage.
	for _, attachmentID := range attachmentIDs {
		if attachmentID == uuid.Nil {
			return nil, domain.NewValidationError("attachment ids must not be empty")
		}
		if _, duplicate := seen[attachmentID]; duplicate {
			return nil, domain.NewValidationError("attachment ids must be unique")
		}
		seen[attachmentID] = struct{}{}
	}

	// Validate the complete group before minting any grant. This keeps linked,
	// missing, or non-image groups from exposing even a subset of files.
	for _, attachmentID := range attachmentIDs {
		attachment, err := s.source.DraftAttachment(ctx, attachmentID)
		if err != nil {
			return nil, err
		}
		if !isImageMediaType(attachment.MediaType) {
			return nil, domain.NewUnsupportedAttachmentMediaTypeError(attachment.MediaType)
		}
	}
	expenseCategories, incomeCategories, err := s.categories.TransactionCategories(ctx)
	if err != nil {
		return nil, err
	}

	contentURLs := make([]string, 0, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		token, err := s.source.ContentGrant(ctx, attachmentID)
		if err != nil {
			return nil, err
		}
		contentURLs = append(contentURLs, strings.TrimRight(s.config.PublicBaseURL, "/")+contentPath+"?grant="+url.QueryEscape(token))
	}

	start := time.Now()
	result, err := s.extractor.ExtractTransactionPrefill(ctx, TransactionPrefillInput{
		ImageURLs:         contentURLs,
		ExpenseCategories: expenseCategories,
		IncomeCategories:  incomeCategories,
	})
	elapsed := time.Since(start)
	if err != nil {
		trace.Debug("extractor failed",
			"attachment_count", len(attachmentIDs),
			"duration_ms", elapsed.Milliseconds(),
			"error", err.Error(),
		)
		return nil, domain.NewAttachmentExtractionFailedError(err)
	}

	prefill := toPrefill(attachmentIDs, result, elapsed)
	trace.Debug("extractor returned",
		"attachment_count", len(attachmentIDs),
		"duration_ms", prefill.DurationMs,
		"result", result,
	)
	return prefill, nil
}

// toPrefill maps the extractor's result into the service-level Prefill. A nil
// result (provider returned nothing usable) yields a Prefill carrying only the
// attachment id, never a panic.
func toPrefill(attachmentIDs []uuid.UUID, result *TransactionPrefillResult, elapsed time.Duration) *Prefill {
	prefill := &Prefill{
		AttachmentID:  attachmentIDs[0],
		AttachmentIDs: append([]uuid.UUID(nil), attachmentIDs...),
		DurationMs:    elapsed.Milliseconds(),
	}
	if result == nil {
		return prefill
	}
	prefill.Type = result.Type
	prefill.Amount = result.Amount
	prefill.CurrencyCode = result.CurrencyCode
	prefill.OccurredOn = result.OccurredOn
	prefill.OccurredAt = result.OccurredAt
	prefill.Name = result.Name
	prefill.Description = result.Description
	prefill.CategoryHint = result.CategoryHint
	prefill.SubTransactions = result.SubTransactions
	prefill.Warnings = result.Warnings
	return prefill
}

// isImageMediaType accepts only a parseable image/* media type. It is stricter
// than transaction_attachment's previewableMediaType on purpose: extraction is
// images-only in v1, so application/pdf (previewable elsewhere) is unsupported
// here.
func isImageMediaType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(mediaType), "image/")
}
