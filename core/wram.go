package core

// 0x7E:0000..7F:FFFF
type wram struct {
	buf  []uint8
	addr uint32 // WMADD
}

func newWram() *wram {
	w := &wram{
		buf: make([]uint8, 128*KB),
	}
	return w
}

func (w *wram) reset() {
	for i := range w.buf {
		w.buf[i] = 0
	}
}

// addr: 00-01:0000-FFFF
func (w *wram) read(addr uint, defaultVal uint8) uint8 {
	return w.buf[addr]
}

// addr: 00-01:0000-FFFF
func (w *wram) write(addr uint, val uint8) {
	w.buf[addr] = val
}

// addr: 0,1,2,3
func (w *wram) readIO(addr uint, defaultVal uint8) uint8 {
	switch addr {
	case 0: // WMDATA
		addr := w.addr
		w.addr = (w.addr + 1) & 0x1_FFFF
		return w.buf[addr]
	}
	return defaultVal
}

// addr: 0,1,2,3
func (w *wram) writeIO(addr uint, val uint8) {
	switch addr {
	case 0: // WMDATA
		addr := int(w.addr)
		w.addr = (w.addr + 1) & 0x1_FFFF
		if addr >= len(w.buf) {
			crash("invalid WRAM address: 0x%X", addr)
		}
		w.buf[addr] = val
	case 1: // WMADDL
		w.addr &= 0x1_FF00
		w.addr |= uint32(val)
	case 2: // WMADDM
		w.addr &= 0x1_00FF
		w.addr |= (uint32(val) << 8)
	case 3: // WMADDH
		val &= 0b1
		w.addr &= 0x0_FFFF
		w.addr |= uint32(val) << 16
	}
}
