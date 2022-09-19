package core

import "github.com/pokemium/iro"

type palette struct {
	buf         [256]iro.RGB555
	idx         uint8
	is2ndAccess bool
	lastWritten uint8
}

func (p *palette) setAddr(val uint8) {
	p.idx = val
	p.is2ndAccess = false
}

func (p *palette) read() uint8 {
	second := p.is2ndAccess

	rgb555 := p.buf[p.idx]
	if second {
		// odd(2nd)
		p.idx++
		p.is2ndAccess = false
		val := uint8((rgb555 >> 8) & 0x7f)
		return val
	}

	// even(1st)
	p.is2ndAccess = true
	return uint8(rgb555)
}

func (p *palette) write(val uint8) {
	second := p.is2ndAccess
	memo := p.lastWritten
	p.lastWritten = val

	if second {
		// odd(2nd)
		val &= 0x7f
		p.buf[p.idx] = iro.RGB555(uint16(val)<<8 | uint16(memo))
		p.idx++
		p.is2ndAccess = false
		return
	}

	// even(1st)
	p.is2ndAccess = true
}
