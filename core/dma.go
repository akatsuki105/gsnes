package core

var gdmaIncrements = [4]int{1, 0, -1, 0}

type dmaController struct {
	c              *sfc
	gdmaen, hdmaen uint8
	chans          [8]dmaChan
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

// MDMAEN を見ていって次にすべき GDMA を開始する
func (c *dmaController) update() {
	ch := -1
	for i := 0; i < 8; i++ {
		if bit(c.gdmaen, 7-i) {
			ch = 7 - i
		}
	}
	if ch >= 0 {
		c.chans[ch].trigger(false)
		return
	}
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
			return uint8(ch.remaining)
		case 1:
			return uint8(ch.remaining >> 8)
		case 2:
			return 0
		}

	case 0x08, 0x09: // A2Ax
		return uint8(ch.table.addr.offset >> (8 * (r - 0x08)))

	case 0x0A: // NTRLx
		return ch.table.hdr

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
		old := ch.remaining
		switch r - 0x05 {
		case 0:
			ch.remaining = (old & 0x1FF00) | uint32(val)
		case 1:
			ch.remaining = uint32(val)<<8 | (old & 0x100FF)
		case 2:
			ch.table.bank = val
		}

	case 0x08, 0x09: // A2Ax
		switch r - 0x08 {
		case 0:
			// TODO
		case 1:

			// TODO
		}
	case 0x0A: // NTRLx
		// TODO

	case 0x0B, 0x0F: // UNUSEDx
		ch.ram = val
	}
}
