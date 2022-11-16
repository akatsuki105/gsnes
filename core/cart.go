package core

import (
	"fmt"

	cart "github.com/pokemium/gsnes/core/cartridge"
)

type cartridge struct {
	c    *sfc
	h    cart.Header
	rom  []uint8
	sram []uint8
}

func newCartridge(c *sfc) *cartridge {
	return &cartridge{
		c:    c,
		sram: make([]uint8, 32*KB),
	}
}

func (c *cartridge) loadROM(romData []uint8) {
	c.h = *cart.NewHeader(romData)
	c.rom = romData
	s := c.c

	switch c.h.T {
	case cart.LoROM:
		s.m.mmap(memblock("00-7D,80-FF:8000-FFFF", c.read, c.write).mask(0x3F_7FFF))

		// System mirror
		s.m.mmap(memblock("40-7D,C0-FF:0000-1FFF", s.w.wram.read, s.w.wram.write).mask(0x1FFF))    // WRAM(mirror)
		s.m.mmap(memblock("40-7D,C0-FF:2100-213F", s.ppu.readIO, s.ppu.writeIO).mask(0x3F))        // PPU
		s.m.mmap(memblock("40-7D,C0-FF:2140-217F", s.apu.readIO, s.apu.writeIO).mask(0x3))         // APU
		s.m.mmap(memblock("40-7D,C0-FF:2180-2183,4016-4017,4200-421F", s.w.readCPU, s.w.writeCPU)) // CPU
		s.m.mmap(memblock("40-7D,C0-FF:4300-437F", s.dma.readIO, s.dma.writeIO).mask(0x7F))        // DMA
		if cart.HaveSRAM(&c.h) {
			s.m.mmap(memblock("70-7D,F0-FF:0000-7FFF", c.readSRAM, c.writeSRAM).mask(0x7FFF)) // SRAM
		}

	case cart.HiROM:
		s.m.mmap(memblock("40-7D,C0-FF:0000-FFFF", c.read, c.write).mask(0x3F_FFFF))
		s.m.mmap(memblock("00-3F,80-BF:8000-FFFF", c.read, c.write).mask(0x3F_FFFF))
		if cart.HaveSRAM(&c.h) {
			s.m.mmap(memblock("30-3F,B0-BF:6000-7FFF", c.readSRAM, c.writeSRAM).mask(0x1FFF))
		}

	case cart.ExHiROM:
		crash("ExHiROM is not implemented")
	}
}

// addr is 00-3F:8000-FFFF
func (c *cartridge) read(addr uint, defaultVal uint8) uint8 {
	bank, offset := addr>>16, (addr & 0xFFFF)

	idx := 0
	switch c.h.T {
	case cart.LoROM:
		idx = int(32*KB*bank + offset)
		if idx > len(c.rom) {
			idx &= len(c.rom) - 1
		}

	case cart.HiROM:
		idx = int(64*KB*bank + offset)
	case cart.ExHiROM:
		crash("ExHiROM is not implemented")
	}

	if idx >= len(c.rom) {
		fmt.Printf("invalid address: %v(ROM: 0x%X, idx: 0x%X)\n", toU24(uint32(addr)), len(c.rom), idx)
	}
	return c.rom[idx]
}

// addr is 00-3F:8000-FFFF
func (c *cartridge) write(addr uint, val uint8) {
	// nop
}

func (c *cartridge) readSRAM(addr uint, _ uint8) uint8 {
	return c.sram[addr]
}

func (c *cartridge) writeSRAM(addr uint, val uint8) {
	c.sram[addr] = val
}

func PrintCartInfo(romData []uint8) {
	fmt.Println(cart.NewHeader(romData))
}
