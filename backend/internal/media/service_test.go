package media

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"testing"
)

func TestInspectImageAcceptsPNG(t *testing.T) {
	var encoded bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 32, 18))
	picture.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&encoded, picture); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}

	mimeType, width, height, animated, err := inspectImage(encoded.Bytes())
	if err != nil {
		t.Fatalf("inspectImage() error = %v", err)
	}
	if mimeType != "image/png" || width != 32 || height != 18 || animated {
		t.Fatalf("inspectImage() = %q, %d, %d, %v", mimeType, width, height, animated)
	}
}

func TestInspectImageMarksGIFAnimated(t *testing.T) {
	var encoded bytes.Buffer
	picture := &image.Paletted{
		Pix:     []uint8{0},
		Stride:  1,
		Rect:    image.Rect(0, 0, 1, 1),
		Palette: color.Palette{color.Black},
	}
	if err := gif.Encode(&encoded, picture, nil); err != nil {
		t.Fatalf("gif.Encode() error = %v", err)
	}

	mimeType, _, _, animated, err := inspectImage(encoded.Bytes())
	if err != nil {
		t.Fatalf("inspectImage() error = %v", err)
	}
	if mimeType != "image/gif" || !animated {
		t.Fatalf("inspectImage() mime = %q, animated = %v", mimeType, animated)
	}
}

func TestInspectImageRejectsNonImageAndOversizedPayload(t *testing.T) {
	if _, _, _, _, err := inspectImage([]byte("not an image")); err == nil {
		t.Fatal("inspectImage() accepted non-image content")
	}
	if _, _, _, _, err := inspectImage(make([]byte, MaxUploadBytes+1)); err == nil {
		t.Fatal("inspectImage() accepted an oversized payload")
	}
}
