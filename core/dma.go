package core

/*
GDMA timing refers to Anomie's SNES Timing Doc.
https://www.romhacking.net/documents/199/
*/

var gdmaIncrements = [4]int{1, 0, -1, 0}

type dmaController struct {
	c              *sfc
	gdmaen, hdmaen uint8
	chans          [8]dmaChan

	// For GDMA
	pending bool  // GDMA pending
	counter int64 // GDMA cycle
}

func newDmaController(c *sfc) *dmaController {
	d := &dmaController{
		c: c,
	}
	for i := range d.chans {
		ch := &d.chans[i]
		ch.idx, ch.c = i, c
	}
	d.reset()
	return d
}

func (c *dmaController) reset() {
	for i := range c.chans {
		c.chans[i].reset()
	}
}

func (c *dmaController) cpuCounter() int64 {
	return c.c.s.Cycle() & 0b111
}

func (c *dmaController) initGDMA() {
	c.pending = false

	w := c.c.w
	w.lock = setBit(w.lock, BLOCK_DMA, true) // CPU block

	// wait for 8x cycles
	c.counter = 8 - c.cpuCounter()
	addCycle(w.cycles, c.counter)

	// DMA initialization
	c.counter += 8
	addCycle(w.cycles, 8)

	c.update()
}

// MDMAEN を見ていって次にすべき GDMA を開始する
func (c *dmaController) update() {
	w := c.c.w

	ch := -1
	for i := 0; i < 8; i++ {
		if bit(c.gdmaen, 7-i) {
			ch = 7 - i
		}
	}
	if ch >= 0 {
		ch := &c.chans[ch]
		c.c.s.ReSchedule(&ch.event, 8) // each channel takes 8 cycles for initialization
		return
	}

	// All GDMA complete
	addCycle(w.cycles, w.waitstate-(c.counter%w.waitstate)) // To reach a whole number of CPU Clock cycles since the pause
	w.lock = setBit(w.lock, BLOCK_DMA, false)
}

func (c *dmaController) reloadHDMA() int64 {
	cycles := int64(0)
	for i := 0; i < 8; i++ {
		if bit(c.hdmaen, i) {
			d := &c.chans[i]
			d.a2ax = d.bus.a.offset
			cycles += d.loadHDMATable()
		}
	}
	return cycles
}

// addr: 0x00..7F
func (c *dmaController) readIO(addr uint, defaultVal uint8) uint8 {
	ch := &c.chans[addr>>4]

	switch r := addr & 0b1111; r {
	case 0x00: // DMAPx
		return ch.param

	case 0x01: // BBADx
		return ch.bus.b

	case 0x02, 0x03, 0x04: // A1Tx
		return uint8(ch.bus.a.u32() >> (8 * (addr - 0x02)))

	case 0x05, 0x06, 0x07: // DASx
		switch r - 0x05 {
		case 0:
			return uint8(ch.dasx.offset)
		case 1:
			return uint8(ch.dasx.offset >> 8)
		case 2:
			return ch.dasx.bank
		}

	case 0x08, 0x09: // A2Ax
		return uint8(ch.a2ax >> (8 * (r - 0x08)))

	case 0x0A: // NTRLx
		return ch.ntrlx.u8()

	case 0x0B, 0x0F: // UNUSEDx
		return ch.ram
	}
	return defaultVal
}

// addr: 0x00..7F
func (c *dmaController) writeIO(addr uint, val uint8) {
	ch := &c.chans[addr>>4]

	switch r := addr & 0b1111; r {
	case 0x00: // DMAPx
		ch.param = val

	case 0x01: // BBADx
		ch.bus.b = val

	case 0x02, 0x03, 0x04: // A1Tx
		a := &ch.bus.a
		old := a.offset
		switch r - 0x02 {
		case 0:
			a.offset = (old & 0xFF00) | uint16(val)
		case 1:
			a.offset = uint16(val)<<8 | (old & 0xFF)
		case 2:
			a.bank = val
		}

	case 0x05, 0x06, 0x07: // DASx
		old := uint16(ch.dasx.offset)
		switch r - 0x05 {
		case 0:
			ch.dasx.offset = (old & 0xFF00) | uint16(val)
		case 1:
			ch.dasx.offset = uint16(val)<<8 | (old & 0x00FF)
		case 2:
			ch.dasx.bank = val
		}

	case 0x08, 0x09: // A2Ax
		old := ch.a2ax
		switch r - 0x08 {
		case 0:
			ch.a2ax = (old & 0xFF00) | uint16(val)
		case 1:
			ch.a2ax = uint16(val)<<8 | (old & 0x00FF)
		}

	case 0x0A: // NTRLx
		ch.ntrlx = ntrlx(val)

	case 0x0B, 0x0F: // UNUSEDx
		ch.ram = val
	}
}
