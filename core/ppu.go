package core

import (
	"fmt"

	"github.com/pokemium/gsnes/core/scheduler"
	"github.com/pokemium/iro"
)

var fblankScreen = [HORIZONTAL * VERTICAL]iro.RGB555{}

type ppu struct {
	c              *sfc
	vram           vram
	pal            palette
	oam            oam
	mode           uint8  // BGMODE
	hcount, vcount uint16 // 0..339, 0..261
	htime, vtime   uint16 // 0x4207, 0x4209
	event          scheduler.Event
	irqEvent       scheduler.Event
	nmiEvent       scheduler.Event

	inHBlank, inVBlank bool // HVBJOY.6, HVBJOY.7

	inFBlank bool // INIDISP.7
	r        *renderer

	openbus [2]uint8 // 0: PPU1, 1: PPU2

	mpy struct {
		a      uint16
		second bool
		b      uint8
		result uint24
	}
	stat78 struct {
		latch     bool // bit6
		interlace bool // 2nd frame on interlace
	}
	// 0: OPHCT, 1: OPVCT
	opct [2]struct {
		count  uint16 // latched value
		second bool
	}
}

func newPpu(c *sfc) *ppu {
	p := &ppu{
		c:     c,
		event: *scheduler.NewEvent(EVENT_VIDEO, nil, EVENT_VIDEO_PRIO),
		pal:   *newPalette(),
	}
	p.r = newRenderer(&p.vram, &p.pal, &p.oam)
	p.opct[0].count, p.opct[1].count = 0x1FF, 0x1FF

	p.irqEvent = *scheduler.NewEvent(EVENT_IRQ, p.requestIrq, EVENT_IRQ_PRIO)
	p.nmiEvent = *scheduler.NewEvent(EVENT_NMI, p.setNMI, EVENT_NMI_PRIO)

	return p
}

func (p *ppu) reset() {
	p.r.reset()
	p.vram.reset()
	p.oam.reset()

	for i := range p.pal.buf {
		p.pal.buf[i] = iro.RGB555(0x0000)
	}

	p.hcount, p.vcount = INIT_CYCLE/4+1, 0
	p.inHBlank, p.inVBlank = false, false

	p.event.Callback = p.incrementHCount
	p.c.s.ReSchedule(&p.event, 4)
}

// (x, y) = (0, any)
func (p *ppu) newline(cyclesLate int64) {
	p.vcount++

	switch p.vcount {
	case VERTICAL + 1:
		// start VBlank
		p.setVBlank(true, cyclesLate)
		p.c.earlyExit = true
		p.c.frame++

	case TOTAL_SCANLINE:
		p.vcount = 0
		p.setVBlank(false, cyclesLate) // End of VBlank
	}
}

func (p *ppu) setNMI(cyclesLate int64) {
	w := p.c.w
	old := bit(w.nmitimen, 7) && bit(w.rdnmi, 7)
	p.c.w.rdnmi = setBit(p.c.w.rdnmi, 7, true)

	// NMITIMEN.7 & RDNMI.7 が 0 から 1 になったなら 内部NMIフラグをセットする
	now := bit(w.nmitimen, 7) && bit(w.rdnmi, 7)
	if !old && now {
		w.nmiPending = true
	}
}

// (x, y) = (274, any)
func (p *ppu) startHBlank() {
	p.inHBlank = true
	if p.vcount > 0 && p.vcount < VERTICAL && !p.inFBlank {
		p.r.drawScanline(p.vcount)
	}
}

func (p *ppu) setVBlank(b bool, cyclesLate int64) {
	if b {
		// (H, V) = (0, 225)
		p.inVBlank = true
		p.c.s.Schedule(&p.nmiEvent, 2-cyclesLate) // (H, V) = (0.5, 225)
		return
	}

	// (H, V) = (0, 0)
	p.inVBlank = false
	p.c.w.rdnmi = setBit(p.c.w.rdnmi, 7, false) // Auto ACK
}

func (p *ppu) incrementHCount(cyclesLate int64) {
	w := p.c.w
	dma := w.c.dma

	p.hcount++
	switch p.hcount {
	case 1:
		p.inHBlank = false
		if p.vcount == 0 {
			// toggle interlace frame
			p.stat78.interlace = !p.stat78.interlace
		}

	case 6:
		if p.vcount == 0 {
			cycles := dma.reloadHDMA()
			if cycles > 0 {
				w.block(BLOCK_HDMA, cycles, cyclesLate)
			}
		}

	case 10:
		if p.vcount == 225 {
			if !p.inFBlank {
				p.oam.addr = (p.oam.reload << 1) & 0x3FF
			}
		}

	case 32:
		if p.vcount == 225 {
			if autoJoypadRead := bit(w.nmitimen, 0); autoJoypadRead {
				w.ajr = true
				for i := range w.joypads {
					j := &w.joypads[i]
					for i, b := range j.buf {
						j.val = setBit(j.val, i+4, b)
					}
				}
			}
		}

	case 92:
		if p.vcount == 225 {
			w.ajr = false
		}

	// DRAM Refresh
	case 133:
		w.block(BLOCK_DRAM, 40, cyclesLate)

	case 274:
		p.startHBlank()

	case 278:
		if p.vcount < 225 {
			// (H, V) = (278, 0..224)
			// perform HDMA transfers
			hdmaen := dma.hdmaen

			cycles := int64(0)
			if hdmaen != 0 {
				cycles = 18
			}

			for i := 0; i < 8; i++ {
				if bit(hdmaen, i) {
					cycles += dma.chans[i].runHDMA()
				}
			}

			if cycles > 0 {
				w.block(BLOCK_HDMA, cycles, cyclesLate)
			}
		}

	case 340:
		// new line
		p.hcount = 0
		p.newline(cyclesLate)
	}
	p.checkIrq(cyclesLate)

	p.c.s.Schedule(&p.event, 4-cyclesLate)
}

func (p *ppu) checkIrq(cyclesLate int64) {
	nmitimen := p.c.w.nmitimen
	requested := false
	after := int64(0)

	switch (nmitimen >> 4) & 0b11 {
	case 1:
		// H-IRQ
		requested = p.hcount == p.htime
		after = 14 // (H, V) = (HTIME+3.5, any)

	case 2:
		// V-IRQ
		requested = p.hcount == 0 && p.vcount == p.vtime
		after = 10 // (H, V) = (2.5, VTIME)

	case 3:
		// HV-IRQ
		requested = p.hcount == p.htime && p.vcount == p.vtime
		after = 14 // (H, V) = (HTIME+3.5, VTIME)
	}

	if requested {
		p.c.s.Schedule(&p.irqEvent, after-cyclesLate)
	}
}

func (p *ppu) requestIrq(cyclesLate int64) {
	w := p.c.w
	w.timeup = setBit(w.timeup, 7, true)
}

func (p *ppu) latchHV() {
	p.opct[0].count = p.hcount
	p.opct[1].count = p.vcount
	p.stat78.latch = true
}

func (p *ppu) Status(region string) string {
	s := p.c

	switch region {
	case "PPU":
		nmitimen := s.w.nmitimen
		nmi := bit(nmitimen, 7)
		hvirq := [4]string{"--", "H-", "-V", "HV"}[(nmitimen>>4)&0b11]

		mode := p.mode

		return fmt.Sprintf(
			"H: %03d, V: %03d, F: %03d\nNMI: %v, IRQ: %s, HTIME: %03d, VTIME: %03d\nMode: %d",
			p.hcount, p.vcount, s.frame, nmi, hvirq, p.htime, p.vtime, mode,
		)

	case "SCREEN":
		bg1, bg2, bg3, bg4 := p.r.bg1, p.r.bg2, p.r.bg3, p.r.bg4
		bg3a := ""
		if p.r.bg3a {
			bg3a = "a"
		}
		return fmt.Sprintf(
			`BG1:
  %v
BG2:
  %v
BG3%s:
  %v
BG4:
  %v`, bg1, bg2, bg3a, bg3, bg4,
		)

	case "OAM":
		oam1 := p.oam.tileDataAddr * 2
		oam2 := oam1 + 0x2000 + (2 * p.oam.gap)

		addr := fmt.Sprintf("Tiles: 0x%04X, 0x%04X", oam1, oam2)
		size := fmt.Sprintf("Large: (%d, %d)", p.oam.size[1], p.oam.size[1])
		objs := ""
		for i := 0; i < 16; i++ {
			objs += fmt.Sprintf("    OBJ%03d: %s\n", i, &p.oam.objs[i])
		}
		return fmt.Sprintf(
			"OBJ\n  %s\n  %s\n%s", addr, size, objs,
		)

	default:
		return ""
	}
}
