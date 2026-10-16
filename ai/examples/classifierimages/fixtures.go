package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/tokenbeat-lab/barness/ai"
)

// Both formats contain an actual clockwise rotation and crop. Provenance is
// Host data; the classifier sees only the resulting 320x480 file.
type geometry struct {
	OriginalWidth    int             `json:"originalWidth"`
	OriginalHeight   int             `json:"originalHeight"`
	ClockwiseDegrees int             `json:"clockwiseDegrees"`
	Crop             image.Rectangle `json:"crop"`
	Width            int             `json:"width"`
	Height           int             `json:"height"`
}

type fixture struct {
	Name     string                  `json:"name"`
	Identity ai.ClassifierImageInput `json:"identity"`
	Geometry geometry                `json:"geometry"`
	Image    ai.Image                `json:"-"`
	Bytes    []byte                  `json:"-"`
}

func fixtures() ([]fixture, []byte, error) {
	original := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(original, original.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(original, image.Rect(200, 200, 440, 360), image.NewUniform(color.RGBA{R: 220, A: 255}), image.Point{}, draw.Src)
	rotated := image.NewRGBA(image.Rect(0, 0, 480, 640))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			rotated.Set(479-y, x, original.At(x, y))
		}
	}
	crop := image.Rect(40, 80, 360, 560)
	processed := image.NewRGBA(image.Rect(0, 0, 320, 480))
	draw.Draw(processed, processed.Bounds(), rotated, crop.Min, draw.Src)
	var source, p, j bytes.Buffer
	if err := png.Encode(&source, original); err != nil {
		return nil, nil, err
	}
	if err := png.Encode(&p, processed); err != nil {
		return nil, nil, err
	}
	if err := jpeg.Encode(&j, processed, &jpeg.Options{Quality: 90}); err != nil {
		return nil, nil, err
	}
	g := geometry{640, 480, 90, crop, 320, 480}
	var result []fixture
	for _, file := range []struct {
		name, mime string
		data       []byte
	}{{"processed.png", "image/png", p.Bytes()}, {"processed.jpg", "image/jpeg", j.Bytes()}} {
		sum := sha256.Sum256(file.data)
		result = append(result, fixture{file.name, ai.ClassifierImageInput{MimeType: file.mime, SHA256: hex.EncodeToString(sum[:])}, g, ai.Image{MimeType: file.mime, Data: base64.StdEncoding.EncodeToString(file.data)}, file.data})
	}
	return result, source.Bytes(), nil
}
