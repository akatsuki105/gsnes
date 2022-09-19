package apu

type Timer struct {
	cycles  byte
	divider byte
	target  byte
	counter byte
	enabled bool
}

type APU interface {
	Reset()
	Cycle()
	// Read 1 byte from APU. idx is 0,1,2,3.
	Read(port int) byte
	// Write 1 byte into APU. idx is 0,1,2,3.
	Write(port int, val byte)
}

type apu struct {
	spc           *SPC
	dsp           *DSP
	ram           [0x10000]byte
	romReadable   bool
	dspAddr       byte
	cycles        uint32
	inPorts       [6]byte
	outPorts      [4]byte
	timer         [3]Timer
	cpuCyclesLeft byte
}

var bootRom [0x40]byte = [0x40]byte{
	0xcd, 0xef, 0xbd, 0xe8, 0x00, 0xc6, 0x1d, 0xd0, 0xfc, 0x8f, 0xaa, 0xf4, 0x8f, 0xbb, 0xf5, 0x78,
	0xcc, 0xf4, 0xd0, 0xfb, 0x2f, 0x19, 0xeb, 0xf4, 0xd0, 0xfc, 0x7e, 0xf4, 0xd0, 0x0b, 0xe4, 0xf5,
	0xcb, 0xf4, 0xd7, 0x00, 0xfc, 0xd0, 0xf3, 0xab, 0x01, 0x10, 0xef, 0x7e, 0xf4, 0x10, 0xeb, 0xba,
	0xf6, 0xda, 0x00, 0xba, 0xf4, 0xc4, 0xf4, 0xdd, 0x5d, 0xd0, 0xdb, 0x1f, 0x00, 0x00, 0xc0, 0xff,
}

func New() APU {
	a := &apu{}
	a.spc = NewSPC(a)
	a.dsp = NewDSP(a)
	return a
}

func (apu *apu) Reset() {
	apu.romReadable = true // before resetting spc, because it reads reset vector from it
	apu.spc.Reset()
	apu.dsp.Reset()
	apu.cpuCyclesLeft = 7
}

func (apu *apu) Read(port int) byte {
	return apu.outPorts[port]
}

func (apu *apu) Write(port int, val byte) {
	apu.inPorts[port] = val
}

func (apu *apu) Cycle() {
	if apu.cpuCyclesLeft == 0 {
		apu.cpuCyclesLeft = byte(apu.spc.runOpcode())
	}
	apu.cpuCyclesLeft--

	if (apu.cycles & 0x1F) == 0 {
		// every 32 cycles
		apu.dsp.Cycle()
	}

	// handle timers
	for i := 0; i < len(apu.timer); i++ {
		if apu.timer[i].cycles == 0 {
			if i == 2 {
				apu.timer[i].cycles = 16
			} else {
				apu.timer[i].cycles = 128
			}

			if apu.timer[i].enabled {
				apu.timer[i].divider++

				if apu.timer[i].divider == apu.timer[i].target {
					apu.timer[i].divider = 0
					apu.timer[i].counter++
					apu.timer[i].counter &= 0xF
				}
			}
		}
		apu.timer[i].cycles--
	}

	apu.cycles++
}

func (apu *apu) read(addr uint16) byte {
	switch addr {
	case 0xf0, 0xf1, 0xfa, 0xfb, 0xfc:
		return 0
	case 0xf2:
		return apu.dspAddr
	case 0xf3:
		return apu.dsp.Read(apu.dspAddr & 0x7f)
	case 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9:
		return apu.inPorts[addr-0xf4]
	case 0xfd, 0xfe, 0xff:
		var ret byte = apu.timer[addr-0xfd].counter
		apu.timer[addr-0xfd].counter = 0
		return ret
	}

	if apu.romReadable && addr >= 0xffc0 {
		return bootRom[addr-0xffc0]
	}

	return apu.ram[addr]
}

func (apu *apu) write(addr uint16, value byte) {
	switch addr {
	case 0xf0:
		break // test register
	case 0xf1:
		for i := 0; i < len(apu.timer); i++ {
			nextEnabled := (value & (1 << i)) > 0
			if !apu.timer[i].enabled && nextEnabled {
				apu.timer[i].divider = 0
				apu.timer[i].counter = 0
			}
			apu.timer[i].enabled = nextEnabled
		}
		if (value & 0x10) > 0 {
			apu.inPorts[0] = 0
			apu.inPorts[1] = 0
		}
		if (value & 0x20) > 0 {
			apu.inPorts[2] = 0
			apu.inPorts[3] = 0
		}
		apu.romReadable = (value & 0x80) > 0
	case 0xf2:
		apu.dspAddr = value
	case 0xf3:
		if apu.dspAddr < 0x80 {
			apu.dsp.Write(apu.dspAddr, value)
		}
	case 0xf4, 0xf5, 0xf6, 0xf7:
		apu.outPorts[addr-0xf4] = value
	case 0xf8, 0xf9:
		apu.inPorts[addr-0xf4] = value
	case 0xfa, 0xfb, 0xfc:
		apu.timer[addr-0xfa].target = value
	}

	apu.ram[addr] = value
}
