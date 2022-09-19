package core

import "fmt"

var (
	READ_PC = u24(0, 0x2000)
)

type uint24 struct {
	bank   uint8
	offset uint16
}

func u24(bank uint8, offset uint16) uint24 {
	return uint24{bank, offset}
}

func toU24(addr uint32) uint24 {
	return uint24{uint8(addr >> 16), uint16(addr)}
}

func (b uint24) u32() uint32 {
	return uint32(b.bank)<<16 | uint32(b.offset)
}

func (b uint24) plus(i int) uint24 {
	val := uint32(int(b.u32()) + i)
	b = toU24(val)
	return b
}

func (b uint24) String() string {
	return fmt.Sprintf("%02X:%04X", b.bank, b.offset)
}
