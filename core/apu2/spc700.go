package apu

type spc700 struct {
	a         *apu
	r         reg
	cycles    *int64
	nextEvent *int64
}

func newSpc700(a *apu) *spc700 {
	return &spc700{
		a:         a,
		cycles:    &a.s.RelativeCycles,
		nextEvent: &a.s.NextEvent,
	}
}

func (c *spc700) reset() {
	c.r.reset()
}

func addCycle(cycle *int64, added int64) {
	if cycle == nil {
		return
	}
	*cycle += added
}

func zn(val uint8) (z, n bool) {
	return val == 0, val>>7 == 1
}

func (c *spc700) load8(addr uint16, cycles *int64) uint8 {
	return 0xFF
}

func (c *spc700) load16(addr uint16, cycles *int64) uint16 {
	lo := uint16(c.load8(addr, cycles))
	hi := uint16(c.load8(addr+1, cycles))
	return (hi << 8) | lo
}

type reg struct {
	a    uint8
	x, y uint8
	sp   uint8
	pc   uint16
	p    struct {
		// carry
		c bool

		// zero
		z bool

		// interrupt disable
		i bool

		// half carry
		h bool

		// break
		b bool

		// zero page location
		p bool

		// overflow
		v bool

		// negative
		n bool
	}
}

func (r *reg) reset() {
}
