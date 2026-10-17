package kernel

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// WriteDownload streams body to the client as a forced download. It is the one
// documented exception to the JSON success envelope: bytes on success, and JSON
// errors only before streaming begins. Callers that can still fail before
// writing should do so first.
func WriteDownload(w http.ResponseWriter, filename, mediaType string, byteSize int64, body io.ReadCloser) {
	defer body.Close()
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mediaType)
	if byteSize >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(byteSize, 10))
	}
	w.Header().Set("Content-Disposition", attachmentDisposition(filename))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, body)
}

// attachmentDisposition builds an RFC 6266 attachment header: a conservative
// ASCII fallback plus a UTF-8 (RFC 5987) form for non-ASCII filenames.
func attachmentDisposition(filename string) string {
	ascii := strings.Map(func(r rune) rune {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, filename)
	if ascii == "" {
		ascii = "download"
	}
	escaped := url.PathEscape(filename)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, ascii, escaped)
}
