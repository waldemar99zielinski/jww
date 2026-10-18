package kernel

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"jww/internal/service/object_storage"
)

var ErrInvalidRange = errors.New("invalid or unsatisfiable byte range")

// WriteDownload streams body to the client as a forced download. It is the one
// documented exception to the JSON success envelope: bytes on success, and JSON
// errors only before streaming begins. Callers that can still fail before
// writing should do so first.
func WriteDownload(w http.ResponseWriter, filename, mediaType string, byteSize int64, body io.ReadCloser) {
	writeAttachmentContent(w, filename, mediaType, byteSize, false, nil, body)
}

// WriteGrantedContent streams a full or partial granted attachment. Preview
// controls inline versus attachment disposition.
func WriteGrantedContent(w http.ResponseWriter, filename, mediaType string, totalSize int64, preview bool, byteRange *object_storage.ByteRange, body io.ReadCloser) {
	writeAttachmentContent(w, filename, mediaType, totalSize, preview, byteRange, body)
}

func writeAttachmentContent(w http.ResponseWriter, filename, mediaType string, totalSize int64, preview bool, byteRange *object_storage.ByteRange, body io.ReadCloser) {
	defer body.Close()
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mediaType)
	contentSize := totalSize
	status := http.StatusOK
	if byteRange != nil {
		contentSize = byteRange.Length()
		status = http.StatusPartialContent
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", byteRange.Start, byteRange.End, totalSize))
	}
	if contentSize >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(contentSize, 10))
	}
	disposition := "attachment"
	if preview {
		disposition = "inline"
		w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	}
	w.Header().Set("Content-Disposition", contentDisposition(disposition, filename))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(status)
	_, _ = io.Copy(w, body)
}

// ParseSingleByteRange parses one RFC 7233 bytes range and normalizes it to an
// inclusive range bounded by size. Multiple ranges are deliberately rejected.
func ParseSingleByteRange(value string, size int64) (*object_storage.ByteRange, error) {
	if value == "" {
		return nil, nil
	}
	if size <= 0 || !strings.HasPrefix(value, "bytes=") || strings.Contains(value, ",") {
		return nil, ErrInvalidRange
	}
	spec := strings.TrimPrefix(value, "bytes=")
	startText, endText, ok := strings.Cut(spec, "-")
	if !ok || (startText == "" && endText == "") {
		return nil, ErrInvalidRange
	}

	if startText == "" {
		suffix, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffix <= 0 {
			return nil, ErrInvalidRange
		}
		if suffix > size {
			suffix = size
		}
		return &object_storage.ByteRange{Start: size - suffix, End: size - 1}, nil
	}

	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= size {
		return nil, ErrInvalidRange
	}
	end := size - 1
	if endText != "" {
		end, err = strconv.ParseInt(endText, 10, 64)
		if err != nil || end < start {
			return nil, ErrInvalidRange
		}
		if end >= size {
			end = size - 1
		}
	}
	return &object_storage.ByteRange{Start: start, End: end}, nil
}

func WriteRangeNotSatisfiable(w http.ResponseWriter, size int64) {
	w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
}

// attachmentDisposition builds an RFC 6266 attachment header: a conservative
// ASCII fallback plus a UTF-8 (RFC 5987) form for non-ASCII filenames.
func attachmentDisposition(filename string) string {
	return contentDisposition("attachment", filename)
}

func contentDisposition(disposition, filename string) string {
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
	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`, disposition, ascii, escaped)
}
