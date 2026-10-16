//go:build ignore

// Run from repository root. Scenes contain drawing instructions, never labels.
package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	dir := "ai/examples/visualeval/testdata"
	raw, err := os.ReadFile(filepath.Join(dir, "scenes.json"))
	must(err)
	var scenes []struct {
		ID, Shape, Format   string
		X, Y, Width, Height int
	}
	must(json.Unmarshal(raw, &scenes))
	for _, s := range scenes {
		im := image.NewRGBA(image.Rect(0, 0, 320, 320))
		draw.Draw(im, im.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		switch s.Shape {
		case "red":
			draw.Draw(im, image.Rect(s.X, s.Y, s.X+s.Width, s.Y+s.Height), image.NewUniform(color.RGBA{220, 20, 20, 255}), image.Point{}, draw.Src)
		case "blue":
			for y := 0; y < 320; y++ {
				for x := 0; x < 320; x++ {
					dx, dy := x-160, y-160
					if dx*dx+dy*dy <= 6400 {
						im.Set(x, y, color.RGBA{20, 20, 220, 255})
					}
				}
			}
		}
		var source, processed bytes.Buffer
		must(png.Encode(&source, im))
		must(os.WriteFile(filepath.Join(dir, s.ID+"-source.png"), source.Bytes(), 0600))
		if s.Format == "jpg" {
			must(jpeg.Encode(&processed, im, &jpeg.Options{Quality: 90}))
		} else {
			must(png.Encode(&processed, im))
		}
		must(os.WriteFile(filepath.Join(dir, s.ID+"."+s.Format), processed.Bytes(), 0600))
	}
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
