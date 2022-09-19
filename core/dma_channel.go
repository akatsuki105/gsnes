package core

import (
	"fmt"

	"github.com/pokemium/gsnes/core/scheduler"
)

type dmaChan struct {
	c   *sfc
	idx int // 0, 1, 2, .., 7

	param uint8 // DMAPx(0x43x0)
	ram   uint8 // UNUSEDx(0x43xB)
	event scheduler.Event
	bus   struct {
		a uint24 // A1Tx(0x43x2,0x43x3,0x43x4)
		b uint8  // BBADx(0x43x1)
	}
	remaining uint32 // DASx(GDMA) or 現在のHDMAテーブルエントリの残りユニット数(HDMA)

	// For HDMA
	isHdma bool
	table  struct {
		addr uint24 // A2Ax
		hdr  uint8  // NTRLx
		wait int64  // 次のテーブルエントリまでの待機行数
		bank uint8
	}
}

func (d *dmaChan) reset() {
	d.event = *scheduler.NewEvent(EVENT_DMA, d.gdmaTransfer, EVENT_DMA_PRIO|uint(d.idx))
	d.bus.a, d.bus.b = u24(0, 0), 0
	d.remaining = 0
}

// Make a reservation to start DMA transfer.
func (d *dmaChan) trigger(isHdma bool) {
	d.isHdma = isHdma
	if d.isHdma {
		d.remaining = 0x0
		d.table.addr, d.table.wait = d.bus.a, 0x0
		return
	}

	if d.remaining == 0 {
		d.remaining = 0x10000
	}
	d.c.s.ReSchedule(&d.event, FAST) // After $420b is written, the CPU gets one more CPU cycle before the pause
}

func (d *dmaChan) gdmaTransfer(cyclesLate int64) {
	c := d.c.dma
	w := d.c.w
	w.blocked = true

	mode := d.param & 0b111
	inc := gdmaIncrements[(d.param>>3)&0b11]

	switch mode {
	// 1x1
	case 0:
		if d.remaining > 0 {
			src, dst := d.srcdst(0)
			val := w.load8(src, nil)
			w.store8(dst, val, nil)
			addCycle(w.cycles, MEDIUM)
			d.bus.a = d.bus.a.plus(inc)
			d.remaining--
		}

	// 2x1
	case 1:
		for i := 0; i < 2; i++ {
			if d.remaining > 0 {
				src, dst := d.srcdst(i)
				val := w.load8(src, nil)
				w.store8(dst, val, nil)
				addCycle(w.cycles, MEDIUM)
				d.bus.a = d.bus.a.plus(inc)
				d.remaining--
			}
		}

	// 1x2
	case 2, 6:
		for i := 0; i < 2; i++ {
			if d.remaining > 0 {
				src, dst := d.srcdst(0)
				val := w.load8(src, nil)
				w.store8(dst, val, nil)
				addCycle(w.cycles, MEDIUM)
				d.bus.a = d.bus.a.plus(inc)
				d.remaining--
			}
		}

	// 2x2
	case 3, 7:
		crash("A -> B DMA Mode3 is not implemented")

	// 1x4
	case 4:
		crash("A -> B DMA Mode4 is not implemented")

	case 5:
		crash("A -> B DMA Mode5 is not implemented")
	}

	if d.remaining > 0 {
		d.c.s.ReSchedule(&d.event, -cyclesLate)
		return
	}

	// end
	w.blocked = false
	c.gdmaen = setBit(c.gdmaen, d.idx, false)
	c.update()
}

func (d *dmaChan) runHDMA() {
	c := d.c.dma
	if d.table.wait > 0 {
		d.table.wait--
		return
	}

	w := d.c.w

	// Next table entry
	if d.remaining == 0 {
		d.table.hdr = w.load8(d.table.addr, nil)
		d.table.addr.offset++

		// end
		if d.table.hdr == 0 {
			w.blocked = false
			c.hdmaen = setBit(c.hdmaen, d.idx, false)
			return
		}

		if bit(d.table.hdr, 7) {
			d.remaining = uint32(d.table.hdr & 0x7F)
			d.table.wait = 0x0
		} else {
			d.remaining = 0x1
			d.table.wait = int64(d.table.hdr&0x7F) - 1
		}
	}

	w.blocked = true

	mode := d.param & 0b111
	size := uint16(0)
	switch mode {
	// 1x1
	case 0:
		size = 1
		src, dst := d.srcdst(0)
		val := w.load8(src, nil)
		w.store8(dst, val, nil)
		addCycle(w.cycles, MEDIUM)

	// 2x1
	case 1:
		size = 2
		for i := 0; i < 2; i++ {
			src, dst := d.srcdst(i)
			val := w.load8(src, nil)
			w.store8(dst, val, nil)
			addCycle(w.cycles, MEDIUM)
		}

	// 1x2
	case 2, 6:
		size = 2
		for i := 0; i < 2; i++ {
			src, dst := d.srcdst(0)
			val := w.load8(src, nil)
			w.store8(dst, val, nil)
			addCycle(w.cycles, MEDIUM)
		}

	// 2x2
	case 3, 7:
		crash("HDMA Mode3 is not implemented")

	// 1x4
	case 4:
		crash("HDMA Mode4 is not implemented")

	case 5:
		crash("HDMA Mode5 is not implemented")
	}
	d.remaining--

	if bit(d.param, 6) {
		if d.remaining == 0 {
			d.table.addr.offset++
		}
	} else {
		d.table.addr.offset += size
	}
}

func (d *dmaChan) srcdst(plus int) (src uint24, dst uint24) {
	a := d.bus.a
	b := u24(0, 0x2100+uint16(d.bus.b)).plus(plus)

	if d.isHdma {
		a = d.table.addr
		if bit(d.param, 6) {
			ofs := d.c.w.load16(a, nil)
			progress := uint16(uint32(d.table.hdr&0x7F) - d.remaining)
			a = u24(d.table.bank, ofs+progress)
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
	s := fmt.Sprintf("DMA[%d] %v%s -> %v             C:%05x U:%02x in %v", d.idx, src, step, dst, d.remaining, mode, pc)
	switch dst.u32() {
	case 0x2122, 0x213B:
		pal := &d.c.ppu.pal
		s = fmt.Sprintf("DMA[%d] %v%s -> %v(CGRAM:%04x) C:%05x U:%02x in %v", d.idx, src, step, dst, pal.idx*2, d.remaining, mode, pc)
	case 0x2118, 0x2139:
		v := &d.c.ppu.vram
		s = fmt.Sprintf("DMA[%d] %v%s -> %v(VRAM:%04x)  C:%05x U:%02x in %v", d.idx, src, step, dst, v.idx*2, d.remaining, mode, pc)
	}
	return s
}
