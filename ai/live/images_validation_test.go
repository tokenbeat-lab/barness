//go:build live

package live

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"

	"github.com/tokenbeat-lab/barness/ai"
)

// ImageVerification is independently reproducible from the saved raster. The
// header is bounded before full decoding; no output pixel equality is asserted.
type imageVerification struct {
	SHA256 string `json:"sha256,omitempty"`
	Bytes  int    `json:"bytes"`
	Format string `json:"format,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func verifyImages(res ai.ImagesResult, format string) (imageVerification, error) {
	out := imageVerification{}
	if len(res.Content) != 1 {
		return out, errors.New("expected exactly one output image")
	}
	block, ok := res.Content[0].(ai.ImageOutputImage)
	if !ok || block.MimeType != "image/"+format {
		return out, errors.New("output MIME disagrees with requested format")
	}
	if base64.StdEncoding.DecodedLen(len(block.Data)) > 8<<20 {
		return out, errors.New("output image exceeds validation byte budget")
	}
	data, err := base64.StdEncoding.Strict().DecodeString(block.Data)
	if err != nil {
		return out, errors.New("invalid output base64")
	}
	config, actual, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || actual != format {
		return out, errors.New("unreadable output header or format mismatch")
	}
	out.Bytes, out.Format, out.Width, out.Height = len(data), actual, config.Width, config.Height
	if config.Width != 1024 || config.Height != 1024 {
		return out, errors.New("output size must be 1024x1024 within the 1K smoke budget")
	}
	raster, actual, err := image.Decode(bytes.NewReader(data))
	if err != nil || actual != format || raster.Bounds().Dx() != 1024 || raster.Bounds().Dy() != 1024 {
		return out, errors.New("output raster is not fully readable")
	}
	sum := sha256.Sum256(data)
	out.SHA256 = hex.EncodeToString(sum[:])
	return out, nil
}
