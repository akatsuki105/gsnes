package main

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
)

var (
	gridColor = color.RGBA{0x4f, 0x4f, 0x4f, 0xff}
)

type queue []*command

type command struct {
	Callback func()
}

func newCommand(callback func()) *command {
	return &command{
		Callback: callback,
	}
}

func (q *queue) exec() {
	for _, cmd := range *q {
		if cmd.Callback != nil {
			cmd.Callback()
		}
	}

	*q = make([]*command, 0)
}

// WritePNG writes image to png format
func writePNG(dstPath string, i *image.RGBA) error {
	if i == nil {
		return errors.New("image is nil")
	}

	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = png.Encode(file, i); err != nil {
		return err
	}

	return nil
}

func writeGrid(src *image.RGBA, unit int) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()

	// draw horizontally
	for row := 0; row < (h / unit); row++ {
		for x := 0; x < w; x++ {
			src.SetRGBA(x, row*unit, gridColor)
		}
	}
	for x := 0; x < w; x++ {
		src.SetRGBA(x, h-1, gridColor)
	}

	// draw vertically
	for col := 0; col < (w / unit); col++ {
		for y := 0; y < h; y++ {
			src.SetRGBA(col*unit, y, gridColor)
		}
	}
	for y := 0; y < h; y++ {
		src.SetRGBA(w-1, y, gridColor)
	}
}
