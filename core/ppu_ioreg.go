package core

func (p *ppu) readIO(addr uint, defaultVal uint8) uint8 {
	switch addr {
	case 0x15, 0x16, 0x18, 0x19: // PPU1
		return p.openbus[0]

	case 0x34, 0x35, 0x36: // MPY
		p.openbus[0] = uint8(p.mpy.result.u32() >> (8 * (addr - 0x34)))
		return p.openbus[0]

	case 0x37: // SLHV
		p.latchHV()
		return p.c.w.bus.data // CPU open bus

	case 0x38: // RDOAM
		return p.oam.read()

	case 0x39, 0x3A: // RDVRAM
		p.openbus[0] = p.vram.read(addr == 0x3A)
		return p.openbus[0]

	case 0x3B: // RDCGRAM
		second := p.pal.is2ndAccess
		val := p.pal.read()
		if second {
			val = setBit(val, 7, bit(p.openbus[1], 7))
		}
		p.openbus[1] = val
		return val

	case 0x3C, 0x3D: // OPHCT, OPVCT
		r := &p.opct[addr-0x3C]
		if r.second {
			r.second = false
			val := uint8(r.count>>8) & 0b1
			val |= p.openbus[1] & 0xFE
			p.openbus[1] = val
			return p.openbus[1]
		}
		r.second = true
		p.openbus[1] = uint8(r.count)
		return p.openbus[1]

	case 0x3E: // STAT77
		val := uint8(0x01)
		p.openbus[0] = val
		return val

	case 0x3F: // STAT78
		r := &p.stat78
		val := uint8(0x03)
		val = setBit(val, 6, r.latch)
		r.latch = false
		p.opct[0].second, p.opct[1].second = false, false
		p.openbus[1] = setBit(val, 7, r.interlace)
		return p.openbus[1]

	default:
		return defaultVal
	}
}

func (p *ppu) writeIO(addr uint, val uint8) {
	v := &p.vram
	bg1, bg2, bg3, bg4, objs := p.r.bg1, p.r.bg2, p.r.bg3, p.r.bg4, p.r.objs[:]

	switch addr {
	case 0x00: // INIDISP
		p.inFBlank = bit(val, 7)

	case 0x01: // OBSEL
		o := &p.oam
		o.size = objSizes[val>>5]
		o.tileDataAddr = uint(val&0b111) * (8 * KB)
		o.gap = uint((val>>3)&0b11) * (4 * KB)

	case 0x02, 0x03: // OAMADD
		o := &p.oam
		switch addr - 0x02 {
		case 0:
			o.reload = (o.reload & 0xFF00) | uint16(val)
		case 1:
			o.reload = (uint16(val) << 8) | (o.reload & 0x00FF)
			o.rotate = bit(val, 7)
		}
		o.reload &= 0x01FF

	case 0x04: // OAMDATA
		p.oam.write(val)

	case 0x05: // BGMODE
		p.mode = val & 0b111
		p.r.bg3a = bit(val, 3)
		p.r.setBgMode(p.mode)

		tilesizes := [2]uint16{8, 16}
		bg1.tilesize = tilesizes[(val>>4)&0b1]
		bg2.tilesize = tilesizes[(val>>5)&0b1]
		bg3.tilesize = tilesizes[(val>>6)&0b1]
		bg4.tilesize = tilesizes[(val>>7)&0b1]

	case 0x06: // MOSAIC
		// TODO

	case 0x07, 0x08, 0x09, 0x0A: // BGNSC
		bg := [4]*bg{p.r.bg1, p.r.bg2, p.r.bg3, p.r.bg4}[addr-0x07]
		bg.size = [2]uint16{32, 32}
		if bit(val, 0) {
			bg.size[0] = 64
		}
		if bit(val, 1) {
			bg.size[1] = 64
		}
		bg.tilemapAddr = (uint(val>>2) * (2 * KB)) & 0xFFFF

	case 0x0B: // BGNBA12
		p.r.bg1.tileDataAddr = uint(val&0b1111) * (8 * KB)
		p.r.bg2.tileDataAddr = uint((val>>4)&0b1111) * (8 * KB)
	case 0x0C: // BGNBA34
		p.r.bg3.tileDataAddr = uint(val&0b1111) * (8 * KB)
		p.r.bg4.tileDataAddr = uint((val>>4)&0b1111) * (8 * KB)

	case 0x0D, 0x0F, 0x11, 0x13: // BGnHOFS
		n := (addr - 0x0D) / 2
		bg := [4]*bg{bg1, bg2, bg3, bg4}[n]
		prev := uint16(bg.sc[0].prev)
		bg.sc[0].prev = val
		bg.sc[0].reg = (uint16(val) << 8) | (prev & 0xFFF8) | ((bg.sc[0].reg >> 8) & 7)
		bg.sc[0].val = bg.sc[0].reg & 0x3FF

	case 0x0E, 0x10, 0x12, 0x14: // BGnVOFS
		n := (addr - 0x0E) / 2
		bg := [4]*bg{bg1, bg2, bg3, bg4}[n]
		prev := uint16(bg.sc[1].prev)
		bg.sc[1].prev = val
		bg.sc[1].reg = (uint16(val)<<8 | prev)
		bg.sc[1].val = bg.sc[1].reg & 0x3FF

	case 0x15: // VMAIN
		v.incType = increment(val >> 7)
		v.rotate = rotates[(val>>2)&0b11]
		v.incAmount = increments[val&0b11]

	case 0x16, 0x17: // VMADD
		old := v.idx
		switch addr - 0x16 {
		case 0:
			v.idx = ((v.idx & 0xFF00) | uint16(val)) & 0x7FFF
		case 1:
			v.idx = ((uint16(val) << 8) | (v.idx & 0xFF)) & 0x7FFF
		}
		if old != v.idx {
			v.fresh = true
			v.prefetch()
		}

	case 0x18, 0x19: // VMDATA
		p.vram.write(addr == 0x19, val)

	case 0x1B: // M7A
		mul := func(a int) {
			b := int(int8(p.mpy.b))
			ab := uint32(a * b)
			p.mpy.result = toU24(ab)
		}

		if p.mpy.second {
			p.mpy.second = false
			p.mpy.a = uint16(val)<<8 | (p.mpy.a & 0x00FF)
			mul(int(int16(p.mpy.a)))
			return
		}
		p.mpy.second = true
		p.mpy.a = (p.mpy.a & 0xFF00) | uint16(val)
		mul(int(int16(p.mpy.a)))

	case 0x1C: // M7B
		p.mpy.b = val
		a := int(int16(p.mpy.a))
		b := int(int8(p.mpy.b))
		ab := uint32(a * b)
		p.mpy.result = toU24(ab)

	case 0x21: // CGADD
		p.pal.setAddr(val)
	case 0x22: // CGDATA
		p.pal.write(val)

	case 0x23, 0x24, 0x25: // W12SEL, W34SEL, WOBJSEL
		i := (addr - 0x23) * 2
		p.r.w.win1.mask[i+0] = windowMask(val >> 0)
		p.r.w.win2.mask[i+0] = windowMask(val >> 2)
		p.r.w.win1.mask[i+1] = windowMask(val >> 4)
		p.r.w.win2.mask[i+1] = windowMask(val >> 6)

	case 0x26: // WH0
		p.r.w.win1.left = val
	case 0x27: // WH1
		p.r.w.win1.right = val
	case 0x28: // WH2
		p.r.w.win2.left = val
	case 0x29: // WH3
		p.r.w.win2.right = val

	case 0x2A: // WBGLOG
		p.r.w.logic[0] = maskLogic((val >> 0) & 0b11)
		p.r.w.logic[1] = maskLogic((val >> 2) & 0b11)
		p.r.w.logic[2] = maskLogic((val >> 4) & 0b11)
		p.r.w.logic[3] = maskLogic((val >> 6) & 0b11)

	case 0x2B: // WOBJLOG
		p.r.w.logic[4] = maskLogic((val >> 0) & 0b11)
		p.r.w.logic[5] = maskLogic((val >> 2) & 0b11)

	case 0x2C: // TM
		bg1.mainsc = bit(val, 0)
		bg2.mainsc = bit(val, 1)
		bg3.mainsc = bit(val, 2)
		bg4.mainsc = bit(val, 3)
		mainsc := bit(val, 4)
		for i := range objs {
			objs[i].mainsc = mainsc
		}
	case 0x2D: // TS
		bg1.subsc = bit(val, 0)
		bg2.subsc = bit(val, 1)
		bg3.subsc = bit(val, 2)
		bg4.subsc = bit(val, 3)
		subsc := bit(val, 4)
		for i := range objs {
			objs[i].subsc = subsc
		}
	}
}
