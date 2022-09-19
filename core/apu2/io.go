package apu

type io8 interface {
	Read() uint8
	Write(val uint8)
}
