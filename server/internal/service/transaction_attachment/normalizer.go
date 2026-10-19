package transaction_attachment

import (
	"io"

	"jww/internal/domain"
)

// NormalizedContent is the filename, media type, and bytes that should actually
// be persisted. A normalizer may replace the source format (for example
// HEIC/HEIF with PNG) while leaving every other upload untouched.
type NormalizedContent struct {
	Filename  string
	MediaType string
	Body      io.Reader
}

// ContentNormalizer prepares an upload for storage. Implementations must only
// read body when conversion is required; otherwise they return it unchanged so
// ordinary uploads remain streamed.
type ContentNormalizer interface {
	Normalize(filename, mediaType string, body io.Reader) (NormalizedContent, error)
}

func (s *Service) normalizeContent(filename, mediaType string, body io.Reader) (NormalizedContent, error) {
	if s.normalizer == nil {
		return NormalizedContent{Filename: filename, MediaType: mediaType, Body: body}, nil
	}
	content, err := s.normalizer.Normalize(filename, mediaType, body)
	if err != nil {
		return NormalizedContent{}, domain.NewValidationError("attachment image could not be converted")
	}
	return content, nil
}
