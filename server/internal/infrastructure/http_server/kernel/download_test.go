package kernel

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteDownload_SetsForcedDownloadHeadersAndBody(t *testing.T) {
	const body = "file-bytes"
	rec := httptest.NewRecorder()

	WriteDownload(rec, "quarterly report.pdf", "application/pdf", int64(len(body)), io.NopCloser(strings.NewReader(body)))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
	assert.Equal(t, "10", rec.Header().Get("Content-Length"))
	assert.Equal(t, "private, no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, body, rec.Body.String())

	disposition := rec.Header().Get("Content-Disposition")
	assert.Contains(t, disposition, "attachment;")
	assert.Contains(t, disposition, "filename*=UTF-8''")
}

func TestWriteDownload_DefaultsMissingMediaType(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteDownload(rec, "a.bin", "", 0, io.NopCloser(strings.NewReader("")))

	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}

func TestAttachmentDisposition(t *testing.T) {
	t.Run("ascii filename", func(t *testing.T) {
		assert.Equal(t, `attachment; filename="receipt.pdf"; filename*=UTF-8''receipt.pdf`, attachmentDisposition("receipt.pdf"))
	})

	t.Run("quotes and non-ascii are escaped in the fallback", func(t *testing.T) {
		got := attachmentDisposition(`a"b.pdf`)
		require.NotContains(t, got, `filename="a"b`)
	})
}
