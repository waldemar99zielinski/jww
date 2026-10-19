// Package imageprep normalizes uploaded images before they are stored. Keeping
// format conversion here means clients upload original bytes and every stored
// attachment is decodable by the configured vision provider.
package imageprep

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime"
	"path"
	"strings"

	"github.com/gen2brain/h265/heic"
	xdraw "golang.org/x/image/draw"

	"jww/internal/service/transaction_attachment"
)

// maxDimension bounds the long edge after conversion. Phone photos are far
// larger than a receipt needs, and an oversized PNG can time out when the
// vision provider fetches it through a public tunnel.
const maxDimension = 2000

// maxFramePixels refuses pathological HEIC inputs before decoding their full
// raster. A 48 MP phone photo still fits comfortably.
const maxFramePixels = 48_000_000

// Normalizer converts HEIC/HEIF uploads to PNG and leaves every other media
// type streaming untouched.
type Normalizer struct{}

func New() *Normalizer { return &Normalizer{} }

func (n *Normalizer) Normalize(filename, mediaType string, body io.Reader) (transaction_attachment.NormalizedContent, error) {
	if !isHEIC(filename, mediaType) {
		return transaction_attachment.NormalizedContent{Filename: filename, MediaType: mediaType, Body: body}, nil
	}

	decoded, err := heic.Decode(body, heic.Options{
		AutoRotate:     true,
		FrameSizeLimit: maxFramePixels,
	})
	if err != nil {
		return transaction_attachment.NormalizedContent{}, fmt.Errorf("decoding HEIC image: %w", err)
	}

	normalized := fitWithin(decoded, maxDimension)
	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, normalized); err != nil {
		return transaction_attachment.NormalizedContent{}, fmt.Errorf("encoding PNG image: %w", err)
	}
	return transaction_attachment.NormalizedContent{
		Filename:  pngFilename(filename),
		MediaType: "image/png",
		Body:      bytes.NewReader(pngBytes.Bytes()),
	}, nil
}

func isHEIC(filename, mediaType string) bool {
	parsed, _, err := mime.ParseMediaType(mediaType)
	if err == nil {
		parsed = strings.ToLower(parsed)
		if parsed == "image/heic" || parsed == "image/heif" {
			return true
		}
	}
	extension := strings.ToLower(path.Ext(filename))
	return extension == ".heic" || extension == ".heif"
}

func pngFilename(filename string) string {
	extension := path.Ext(filename)
	if extension == "" {
		return filename + ".png"
	}
	return strings.TrimSuffix(filename, extension) + ".png"
}

// fitWithin scales the long edge down to max without upscaling. draw.CatmullRom
// preserves receipt text far better than nearest-neighbour sampling.
func fitWithin(source image.Image, maxEdge int) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	longEdge := max(width, height)
	if longEdge <= maxEdge {
		return source
	}

	scale := float64(maxEdge) / float64(longEdge)
	targetWidth := maxInt(1, int(float64(width)*scale))
	targetHeight := maxInt(1, int(float64(height)*scale))
	target := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.CatmullRom.Scale(target, target.Bounds(), source, bounds, xdraw.Src, nil)
	return target
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
