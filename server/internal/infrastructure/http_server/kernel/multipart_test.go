package kernel

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMultipartRequest builds a multipart/form-data request from a function that
// writes the parts.
func newMultipartRequest(t *testing.T, build func(*multipart.Writer)) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	build(writer)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transaction-attachments", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func writeFilePart(t *testing.T, w *multipart.Writer, name, filename, contentType, content string) {
	t.Helper()
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="`+name+`"; filename="`+filename+`"`)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := w.CreatePart(header)
	require.NoError(t, err)
	_, err = io.WriteString(part, content)
	require.NoError(t, err)
}

func TestReadSingleMultipartFile(t *testing.T) {
	t.Run("valid single file part", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			writeFilePart(t, w, "file", "receipt.pdf", "application/pdf", "%PDF-1.7 data")
		})

		file, err := ReadSingleMultipartFile(req)
		require.NoError(t, err)
		assert.Equal(t, "receipt.pdf", file.Filename)
		assert.Equal(t, "application/pdf", file.MediaType)

		body, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "%PDF-1.7 data", string(body))
		require.NoError(t, file.Finish())
	})

	t.Run("missing content type defaults to octet-stream", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			writeFilePart(t, w, "file", "blob.bin", "", "bytes")
		})

		file, err := ReadSingleMultipartFile(req)
		require.NoError(t, err)
		assert.Equal(t, "application/octet-stream", file.MediaType)
	})

	t.Run("a generic or empty type is inferred from the extension", func(t *testing.T) {
		tests := []struct {
			name        string
			filename    string
			contentType string
			want        string
		}{
			{"dng, no type", "IMG_0001.DNG", "", "image/dng"},
			{"dng, octet-stream", "IMG_0001.dng", "application/octet-stream", "image/dng"},
			{"heic, no type", "IMG_0002.heic", "", "image/heic"},
			{"heif, octet-stream", "IMG_0003.heif", "application/octet-stream", "image/heif"},
			{"jpeg, no type", "receipt.jpeg", "", "image/jpeg"},
			{"pdf, octet-stream", "statement.pdf", "application/octet-stream", "application/pdf"},
			{"unknown extension stays octet-stream", "archive.zip", "", "application/octet-stream"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := newMultipartRequest(t, func(w *multipart.Writer) {
					writeFilePart(t, w, "file", tt.filename, tt.contentType, "bytes")
				})

				file, err := ReadSingleMultipartFile(req)
				require.NoError(t, err)
				assert.Equal(t, tt.want, file.MediaType)
			})
		}
	})

	t.Run("an explicit type is never overridden by the extension", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			writeFilePart(t, w, "file", "photo.dng", "image/jpeg", "bytes")
		})

		file, err := ReadSingleMultipartFile(req)
		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", file.MediaType)
	})

	t.Run("path components are stripped from the filename", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			writeFilePart(t, w, "file", "../../etc/passwd", "text/plain", "x")
		})

		file, err := ReadSingleMultipartFile(req)
		require.NoError(t, err)
		assert.Equal(t, "passwd", file.Filename)
	})

	t.Run("wrong part name is rejected", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			writeFilePart(t, w, "upload", "a.pdf", "application/pdf", "x")
		})

		_, err := ReadSingleMultipartFile(req)
		assert.Error(t, err)
	})

	t.Run("an extra part after the file fails Finish", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			writeFilePart(t, w, "file", "a.pdf", "application/pdf", "x")
			writeFilePart(t, w, "file", "b.pdf", "application/pdf", "y")
		})

		file, err := ReadSingleMultipartFile(req)
		require.NoError(t, err)
		_, _ = io.ReadAll(file)
		assert.Error(t, file.Finish())
	})

	t.Run("a leading non-file field is rejected", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {
			require.NoError(t, w.WriteField("kind", "receipt"))
			writeFilePart(t, w, "file", "a.pdf", "application/pdf", "x")
		})

		_, err := ReadSingleMultipartFile(req)
		assert.Error(t, err)
	})

	t.Run("empty body is rejected", func(t *testing.T) {
		req := newMultipartRequest(t, func(w *multipart.Writer) {})

		_, err := ReadSingleMultipartFile(req)
		assert.Error(t, err)
	})

	t.Run("non-multipart content type is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello"))
		req.Header.Set("Content-Type", "application/json")

		_, err := ReadSingleMultipartFile(req)
		assert.Error(t, err)
	})
}

func TestNormalizeFilename(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"plain", "receipt.pdf", "receipt.pdf", false},
		{"posix path", "a/b/c.pdf", "c.pdf", false},
		{"windows path", `a\b\c.pdf`, "c.pdf", false},
		{"control characters stripped", "a\x00b\x1f.pdf", "ab.pdf", false},
		{"blank", "   ", "", true},
		{"dot", ".", "", true},
		{"dotdot", "..", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeFilename(tt.raw)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
