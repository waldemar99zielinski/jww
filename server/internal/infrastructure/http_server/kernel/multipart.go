package kernel

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
)

// Multipart file constraints, mirroring domain.TransactionAttachment's bounds.
const (
	defaultAttachmentMediaType = "application/octet-stream"
	maxAttachmentFilenameRunes = 255
	maxAttachmentMediaTypeLen  = 255
)

// MultipartFile is the single file part of an attachment upload. It is an
// io.Reader over the part's bytes; read it through to the end and then call
// Finish to prove no unexpected extra parts followed before committing.
type MultipartFile struct {
	Filename  string
	MediaType string
	part      *multipart.Part
	reader    *multipart.Reader
	finished  bool
}

// ReadSingleMultipartFile parses a streaming multipart/form-data upload:
// exactly one file part named "file" with a usable filename, and no other
// parts. It uses r.MultipartReader and iterates parts, never
// ParseMultipartForm/ReadForm — those whole-form APIs may retain the body in
// memory or spool it to local temp files.
func ReadSingleMultipartFile(r *http.Request) (*MultipartFile, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, fmt.Errorf("request must be multipart/form-data with a valid boundary")
	}
	part, err := reader.NextPart()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New(`multipart body must contain exactly one file part named "file"`)
		}
		return nil, fmt.Errorf("reading multipart body: %w", err)
	}
	if part.FormName() != "file" || part.FileName() == "" {
		return nil, errors.New(`multipart body must contain exactly one file part named "file"`)
	}
	filename, err := normalizeFilename(part.FileName())
	if err != nil {
		return nil, err
	}
	mediaType := strings.TrimSpace(part.Header.Get("Content-Type"))
	if mediaType == "" {
		mediaType = defaultAttachmentMediaType
	}
	if len(mediaType) > maxAttachmentMediaTypeLen {
		return nil, errors.New("multipart file part media type is too long")
	}
	return &MultipartFile{Filename: filename, MediaType: mediaType, part: part, reader: reader}, nil
}

// Read streams the part's bytes. It is passed straight to the upload workflow,
// so the file is never buffered whole.
func (f *MultipartFile) Read(p []byte) (int, error) { return f.part.Read(p) }

// Finish drains the current part and asserts there is no second part. Its error
// is a 400: the request carried more than one part.
func (f *MultipartFile) Finish() error {
	if f.finished {
		return nil
	}
	f.finished = true
	if _, err := io.Copy(io.Discard, f.part); err != nil {
		return fmt.Errorf("reading multipart file part: %w", err)
	}
	if _, err := f.reader.NextPart(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New(`multipart body must contain exactly one file part named "file"`)
		}
		return fmt.Errorf("reading multipart body: %w", err)
	}
	return nil
}

// normalizeFilename strips path components and control characters and caps the
// result, so a client cannot smuggle a path or a header-injection payload
// through a display filename. The storage key is generated independently.
func normalizeFilename(raw string) (string, error) {
	name := path.Base(strings.ReplaceAll(raw, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || name == "/" {
		return "", errors.New("multipart file part must have a filename")
	}
	if runes := []rune(name); len(runes) > maxAttachmentFilenameRunes {
		name = string(runes[:maxAttachmentFilenameRunes])
	}
	return name, nil
}
