package cartridge

import (
	"errors"
	"fmt"
)

type RomType int

const (
	Unknown RomType = 0x00_0000
	LoROM   RomType = 0x00_7FC0
	HiROM   RomType = 0x00_FFC0
	ExHiROM RomType = 0x40_FFC0
)

type mapping uint8

type Header struct {
	T RomType

	// FFC0..FFD4h
	title [21]uint8

	// FFD5h
	mapping mapping

	// FFD6h
	Chipset *Chipset

	// FFD7h, FFD8h
	romSize, ramSize uint8

	// FFD9h
	destination destination

	// FFDAh
	maker uint8

	// FFDBh
	version uint8

	// FFDC..FFDDh
	checksumc uint16

	// FFDE..FFDFh
	checksum uint16
}

func NewHeader(romData []uint8) *Header {
	h := &Header{}
	h.load(romData)
	return h
}

func detectROMType(romData []uint8) RomType {
	switch {
	case isValidRomHeader(romData, LoROM):
		return LoROM
	case isValidRomHeader(romData, HiROM):
		return HiROM
	case isValidRomHeader(romData, ExHiROM):
		return ExHiROM

	// TODO
	case len(romData) == 1572864:
		return LoROM

	// TODO: For test ROM
	case len(romData) <= 65536:
		return LoROM
	}

	return Unknown
}

func isValidRomHeader(romData []uint8, t RomType) bool {
	romSize := len(romData)
	ofs := int(t)
	if romSize&0x3FF == 0x200 {
		ofs += 0x200
	}

	if ofs >= romSize {
		return false
	}

	hdr := romData[ofs : ofs+32]

	// compare checksum
	expected := uint16(hdr[31])<<8 | uint16(hdr[30])

	// calculate checksum
	checksum := uint16(0)
	for i := range romData {
		switch i {
		case ofs + 28, ofs + 29:
			checksum += 0xFF
		case ofs + 30, ofs + 31:
		default:
			checksum += uint16(romData[i])
		}
	}

	if expected == checksum {
		return true
	}

	size := int(hdr[0x17])
	if size == 8 && romSize == (1024*1024) {
		return true
	}
	return (size << 15) == romSize
}

func (h *Header) load(romData []uint8) {
	h.T = detectROMType(romData)
	if h.T == Unknown {
		panic(fmt.Sprintf("invalid ROM format (size: %s)", formatSize(uint(len(romData)))))
	}

	ofs := int(h.T)
	if len(romData)&0x3FF == 0x200 {
		ofs += 0x200
	}

	romHeader := romData[ofs : ofs+32]
	title := romHeader[0:21]

	if err := isSupportedCartridge(romHeader); err != nil {
		panic(err)
	}

	copy(h.title[:], title)
	h.mapping = mapping(romHeader[0x15])
	h.Chipset = chipset(romHeader[0x16])
	h.romSize = romHeader[0x17]
	h.ramSize = romHeader[0x18]
	h.destination = destination(romHeader[0x19])
	h.maker = romHeader[0x1a]
	h.version = romHeader[0x1b]
	h.checksumc = uint16(romHeader[0x1d])<<8 | uint16(romHeader[0x1c])
	h.checksum = uint16(romHeader[0x1f])<<8 | uint16(romHeader[0x1e])
}

func (m mapping) String() string {
	result := ""
	switch m & 0b1111 {
	case 0x0:
		result = "LoROM (Mode20)"
	case 0x1:
		result = "HiROM (Mode21)"
	case 0x2:
		result = "LoROM + S-DD1 (Mode22)"
	case 0x3:
		result = "LoROM + SA-1 (Mode23)"
	case 0x5:
		result = "ExHiROM (Mode25)"
	case 0xA:
		result = "HiROM + SPC7110 (Mode25)"
	default:
		result = "Unknown"
	}
	return result
}

func (h *Header) String() string {
	title := string(h.title[:])
	romSize := formatSize((1 << h.romSize) * 1024)
	ramSize := formatSize((1 << h.ramSize) * 1024)

	ok := "OK"
	if h.checksum != ^h.checksumc {
		ok = "NG"
	}

	return fmt.Sprintf(`Title: %s
  ROM Size:     %s
  RAM Size:     %s
  Mapping:      %s
  Chipset:      %v
  Destination:  %s
  Maker:        %d
  Version:      v1.%d
  Checksum:     0x%04X(%s)`, title, romSize, ramSize, h.mapping, h.Chipset, h.destination, h.maker, h.version, h.checksum, ok)
}

func isSupportedCartridge(hdr []uint8) error {
	chipset := hdr[0x16]
	subChipset := hdr[0x18]
	switch chipset {
	case 0x35:
		return errors.New("SA-1")
	case 0xF3:
		return errors.New("CX4")
	case 0xF5:
		switch subChipset {
		case 0x00:
			return errors.New("SPC7110")
		case 0x02:
			return errors.New("ST018")
		}
	}

	return nil
}
