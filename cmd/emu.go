package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/akatsuki105/gsnes/core"
	"github.com/akatsuki105/iro"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// pauseボタンがリリースされた場合にtrueに、押された場合にfalseにする
var pauseEnable = true

// pauseEnable のstep版
var stepEnable = true

var keyMap = map[ebiten.Key]string{
	ebiten.KeyX:          "A",
	ebiten.KeyZ:          "B",
	ebiten.KeyS:          "X",
	ebiten.KeyA:          "Y",
	ebiten.KeyQ:          "L",
	ebiten.KeyW:          "R",
	ebiten.KeyBackspace:  "SELECT",
	ebiten.KeyEnter:      "START",
	ebiten.KeyArrowUp:    "UP",
	ebiten.KeyArrowRight: "RIGHT",
	ebiten.KeyArrowDown:  "DOWN",
	ebiten.KeyArrowLeft:  "LEFT",
}

var btnMap = map[ebiten.GamepadButton]string{
	ebiten.GamepadButton2:  "A",
	ebiten.GamepadButton1:  "B",
	ebiten.GamepadButton3:  "B",
	ebiten.GamepadButton4:  "L",
	ebiten.GamepadButton5:  "R",
	ebiten.GamepadButton8:  "SELECT",
	ebiten.GamepadButton9:  "START",
	ebiten.GamepadButton15: "UP",
	ebiten.GamepadButton16: "RIGHT",
	ebiten.GamepadButton17: "DOWN",
	ebiten.GamepadButton18: "LEFT",
}

type emulator struct {
	sfc         core.SuperFamicom
	frameBuffer []iro.RGB555
	frame       uint64
	debug       bool
	win         window
	texts       []*text

	// all queue tasks are executed on each Update()
	queue queue
}

type window struct {
	title           string
	backgroundColor color.Color
}

type text struct {
	id      string
	content string
	x, y    int
}

func (t *text) Pos(x, y int) {
	t.x, t.y = x, y
}

func new() *emulator {
	sfc := core.New()
	w, h := sfc.Resolution()
	return &emulator{
		sfc:         sfc,
		frameBuffer: make([]iro.RGB555, w*h),
		texts:       make([]*text, 0),
		queue:       make([]*command, 0),
		win:         window{"gsnes", color.RGBA{35, 27, 167, 255}},
	}
}

func (e *emulator) setDebugMode(mode bool) {
	e.queue = append(e.queue, newCommand(func() {
		e.debug = mode
		if e.debug {
			ebiten.SetWindowSize(720, 640)

			// mmap RAM
			e.sfc.MMap("WRAM", e.MMap(MMAP_WRAM, int(core.WRAM_SIZE)))
			e.sfc.MMap("VRAM", e.MMap(MMAP_VRAM, int(core.VRAM_SIZE)))
			e.sfc.MMap("PALETTE", e.MMap(MMAP_PALETTE, int(core.PAL_SIZE)))
		} else {
			w, h := e.sfc.Resolution()
			ebiten.SetWindowSize(w*2, h*2)
		}
	}))
}

func (e *emulator) Update() error {
	ebiten.SetWindowTitle(e.win.title)
	e.queue.exec()
	e.pollInput()

	if !e.sfc.Paused() {
		e.sfc.RunFrame()
	}

	if e.debug {
		e.debugPrint("FPS", "FPS: "+fmt.Sprint(ebiten.CurrentTPS())).Pos(4, 230)
		e.debugPrint("Status/CPU", e.sfc.Status("CPU")).Pos(264, 4)
		e.debugPrint("Status/PPU", e.sfc.Status("PPU")).Pos(264, 68)
		e.debugPrint("Status/Sys", e.sfc.Status("SYSTEM")).Pos(264, 116)

		sp, stack := e.sfc.Stack(8)
		s := fmt.Sprintf("%04X ->", sp)
		for i := range stack {
			if i != 0 {
				s += fmt.Sprintf("       %02X\n", stack[i])
			} else {
				s += fmt.Sprintf("%02X\n", stack[i])
			}
		}
		e.debugPrint("Stack", s).Pos(512, 4)

		e.debugPrint("Status/Events", e.sfc.Status("EVENTS")).Pos(4, 250)
		e.debugPrint("Status/SCREEN", e.sfc.Status("SCREEN")).Pos(4, 280)
		e.debugPrint("Status/OAM", e.sfc.Status("OAM")).Pos(264, 250)
	}
	return nil
}

func (e *emulator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	w, h := outsideWidth, outsideHeight
	if !e.debug {
		w, h = e.sfc.Resolution()
	}
	return w, h
}

func (e *emulator) Draw(screen *ebiten.Image) {
	screen.Fill(e.win.backgroundColor)
	for i := range e.texts {
		ebitenutil.DebugPrintAt(screen, e.texts[i].content, e.texts[i].x, e.texts[i].y)
	}

	img := e.draw()
	// writeGrid(img, 8)
	screen.DrawImage(ebiten.NewImageFromImage(img), nil)
	e.frame++
}

func (e *emulator) draw() *image.RGBA {
	w, h := e.sfc.Resolution()
	if e.sfc.Paused() {
		return iro.RGB555ToImage(e.frameBuffer, w, h, nil)
	}

	screen := e.sfc.FrameBuffer()
	copy(e.frameBuffer, screen)
	return iro.RGB555ToImage(e.frameBuffer, w, h, nil)
}

func (e *emulator) pollInput() {
	for key, input := range keyMap {
		if !e.sfc.Paused() {
			pressed := ebiten.IsKeyPressed(key)
			e.sfc.SetKeyInput(input, pressed)
		}
	}

	// Pause emu core
	if ebiten.IsKeyPressed(ebiten.KeyMeta) && ebiten.IsKeyPressed(ebiten.KeyP) {
		if pauseEnable {
			pauseEnable = false
			e.sfc.Pause(!e.sfc.Paused())
		}
	} else {
		pauseEnable = true
	}

	// Step core
	if ebiten.IsKeyPressed(ebiten.KeyMeta) && ebiten.IsKeyPressed(ebiten.KeyS) {
		if stepEnable {
			stepEnable = false
			e.sfc.RunInst()
		}
	} else {
		stepEnable = true
	}
}

func (e *emulator) debugPrint(id, content string) *text {
	for i := range e.texts {
		if id == e.texts[i].id {
			e.texts[i].content = content
			return e.texts[i]
		}
	}

	e.texts = append(e.texts, &text{id, content, 0, 0})
	return e.texts[len(e.texts)-1]
}
