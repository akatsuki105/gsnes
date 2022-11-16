package core

// 00-3f,80-bf:2180-2183,4016-4017,4200-421f
func (w *w65816) readCPU(addr uint, defaultVal uint8) uint8 {
	switch addr {
	case 0x2180, 0x2181, 0x2182, 0x2183:
		return w.wram.readIO(addr&0b11, defaultVal)

	case 0x4016: // JOYA
		return 0b11
	case 0x4017: // JOYB
		return 0b11111

	case 0x4210: // RDNMI
		old := w.rdnmi
		openbus := w.bus.data & 0b0111_0000
		w.rdnmi = setBit(w.rdnmi, 7, false) // ACK
		return old | openbus

	case 0x4211: // TIMEUP
		val := w.timeup & 0x80
		openbus := w.bus.data & 0x7F
		w.timeup = setBit(w.timeup, 7, false)
		return val | openbus

	case 0x4212: // HVBJOY
		val := w.bus.data & 0b0011_1110
		val = setBit(val, 0, w.ajr) // is AJR busy? (AJR= Auto Joypad Read)
		val = setBit(val, 6, w.c.ppu.inHBlank)
		val = setBit(val, 7, w.c.ppu.inVBlank)
		return val

	case 0x4213: // RDIO
		return 0x00

	case 0x4214, 0x4215: // RDDIV
		return uint8(w.div.result >> (8 * (addr - 0x4214)))

	case 0x4216, 0x4217: // RDMPY
		return uint8(w.mul.result >> (8 * (addr - 0x4216)))

	case 0x4218, 0x4219: // JOY1
		return uint8(w.joypads[0].val >> (8 * (addr - 0x4218)))
	case 0x421A, 0x421B: // JOY2
		return uint8(w.joypads[1].val >> (8 * (addr - 0x421A)))
	case 0x421C, 0x421D: // JOY3
		return uint8(w.joypads[2].val >> (8 * (addr - 0x421C)))
	case 0x421E, 0x421F: // JOY4
		return uint8(w.joypads[3].val >> (8 * (addr - 0x421E)))
	}
	return defaultVal
}

// 00-3f,80-bf:2180-2183,4016-4017,4200-421f
func (w *w65816) writeCPU(addr uint, val uint8) {
	p := w.c.ppu

	switch addr {
	case 0x2180, 0x2181, 0x2182, 0x2183:
		w.wram.writeIO(addr&0b11, val)

	case 0x4016: // JOYWR
		// TODO

	case 0x4200: // NMITIMEN
		old := bit(w.nmitimen, 7) && bit(w.rdnmi, 7)
		w.nmitimen = val

		// If IRQ is disabled, ack IRQ
		hIrq, vIrq := bit(val, 4), bit(val, 5)
		if !hIrq && !vIrq {
			w.timeup = setBit(w.timeup, 7, false)
		}

		now := bit(w.nmitimen, 7) && bit(w.rdnmi, 7)
		if !old && now {
			w.nmiPending = true
		}

	case 0x4201: // WRIO
		prev := bit(w.wrio, 7)
		w.wrio = val
		if prev && !bit(w.wrio, 7) {
			w.c.ppu.latchHV()
		}

	case 0x4202: // WRMPYA
		w.mul.a = val

	case 0x4203: // WRMPYB
		cycles := int64(8)
		for i := 0; i < 8; i++ {
			if bit(w.mul.a, 7-i) {
				break
			}
			cycles--
		}

		schedule("Mul", w.c.s, func(_ int64) {
			w.mul.result = uint16(val) * uint16(w.mul.a)
			w.div.result = uint16(val)
		}, FAST*cycles)

	case 0x4204, 0x4205: // WRDIV
		dividend := w.div.dividend
		switch addr {
		case 0x4204:
			w.div.dividend = (dividend & 0xFF00) | uint16(val)
		case 0x4205:
			w.div.dividend = (uint16(val) << 8) | (dividend & 0x00FF)
		}

	case 0x4206: // WRDIVB
		rddiv, rdmpy := uint16(0xFFFF), w.div.dividend
		if val != 0 {
			// not zerodiv
			rddiv = w.div.dividend / uint16(val)
			rdmpy = w.div.dividend % uint16(val)
		}

		schedule("Div", w.c.s, func(_ int64) {
			w.div.result = rddiv
			w.mul.result = rdmpy
		}, FAST*16)

	case 0x4207, 0x4208: // HTIME
		switch addr - 0x4207 {
		case 0:
			p.htime = (p.htime & 0xFF00) | uint16(val)
		case 1:
			p.htime = (uint16(val&0b1) << 8) | (p.htime & 0x00FF)
		}
		if p.htime > 339 {
			crash("htime overflow (%d)", p.htime)
		}

	case 0x4209, 0x420A: // VTIME
		switch addr - 0x4209 {
		case 0:
			p.vtime = (p.vtime & 0xFF00) | uint16(val)
		case 1:
			p.vtime = (uint16(val&0b1) << 8) | (p.vtime & 0x00FF)
		}
		if p.vtime > 261 {
			crash("vtime overflow (%d)", p.vtime)
		}

	case 0x420B: // MDMAEN
		w.c.dma.gdmaen = val
		if w.c.dma.gdmaen != 0 {
			w.c.dma.pending = true
		}

	case 0x420C: // HDMAEN
		w.c.dma.hdmaen = val

	case 0x420D: // MEMSEL
		val &= 0x01
		w.ws2 = [2]int64{MEDIUM, FAST}[val]
	}
}
