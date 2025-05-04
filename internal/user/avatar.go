package user

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/draw"
	"image/png"

	"github.com/lafriks/go-avatars"
)

func avatar(seed string) string {
	size := 80
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	a, err := avatars.Generate(seed)
	if err != nil {
		panic(err)
	}
	av, err := a.Image(avatars.RenderSize(size))
	if err != nil {
		panic(err)
	}
	pos := image.Rect(0, 0, 80, 80)
	draw.Draw(img, pos, av, image.Point{}, draw.Over)

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(encoder, img); err != nil {
		panic(err)
	}
	return buf.String()
}
