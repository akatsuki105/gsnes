package core

import (
	"fmt"

	"github.com/pokemium/gsnes/core/scheduler"
)

type dmaChan struct {
	c      *sfc
	idx    int // 0, 1, 2, .., 7
	isHdma bool
	event  scheduler.Event

	param uint8 // DMAPx(0x43x0)
	bus   struct {
		a uint24 // A1Tx(0x43x2,3,4)
		b uint8  // BBADx(0x43x1)
	}
	dasx uint24 // DASx(0x43x5,6,7)

	// For HDMA
	doTransfer bool
	a2ax       uint16 // A2Ax(0x43x8,9)
	ntrlx      _ntrlx // NTRLx(0x43xA)

	ram uint8 // UNUSEDx(0x43xB)
}

// NTRLx(0x43xA)
type _ntrlx struct {
	repeat bool
	lines  uint8
}

func ntrlx(val uint8) _ntrlx {
	return _ntrlx{
		repeat: bit(val, 7),
		lines:  val & 0x7F,
	}
}

func (r _ntrlx) u8() uint8 {
	val := r.lines & 0x7F
	if r.repeat {
		val |= 0x80
	}
	return val
}

func (d *dmaChan) reset() {
	d.event = *scheduler.NewEvent(EVENT_DMA, d.runGDMA, EVENT_DMA_PRIO|uint(d.idx))
	d.bus.a, d.bus.b = u24(0, 0), 0
	d.dasx = toU24(0)
}

func (d *dmaChan) runGDMA(cyclesLate int64) {
	d.isHdma = false

	c := d.c.dma
	w := d.c.w

	mode := d.param & 0b111
	inc := gdmaIncrements[(d.param>>3)&0b11]

	switch mode {
	// 1x1
	case 0:
		src, dst := d.srcdst(0)
		val := w.load8(src)
		w.store8(dst, val, nil)
		addCycle(w.cycles, MEDIUM)
		d.bus.a = d.bus.a.plus(inc)
		d.dasx.offset--

	// 2x1
	case 1:
		for i := 0; i < 2; i++ {
			src, dst := d.srcdst(i)
			val := w.load8(src)
			w.store8(dst, val, nil)
			addCycle(w.cycles, MEDIUM)
			d.bus.a = d.bus.a.plus(inc)
			d.dasx.offset--
			if d.dasx.offset == 0 {
				break
			}
		}

	// 1x2
	case 2, 6:
		for i := 0; i < 2; i++ {
			src, dst := d.srcdst(0)
			val := w.load8(src)
			w.store8(dst, val, nil)
			addCycle(w.cycles, MEDIUM)
			d.bus.a = d.bus.a.plus(inc)
			d.dasx.offset--
			if d.dasx.offset == 0 {
				break
			}
		}

	// 2x2
	case 3, 7:
		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				src, dst := d.srcdst(i)
				val := w.load8(src)
				w.store8(dst, val, nil)
				addCycle(w.cycles, MEDIUM)
				d.bus.a = d.bus.a.plus(inc)
				d.dasx.offset--
				if d.dasx.offset == 0 {
					break
				}
			}
		}

	// 4x1
	case 4:
		crash("GDMA Mode4 is not implemented")

	case 5:
		crash("GDMA Mode5 is not implemented")
	}

	if d.dasx.offset > 0 {
		d.c.s.ReSchedule(&d.event, -cyclesLate)
		return
	}

	// end
	c.gdmaen = setBit(c.gdmaen, d.idx, false) // The MDMAEN bits are cleared automatically at transfer completion.
	c.update()
}

func (d *dmaChan) runHDMA() int64 {
	d.isHdma = true
	defer func() { d.isHdma = false }()

	w := d.c.w
	cycles := int64(8)

	if d.ntrlx.u8() == 0 {
		return cycles
	}

	if d.doTransfer {
		mode := d.param & 0b111
		switch mode {
		// 1x1
		case 0:
			src, dst := d.srcdst(0)
			val := w.load8(src)
			w.store8(dst, val, nil)
			cycles += 8
			d.incrementHDMAbus(1)

		// 2x1
		case 1:
			for i := 0; i < 2; i++ {
				src, dst := d.srcdst(i)
				val := w.load8(src)
				w.store8(dst, val, nil)
				cycles += 8
				d.incrementHDMAbus(1)
			}

		// 1x2
		case 2, 6:
			for i := 0; i < 2; i++ {
				src, dst := d.srcdst(0)
				val := w.load8(src)
				w.store8(dst, val, nil)
				cycles += 8
				d.incrementHDMAbus(1)
			}

		// 2x2
		case 3, 7:
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					src, dst := d.srcdst(i)
					val := w.load8(src)
					w.store8(dst, val, nil)
					cycles += 8
					d.incrementHDMAbus(1)
				}
			}

		// 4x1
		case 4:
			for i := 0; i < 4; i++ {
				src, dst := d.srcdst(i)
				val := w.load8(src)
				w.store8(dst, val, nil)
				cycles += 8
				d.incrementHDMAbus(1)
			}

		case 5:
			crash("HDMA Mode5 is not implemented")
		}
	}

	// decrement line counter
	d.ntrlx.lines--
	d.doTransfer = d.ntrlx.repeat
	if d.ntrlx.lines == 0 {
		d.loadHDMATable()
	}

	return cycles
}

func (d *dmaChan) loadHDMATable() int64 {
	w := d.c.w
	cycles := int64(8)

	d.ntrlx = ntrlx(w.load8(u24(d.bus.a.bank, d.a2ax)))
	d.a2ax++

	if bit(d.param, 6) {
		// indirect
		d.dasx.offset = w.load16(u24(d.bus.a.bank, d.a2ax), nil)
		d.a2ax += 2
		cycles += 16
	}

	d.doTransfer = true
	return cycles
}

func (d *dmaChan) incrementHDMAbus(size int) {
	if bit(d.param, 6) {
		d.dasx = d.dasx.plus(size)
	} else {
		d.a2ax += uint16(size)
	}
}

func (d *dmaChan) srcdst(plus int) (src uint24, dst uint24) {
	a := d.bus.a
	b := u24(0, 0x2100+uint16(d.bus.b)).plus(plus)

	if d.isHdma {
		a = u24(d.bus.a.bank, d.a2ax)
		if bit(d.param, 6) {
			a = d.dasx
		}
	}

	if bit(d.param, 7) {
		return b, a
	}
	return a, b
}

func (d *dmaChan) String() string {
	if d.isHdma {
		crash("HDMA is not implemented in %v", d.c.w.lastInstAddr)
	}
	step := [4]string{"++", "  ", "--", "  "}[(d.param>>3)&0b11]
	src, dst := d.srcdst(0)
	mode := d.param & 0b111
	pc := d.c.w.r.pc
	s := fmt.Sprintf("DMA[%d] %v%s -> %v             C:%05x U:%02x in %v", d.idx, src, step, dst, d.dasx, mode, pc)
	switch dst.u32() {
	case 0x2122, 0x213B:
		pal := &d.c.ppu.pal
		s = fmt.Sprintf("DMA[%d] %v%s -> %v(CGRAM:%04x) C:%05x U:%02x in %v", d.idx, src, step, dst, pal.idx*2, d.dasx, mode, pc)
	case 0x2118, 0x2139:
		v := &d.c.ppu.vram
		s = fmt.Sprintf("DMA[%d] %v%s -> %v(VRAM:%04x)  C:%05x U:%02x in %v", d.idx, src, step, dst, v.idx*2, d.dasx, mode, pc)
	}
	return s
}
