package core

type increment uint8

const (
	INC_LOW increment = iota
	INC_HIGH
)

var rotates = [4]uint16{0, 8, 9, 10}

var increments = [4]uint16{1, 32, 64, 128}

type vram struct {
	buf        [VRAM_SIZE / 2]uint16
	idx        uint16    // Unit is word
	incType    increment // VMAIN.7
	rotate     uint16    // VMAIN.2-3
	incAmount  uint16    // VMAIN.0-1
	fresh      bool
	prefetched uint16
}

func (v *vram) reset() {
	for i := range v.buf {
		v.buf[i] = 0
	}
	v.idx = 0
	v.incType = INC_LOW
	v.rotate = rotates[0b11]
	v.incAmount = increments[0b11]
}

func (v *vram) read(hi bool) uint8 {
	val := v.prefetched
	result := uint8(val)
	t := INC_LOW
	if hi {
		result = uint8(val >> 8)
		t = INC_HIGH
	}

	if v.incType == t {
		if !v.fresh {
			v.idx += v.incAmount
			v.prefetch()
		}
	}
	v.fresh = false
	return result
}

func (v *vram) write(hi bool, val uint8) {
	idx := ror(v.idx, v.rotate) & 0x7FFF
	old := v.buf[idx]
	if hi {
		v.buf[idx] = (uint16(val) << 8) | (old & 0xFF)
		if v.incType == INC_HIGH {
			v.idx += v.incAmount
		}
		return
	}

	v.buf[idx] = (old & 0xFF00) | uint16(val)
	if v.incType == INC_LOW {
		v.idx += v.incAmount
	}
}

func (v *vram) prefetch() {
	idx := ror(v.idx, v.rotate) & 0x7FFF
	v.prefetched = v.buf[idx]
}

// address rotation
func ror(val uint16, rotate uint16) uint16 {
	switch rotate {
	case 8:
		a := val & 0b1111_1111_0000_0000
		y := val & 0b0000_0000_1110_0000
		x := val & 0b0000_0000_0001_1111
		return a | (y >> 5) | (x << 3)
	case 9:
		a := val & 0b1111_1110_0000_0000
		y := val & 0b0000_0001_1100_0000
		x := val & 0b0000_0000_0011_1110
		p := val & 0b1
		return a | (y >> 6) | (x << 3) | (p << 3)
	case 10:
		a := val & 0b1111_1100_0000_0000
		y := val & 0b0000_0011_1000_0000
		x := val & 0b0000_0000_0111_1100
		p := val & 0b11
		return a | (y >> 7) | (x << 3) | (p << 3)
	}

	return val
}
