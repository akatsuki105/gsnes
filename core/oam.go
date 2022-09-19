package core

import "fmt"

/*
	OAM1 tiledata for tile number 0..255, OAM2 tiledata for tile number 256..511. (tile number is specified at OAM entry)
	  OAM1 tiledata Base Addr: obsel bit0..2 (if N, base addr is N * 16KB)
	  OAM2 tiledata Base Addr: (OAM1 base addr) + (gap * 8KB) (gap is obsel bit3..4)
*/

var objSizes = [6][2]int{
	{8, 16}, // small: 8x8, large: 16x16
	{8, 32},
	{8, 64},
	{16, 32},
	{16, 64},
	{32, 64},
}

type oam struct {
	reload uint16 // Reload value (OAMADD.0-8)
	addr   uint16 // unit is byte
	rotate bool   // OAMADD.15

	objs         [128]obj
	buf          [544]uint8 // for read
	memo         uint8
	tileDataAddr uint   // OBSEL.0-2 (unit is word)
	gap          uint   // OBSEL.3-4 (unit is word)
	size         [2]int // OBSEL.5-7
}

type obj struct {
	x            uint16
	y            uint8
	tile         uint8
	table2       bool // false: OAM1, true: OAM2
	palID        uint8
	prio         uint8
	hflip, vflip bool
	large        bool
}

func (o *oam) reset() {
	// TODO
	o.size = objSizes[0]
	o.reload = 0x0
}

func (o *oam) write(val uint8) {
	addr := o.addr
	if addr >= 0x220 {
		addr &= 0x21F
	}
	o.addr = (o.addr + 1) & 0x3FF

	o.buf[addr] = val
	idx := int((addr >> 2) & 0x7F)

	// >= 512
	if addr > 511 {
		idx = (int(addr) - 512) * 4
		for i := 0; i < 4; i++ {
			obj := &o.objs[idx+i]
			obj.x = setBit(obj.x, 8, bit(val, 2*i))
			obj.large = bit(val, 1+(2*i))
		}
		return
	}

	// < 512
	memo := o.memo
	o.memo = val
	obj := &o.objs[idx]
	if addr&0b1 == 1 {
		if addr%4 == 3 {
			// Byte2(memo), 3(val)
			obj.tile = memo
			obj.table2 = bit(val, 0)
			obj.palID = (val >> 1) & 0b111
			obj.prio = (val >> 4) & 0b11
			obj.hflip, obj.vflip = bit(val, 6), bit(val, 7)
		} else {
			// Byte0(memo), 1(val)
			obj.x = (obj.x & 0x100) | uint16(memo)
			obj.y = val
		}
	}
}

func (o *oam) read() uint8 {
	addr := o.addr
	if addr >= 0x220 {
		addr &= 0x21F
	}
	o.addr = (o.addr + 1) & 0x3FF

	return o.buf[addr]
}

func (o *obj) String() string {
	size := "S"
	if o.large {
		size = "L"
	}
	return fmt.Sprintf("X,Y: %04d,%04d Size: %s, Tile: %02X", o.x, o.y, size, o.tile)
}
