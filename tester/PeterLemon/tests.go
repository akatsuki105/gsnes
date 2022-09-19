package main

import (
	"fmt"

	"github.com/pokemium/gsnes/tester"
)

func testBank(name string) func() error {
	return func() error {
		endpoint := fmt.Sprintf("/BANK/%s/BANK%s", name, name)
		path := BASE_URL + endpoint
		rom := path + ".sfc"
		img := path + ".png"
		return tester.IsExpectedScreen(rom, img, SECONDS*60, tester.CompareWhite)
	}
}

func testCpu(name string) func() error {
	return func() error {
		endpoint := fmt.Sprintf("/CPUTest/CPU/%s/CPU%s", name, name)
		path := BASE_URL + endpoint
		rom := path + ".sfc"
		img := path + ".png"
		return tester.IsExpectedScreen(rom, img, SECONDS*60, tester.CompareWhite)
	}
}

func testBGMap(bpp, name string) func() error {
	return func() error {
		endpoint := fmt.Sprintf("/PPU/BGMAP/8x8/%s/%s/%s", bpp, name, name)
		path := BASE_URL + endpoint
		rom := path + ".sfc"
		img := path + ".png"
		return tester.IsExpectedScreen(rom, img, SECONDS*60, tester.CompareWhite)
	}
}

func testHelloWorld() error {
	path := BASE_URL + "/HelloWorld/HelloWorld"
	rom := path + ".sfc"
	img := path + ".png"
	return tester.IsExpectedScreen(rom, img, SECONDS*60, tester.CompareBlack)
}
