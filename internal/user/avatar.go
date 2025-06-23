package user

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/draw"
	"image/png"
	"sync"

	"github.com/o1egl/govatar"
)

var (
	cache map[string]string = map[string]string{}
	mu    sync.Mutex
)

func avatar(seed string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if img, ok := cache[seed]; ok {
		return img, nil
	}

	size := 400
	a, err := govatar.GenerateForUsername(govatar.FEMALE, seed)
	if err != nil {
		return "", nil
	}
	pos := image.Rect(0, 0, 400, 400)
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, pos, a, image.Point{}, draw.Over)

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(encoder, img); err != nil {
		return "", err
	}
	// Cache before returning
	cache[seed] = buf.String()
	return buf.String(), nil
}
