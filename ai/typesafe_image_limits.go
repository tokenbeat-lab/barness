package ai

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"image"
	"io"
	"strings"
)

// Official JSON image subset, checked independently of Decisions and Host
// budgets. Source: https://docs.typesafe.ai/images (2026-10-16). These bounds
// cannot predict the service's postprocessing size or image token count.
const (
	typeSafeImageBytes  int64 = 16 << 20
	typeSafeBodyBytes   int64 = 32 << 20
	typeSafeKeyBytes          = 1024
	typeSafeAllKeyBytes       = 65536
)

func checkTypeSafeImageFiles(ctx context.Context, images []Image, phase Phase) *Error {
	for _, img := range images {
		if failure := contextError(ctx, phase); failure != nil {
			return failure
		}
		if base64Size(img.Data) > typeSafeImageBytes {
			return limitFailure(phase, "Jev.MaxImageBytes", typeSafeImageBytes)
		}
		if img.MimeType != "image/png" && img.MimeType != "image/jpeg" {
			return imageInputFailure(phase, "Jev images require static PNG or JPEG")
		}
		reader := func() io.Reader {
			return typeSafeImageReader{ctx: ctx, reader: base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(img.Data))}
		}
		cfg, format, err := image.DecodeConfig(reader())
		if failure := contextError(ctx, phase); failure != nil {
			return failure
		}
		if err != nil || (format == "png") != (img.MimeType == "image/png") {
			return imageInputFailure(phase, "invalid Jev image format or dimensions")
		}
		// Bound each edge before multiplication; no raster is decompressed.
		if cfg.Width > 1568 || cfg.Height > 1568 {
			return limitFailure(phase, "Jev.MaxImageEdge", 1568)
		}
		if cfg.Width < 256 || cfg.Height < 256 || int64(cfg.Width)*int64(cfg.Height) < 78400 {
			return imageInputFailure(phase, "Jev image dimensions are below the supported minimum")
		}
		if int64(cfg.Width)*int64(cfg.Height) > 2458624 {
			return limitFailure(phase, "Jev.MaxImagePixels", 2458624)
		}
		if format == "png" && !typeSafeStaticPNG(reader()) {
			if failure := contextError(ctx, phase); failure != nil {
				return failure
			}
			return imageInputFailure(phase, "Jev images require a complete static PNG")
		}
	}
	return contextError(ctx, phase)
}

type typeSafeImageReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r typeSafeImageReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

// Scan chunk framing rather than searching the compressed bytes for an APNG
// marker. A marker inside IDAT is ordinary image data. Memory stays constant.
func typeSafeStaticPNG(r io.Reader) bool {
	var signature [8]byte
	if _, err := io.ReadFull(r, signature[:]); err != nil || string(signature[:]) != "\x89PNG\r\n\x1a\n" {
		return false
	}
	var buffer [4096]byte
	for {
		var header [8]byte
		if _, err := io.ReadFull(r, header[:]); err != nil {
			return false
		}
		n := int64(binary.BigEndian.Uint32(header[:4]))
		kind := string(header[4:])
		if kind == "acTL" || kind == "fcTL" || kind == "fdAT" {
			return false
		}
		if n > typeSafeImageBytes {
			return false
		}
		for remaining := n + 4; remaining > 0; {
			take := min(remaining, int64(len(buffer)))
			if _, err := io.ReadFull(r, buffer[:take]); err != nil {
				return false
			}
			remaining -= take
		}
		if kind == "IEND" {
			if n != 0 {
				return false
			}
			n, err := r.Read(buffer[:1])
			return n == 0 && err == io.EOF
		}
	}
}
