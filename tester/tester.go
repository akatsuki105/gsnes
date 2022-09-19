package tester

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"net/http"

	"github.com/pokemium/gsnes/core"
	"github.com/pokemium/iro"
)

const (
	OK = "✅"
	NG = "❌"
)

func SFC(romData []byte) core.SuperFamicom {
	sfc := core.New()
	sfc.LoadROM(romData)
	return sfc
}

func CompareImage(actual []iro.RGB555, expected image.Image, w, h int, fn func(actual, expected color.Color) bool) error {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !fn(actual[y*w+x], expected.At(x, y)) {
				return fmt.Errorf("two images are not the same in (%d, %d)", x, y)
			}
		}
	}
	return nil
}

// 画像の黒部分(#000000)の座標が全て一致するかチェック
func CompareBlack(actual, expected color.Color) bool {
	r1, g1, b1, _ := actual.RGBA()
	black1 := r1 == 0 && g1 == 0 && b1 == 0

	r2, g2, b2, _ := expected.RGBA()
	black2 := r2 == 0 && g2 == 0 && b2 == 0

	if (black1 || black2) && !(black1 && black2) {
		return false
	}

	return true
}

// 画像の白部分(#E0E0E0より明るい)の座標が全て一致するかチェック
func CompareWhite(actual, expected color.Color) bool {
	r1, g1, b1, _ := actual.RGBA()
	white1 := r1 > 0xE0 && g1 > 0xE0 && b1 > 0xE0

	r2, g2, b2, _ := expected.RGBA()
	white2 := r2 > 0xE0 && g2 > 0xE0 && b2 > 0xE0

	if (white1 || white2) && !(white1 && white2) {
		return false
	}

	return true
}

// URLからバイナリ取得
func fetchResource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// URLから画像取得
func fetchTestImage(p string) (image.Image, error) {
	data, err := fetchResource(p)
	if err != nil {
		return nil, err
	}

	result, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// nフレームだけエミュレーションした後の画面がimgpathの画像と比較して一致するかチェックする
func IsExpectedScreen(rom, imgpath string, n int, compare func(actual color.Color, expected color.Color) bool) error {
	romData, err := fetchResource(rom)
	if err != nil {
		return err
	}

	expected, err := fetchTestImage(imgpath)
	if err != nil {
		return err
	}

	c := SFC(romData)

	for i := 0; i < n; i++ {
		c.RunFrame()
	}
	actual := c.FrameBuffer()

	w, h := c.Resolution()
	return CompareImage(actual, expected, w, h-1, compare)
}
