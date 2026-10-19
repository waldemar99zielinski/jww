package imageprep_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/gen2brain/h265/heic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/imageprep"
)

func TestNormalize_ConvertsHEICToBoundedPNG(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 2600, 2000))
	for y := 0; y < 2000; y++ {
		for x := 0; x < 2600; x++ {
			source.SetNRGBA(x, y, color.NRGBA{R: uint8(x % 255), G: uint8(y % 255), B: 120, A: 255})
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, heic.Encode(&encoded, source))

	content, err := imageprep.New().Normalize("IMG_1234.HEIC", "image/heic", bytes.NewReader(encoded.Bytes()))
	require.NoError(t, err)

	assert.Equal(t, "IMG_1234.png", content.Filename)
	assert.Equal(t, "image/png", content.MediaType)
	decoded, err := png.Decode(content.Body)
	require.NoError(t, err)
	assert.Equal(t, 2000, decoded.Bounds().Dx())
	assert.LessOrEqual(t, decoded.Bounds().Dy(), 2000)
}

func TestNormalize_LeavesOtherUploadsStreaming(t *testing.T) {
	body := strings.NewReader("original bytes")

	content, err := imageprep.New().Normalize("receipt.jpg", "image/jpeg", body)

	require.NoError(t, err)
	assert.Equal(t, "receipt.jpg", content.Filename)
	assert.Equal(t, "image/jpeg", content.MediaType)
	got, err := io.ReadAll(content.Body)
	require.NoError(t, err)
	assert.Equal(t, "original bytes", string(got))
}

func TestNormalize_RejectsCorruptHEIC(t *testing.T) {
	_, err := imageprep.New().Normalize("broken.heic", "image/heic", strings.NewReader("not an image"))

	require.Error(t, err)
}
