package core

import "fmt"

type flag struct {
	f   byte
	bit bool
}

// PSR: Processor Status Register
//
//	7 6 5 4 3 2 1 0
//
// -------------------
//
//	n v m x d i z c
type psr struct {
	r *reg

	// carry
	c bool

	// zero
	z bool

	// interrupt disable
	i bool

	// decimal mode
	d bool

	// index register mode
	x bool

	// accumulator mode
	// In this mode, switches A to 8bit mode
	m bool

	// overflow
	v bool

	// negative
	n bool
}

func (p *psr) String() string {
	bitfield := ""

	n := map[bool]string{true: "N", false: "n"}
	bitfield += n[p.n]

	v := map[bool]string{true: "V", false: "v"}
	bitfield += v[p.v]

	if p.r.emulation {
		bitfield += "1"

		b := map[bool]string{true: "B", false: "b"}
		bitfield += b[p.x]
	} else {
		m := map[bool]string{true: "M", false: "m"}
		bitfield += m[p.m]

		x := map[bool]string{true: "X", false: "x"}
		bitfield += x[p.x]
	}

	d := map[bool]string{true: "D", false: "d"}
	bitfield += d[p.d]

	i := map[bool]string{true: "I", false: "i"}
	bitfield += i[p.i]

	z := map[bool]string{true: "Z", false: "z"}
	bitfield += z[p.z]

	c := map[bool]string{true: "C", false: "c"}
	bitfield += c[p.c]

	return fmt.Sprintf("%02X (%s)", p.pack(), bitfield)
}

func (p *psr) reset() {
	p.c = false
	p.z = false
	p.i = true
	p.d = false
	p.x = true // ソース不明
	p.m = false
	p.v = false
	p.n = false
}

func (p *psr) pack() uint8 {
	packed := uint8(0)
	packed = setBit(packed, 0, p.c)
	packed = setBit(packed, 1, p.z)
	packed = setBit(packed, 2, p.i)
	packed = setBit(packed, 3, p.d)
	packed = setBit(packed, 4, p.x)
	packed = setBit(packed, 5, p.m)
	packed = setBit(packed, 6, p.v)
	packed = setBit(packed, 7, p.n)

	return packed
}

func (p *psr) setPacked(val uint8) {
	p.c = bit(val, 0)
	p.z = bit(val, 1)
	p.i = bit(val, 2)
	p.d = bit(val, 3)
	p.x = bit(val, 4)
	p.m = bit(val, 5)
	p.v = bit(val, 6)
	p.n = bit(val, 7)

	if p.x {
		p.r.x &= 0xFF
		p.r.y &= 0xFF
	}
}

func (p *psr) setFlags(flags ...flag) {
	for _, f := range flags {
		switch f.f {
		case 'c':
			p.c = f.bit
		case 'z':
			p.z = f.bit
		case 'i':
			p.i = f.bit
		case 'd':
			p.d = f.bit
		case 'x':
			p.x = f.bit
		case 'm':
			p.m = f.bit
		case 'v':
			p.v = f.bit
		case 'n':
			p.n = f.bit
		}
	}

	if p.x {
		p.r.x &= 0xFF
		p.r.y &= 0xFF
	}
}

func zn(val uint16, unit int) (z, n flag) {
	switch unit {
	case 8:
		val &= 0xff
		return flag{'z', val == 0}, flag{'n', val>>7 == 1}
	case 16:
		return flag{'z', val == 0}, flag{'n', val>>15 == 1}
	}

	panic(fmt.Errorf("invalid unit: %d", unit))
}
