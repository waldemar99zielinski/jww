package kernel

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/service/object_storage"
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

func TestParseSingleByteRange(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  *object_storage.ByteRange
	}{
		{name: "none", value: "", want: nil},
		{name: "closed", value: "bytes=2-5", want: &object_storage.ByteRange{Start: 2, End: 5}},
		{name: "open", value: "bytes=7-", want: &object_storage.ByteRange{Start: 7, End: 9}},
		{name: "suffix", value: "bytes=-3", want: &object_storage.ByteRange{Start: 7, End: 9}},
		{name: "clamps end", value: "bytes=8-99", want: &object_storage.ByteRange{Start: 8, End: 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSingleByteRange(tt.value, 10)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	for _, value := range []string{"items=0-1", "bytes=", "bytes=10-", "bytes=5-2", "bytes=0-1,3-4"} {
		t.Run("rejects "+value, func(t *testing.T) {
			_, err := ParseSingleByteRange(value, 10)
			assert.ErrorIs(t, err, ErrInvalidRange)
		})
	}
}

func TestWriteGrantedContent_PartialPreview(t *testing.T) {
	rec := httptest.NewRecorder()
	byteRange := &object_storage.ByteRange{Start: 2, End: 4}

	WriteGrantedContent(rec, "receipt.pdf", "application/pdf", 10, true, byteRange, io.NopCloser(strings.NewReader("cde")))

	assert.Equal(t, http.StatusPartialContent, rec.Code)
	assert.Equal(t, "bytes 2-4/10", rec.Header().Get("Content-Range"))
	assert.Equal(t, "3", rec.Header().Get("Content-Length"))
	assert.Equal(t, "bytes", rec.Header().Get("Accept-Ranges"))
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "inline;")
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "cde", rec.Body.String())
}
