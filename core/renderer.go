package core

import (
	"github.com/pokemium/iro"
)

type renderer struct {
	vram    *vram
	pal     *palette
	oam     *oam
	screens [2][HORIZONTAL * VERTICAL]iro.RGB555 // 0: main, 1: sub

	mode                   uint8   // BG Mode(0..7)
	layers                 []layer // idx 0 is backdrop
	bg1, bg2, bg3, bg4     *bg
	bg1h, bg2h, bg3h, bg4h *bg
	objs                   [4]*objl
	backdrop               layer
	bg3a                   bool // BGMODE.3
	w                      windowSystem
}

func newRenderer(vram *vram, pal *palette, oam *oam) *renderer {
	r := &renderer{
		vram: vram,
		pal:  pal,
		oam:  oam,
	}
	r.bg1, r.bg2, r.bg3, r.bg4 = newBg(r, 1), newBg(r, 2), newBg(r, 3), newBg(r, 4)
	r.bg1h, r.bg2h, r.bg3h, r.bg4h = newBgH(r.bg1), newBgH(r.bg2), newBgH(r.bg3), newBgH(r.bg4)
	r.backdrop = newBackdrop(&r.pal.buf[0])

	obj0, obj1, obj2, obj3 := newObjLayer(r, 0), newObjLayer(r, 1), newObjLayer(r, 2), newObjLayer(r, 3)
	r.objs = [4]*objl{obj0, obj1, obj2, obj3}
	return r
}

func (r *renderer) reset() {
	r.setBgMode(0)

	for i := range r.screens[0] {
		r.screens[0][i] = iro.RGB555(0x0000)
		r.screens[1][i] = iro.RGB555(0x0000)
	}
}

func (r *renderer) drawScanline(y uint16) {
	for _, l := range r.layers {
		if l.enable() {
			if l.isMain() {
				l.drawScanline(r.screens[0][HORIZONTAL*(y-1):], y, 0, 256)
			}
			if l.isSub() {
				l.drawScanline(r.screens[1][HORIZONTAL*(y-1):], y, 0, 256)
			}
		}
	}
}

func (r *renderer) frameBuffer() []iro.RGB555 {
	return r.screens[0][:]
}

func (r *renderer) setBgMode(m uint8) {
	r.mode = m
	bg1, bg2, bg3, bg4 := r.bg1, r.bg2, r.bg3, r.bg4
	bg1.color = DISABLED
	bg2.color = DISABLED
	bg3.color = DISABLED
	bg4.color = DISABLED

	switch m {
	case 0:
		bg1.color, bg1.palOfs = COLOR_2BPP, 0x00
		bg2.color, bg2.palOfs = COLOR_2BPP, 0x20
		bg3.color, bg3.palOfs = COLOR_2BPP, 0x40
		bg4.color, bg4.palOfs = COLOR_2BPP, 0x60
		r.layers = []layer{r.backdrop, r.bg4, r.bg3, r.objs[0], r.bg4h, r.bg3h, r.objs[1], r.bg2, r.bg1, r.objs[2], r.bg2h, r.bg1h, r.objs[3]}

	case 1:
		bg1.color, bg1.palOfs = COLOR_4BPP, 0x00
		bg2.color, bg2.palOfs = COLOR_4BPP, 0x00
		bg3.color, bg3.palOfs = COLOR_2BPP, 0x00
		r.layers = []layer{r.backdrop, r.bg3, r.objs[0], r.bg3h, r.objs[1], r.bg2, r.bg1, r.objs[2], r.bg2h, r.bg1h, r.objs[3]}
		if r.bg3a {
			r.layers = []layer{r.backdrop, r.bg3, r.objs[0], r.objs[1], r.bg2, r.bg1, r.objs[2], r.bg2h, r.bg1h, r.objs[3], r.bg3h}
		}

	case 2, 3, 4, 5:
		switch m {
		case 2:
			bg1.color, bg1.palOfs = COLOR_4BPP, 0x00
			bg2.color, bg2.palOfs = COLOR_4BPP, 0x00
		case 3:
			bg1.color, bg1.palOfs = COLOR_8BPP, 0x00
			bg2.color, bg2.palOfs = COLOR_4BPP, 0x00
		case 4:
			bg1.color, bg1.palOfs = COLOR_8BPP, 0x00
			bg2.color, bg2.palOfs = COLOR_2BPP, 0x00
		case 5:
			bg1.color, bg1.palOfs = COLOR_4BPP, 0x00
			bg2.color, bg2.palOfs = COLOR_2BPP, 0x00
		}
		r.layers = []layer{r.backdrop, r.bg2, r.objs[0], r.bg1, r.objs[1], r.bg2h, r.objs[2], r.bg1h, r.objs[3]}

	case 6:
		bg1.color = COLOR_4BPP
		r.layers = []layer{r.backdrop, r.objs[0], r.bg1, r.objs[1], r.objs[2], r.bg1h, r.objs[3]}

	case 7:
		bg1.color = COLOR_8BPP
		bg2.color = COLOR_7BPP
		r.layers = []layer{r.backdrop, r.bg2, r.objs[0], r.bg1, r.objs[1], r.bg2h, r.objs[2], r.objs[3]}
	}
}
