package core

type reg struct {
	// Data bank
	db uint8

	// PSR: Processor Status Register
	p psr

	// Accumulator
	a uint16

	// Zeropage Offset
	d uint16

	// Stack pointer
	s uint16

	// Index X, Index Y, used for addressing offset
	x, y uint16

	// Program counter(24bit)
	pc uint24

	// 6502 Emulation Flag(E)
	emulation bool
}

func (r *reg) reset() {
	r.db = 0
	r.p.reset()
	r.a = 0
	r.d = 0
	r.s = 0x1ff
	r.x, r.y = 0, 0
	r.setEmulation(true)
}

func (r *reg) setEmulation(e bool) {
	r.emulation = e
	if e {
		r.p.m = true
		r.s = 0x0100 | (r.s & 0x00FF)
		r.x &= 0xFF
		r.y &= 0xFF
	}
}

func (r *reg) l(r16 *uint16, val uint8) {
	*r16 = (*r16 & 0xFF00) | uint16(val)
}

func (r *reg) vector(e exception) uint24 {
	switch e {
	case COP:
		if r.emulation {
			return u24(0, 0xFFF4)
		}
		return u24(0, 0xFFE4)
	case BRK:
		if r.emulation {
			return u24(0, 0xFFFE)
		}
		return u24(0, 0xFFE6)
	case ABORT:
		if r.emulation {
			return u24(0, 0xFFF8)
		}
		return u24(0, 0xFFE8)
	case NMI:
		if r.emulation {
			return u24(0, 0xFFFA)
		}
		return u24(0, 0xFFEA)
	case IRQ:
		if r.emulation {
			return u24(0, 0xFFFE)
		}
		return u24(0, 0xFFEE)
	case RESET:
		return u24(0, 0xFFFC)
	default:
		crash("invalid exception")
		return u24(0, 0xFFFC)
	}
}
