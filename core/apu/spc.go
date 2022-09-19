package apu

type SPC struct {
	apu *apu
	// registers
	a  byte
	x  byte
	y  byte
	sp byte
	pc uint16
	// flags
	c byte // carry flag
	z byte // zero flag
	i byte // Interrupt flag
	h byte // HalfCarry flag
	b byte // Break flag
	p byte // ZerPageLocation flag
	v byte // Overflow flag
	n byte // Negative flag
	// stopping
	stopped bool
	// internal use
	cyclesUsed byte // indicates how many cycles an opcode used
}

const (
	SPCFlagsCarry            = 0x01
	SPCFlagsZero             = 0x02
	SPCFlagsInterrupt        = 0x04
	SPCFlagsHalfCarry        = 0x08
	SPCFlagsBreak            = 0x10
	SPCFlagsZeroPageLocation = 0x20
	SPCFlagsOverflow         = 0x40
	SPCFlagsNegative         = 0x80
)

var cyclesPerSPCOpcode [256]int = [256]int{
	2, 8, 4, 5, 3, 4, 3, 6, 2, 6, 5, 4, 5, 4, 6, 8,
	2, 8, 4, 5, 4, 5, 5, 6, 5, 5, 6, 5, 2, 2, 4, 6,
	2, 8, 4, 5, 3, 4, 3, 6, 2, 6, 5, 4, 5, 4, 5, 4,
	2, 8, 4, 5, 4, 5, 5, 6, 5, 5, 6, 5, 2, 2, 3, 8,
	2, 8, 4, 5, 3, 4, 3, 6, 2, 6, 4, 4, 5, 4, 6, 6,
	2, 8, 4, 5, 4, 5, 5, 6, 5, 5, 4, 5, 2, 2, 4, 3,
	2, 8, 4, 5, 3, 4, 3, 6, 2, 6, 4, 4, 5, 4, 5, 5,
	2, 8, 4, 5, 4, 5, 5, 6, 5, 5, 5, 5, 2, 2, 3, 6,
	2, 8, 4, 5, 3, 4, 3, 6, 2, 6, 5, 4, 5, 2, 4, 5,
	2, 8, 4, 5, 4, 5, 5, 6, 5, 5, 5, 5, 2, 2, 12, 5,
	2, 8, 4, 5, 3, 4, 3, 6, 2, 6, 4, 4, 5, 2, 4, 4,
	2, 8, 4, 5, 4, 5, 5, 6, 5, 5, 5, 5, 2, 2, 3, 4,
	2, 8, 4, 5, 4, 5, 4, 7, 2, 5, 6, 4, 5, 2, 4, 9,
	2, 8, 4, 5, 5, 6, 6, 7, 4, 5, 5, 5, 2, 2, 6, 3,
	2, 8, 4, 5, 3, 4, 3, 6, 2, 4, 5, 3, 4, 3, 4, 3,
	2, 8, 4, 5, 4, 5, 5, 6, 3, 4, 5, 4, 2, 2, 4, 3,
}

func NewSPC(apu *apu) *SPC {
	return &SPC{
		apu: apu,
	}
}

func (spc *SPC) Read(addr uint16) byte {
	return spc.apu.read(addr)
}

func (spc *SPC) Write(addr uint16, value byte) {
	spc.apu.write(addr, value)
}

func (spc *SPC) Reset() {
	spc.pc = uint16(spc.Read(0xfffe)) | (uint16(spc.Read(0xffff)) << 8)
}

func (spc *SPC) runOpcode() int {
	spc.cyclesUsed = 0
	if spc.stopped {
		return 1
	}
	var opcode byte = spc.readOpcode()
	spc.cyclesUsed = byte(cyclesPerSPCOpcode[opcode])
	spc.doOpcode(opcode)
	return int(spc.cyclesUsed)
}

func (spc *SPC) readOpcode() byte {
	opcode := spc.Read(spc.pc)
	spc.pc++
	return opcode
}

func (spc *SPC) readOpcodeWord() uint16 {
	var low byte = spc.readOpcode()
	return uint16(low) | (uint16(spc.readOpcode()) << 8)
}

func (spc *SPC) SetAllFlags(flags byte) {
	spc.c = (flags >> 0) & 1
	spc.z = (flags >> 1) & 1
	spc.i = (flags >> 2) & 1
	spc.h = (flags >> 3) & 1
	spc.b = (flags >> 4) & 1
	spc.p = (flags >> 5) & 1
	spc.v = (flags >> 6) & 1
	spc.n = (flags >> 7) & 1
}

func (spc *SPC) SetFlags(flags byte) {
	spc.SetAllFlags(spc.Flags() | flags)
}

func (spc *SPC) ClearFlags(flags byte) {
	spc.SetAllFlags(spc.Flags() & ^flags)
}

func (spc *SPC) Flags() byte {
	var flags byte
	flags |= spc.c << 0
	flags |= spc.z << 1
	flags |= spc.i << 2
	flags |= spc.h << 3
	flags |= spc.b << 4
	flags |= spc.p << 5
	flags |= spc.v << 6
	flags |= spc.n << 7

	return flags
}

func (spc *SPC) CheckFlag(flag byte) bool {
	return (spc.Flags() & flag) == flag
}

func (spc *SPC) setZN(value byte) {
	if value == 0 {
		spc.z = 1
	} else {
		spc.z = 0
	}
	if (value & 0x80) > 0 {
		spc.n = 1
	} else {
		spc.n = 0
	}
}

func (spc *SPC) branch(value byte, check bool) {
	if check {
		spc.cyclesUsed += 2 // taken branch: 2 extra cycles
		spc.pc += uint16(int8(value))
	}
}

func (spc *SPC) pullByte() byte {
	spc.sp++
	return spc.Read(0x0100 | uint16(spc.sp))
}

func (spc *SPC) pushByte(value byte) {
	spc.Write(0x0100|uint16(spc.sp), value)
	spc.sp--
}

func (spc *SPC) pullWord() uint16 {
	var value byte = spc.pullByte()
	return uint16(value) | (uint16(spc.pullByte()) << 8)
}

func (spc *SPC) pushWord(value uint16) {
	spc.pushByte(byte(uint16(value) >> 8))
	spc.pushByte(byte(value & 0xFF))
}

func (spc *SPC) readWord(addrLow uint16, addrHigh uint16) uint16 {
	var value byte = spc.Read(addrLow)
	return uint16(value) | (uint16(spc.Read(addrHigh)) << 8)
}

func (spc *SPC) writeWord(addrLow uint16, addrHigh uint16, value uint16) {
	spc.Write(addrLow, byte(value&0xFF))
	spc.Write(addrHigh, byte(uint16(value)>>8))
}

// addressing modes

func (spc *SPC) addrDp() uint16 {
	return uint16(spc.readOpcode()) | (uint16(spc.p) << 8)
}

func (spc *SPC) addrAbs() uint16 {
	return spc.readOpcodeWord()
}

func (spc *SPC) addrInd() uint16 {
	return uint16(spc.x) | (uint16(spc.p) << 8)
}

func (spc *SPC) addrIdx() uint16 {
	var pointer byte = spc.readOpcode()
	lowBase := uint16(pointer) + uint16(spc.x)
	high := uint16(spc.p) << 8

	return spc.readWord((lowBase&0xff)|(high), ((lowBase+1)&0xff)|(high))
}

func (spc *SPC) addrImm() uint16 {
	v := spc.pc
	spc.pc++
	return v
}

func (spc *SPC) addrDpx() uint16 {
	low := (uint16(spc.readOpcode()) + uint16(spc.x)) & 0xff
	high := uint16(spc.p) << 8
	return low | high
}

func (spc *SPC) addrDpy() uint16 {
	low := (uint16(spc.readOpcode()) + uint16(spc.y)) & 0xff
	high := uint16(spc.p) << 8
	return low | high
}

func (spc *SPC) addrAbx() uint16 {
	return (spc.readOpcodeWord() + uint16(spc.x)) & 0xffff
}

func (spc *SPC) addrAby() uint16 {
	return (spc.readOpcodeWord() + uint16(spc.y)) & 0xffff
}

func (spc *SPC) addrIdy() uint16 {
	var pointer byte = spc.readOpcode()
	var addr uint16 = spc.readWord(uint16(pointer)|(uint16(spc.p)<<8), ((uint16(pointer)+1)&0xff)|(uint16(spc.p)<<8))
	return (addr + uint16(spc.y)) & 0xffff
}

func (spc *SPC) addrDpDp(src *uint16) uint16 {
	*src = uint16(spc.readOpcode()) | (uint16(spc.p) << 8)
	return uint16(spc.readOpcode()) | (uint16(spc.p) << 8)
}

func (spc *SPC) addrDpImm(src *uint16) uint16 {
	*src = spc.pc
	spc.pc++
	return uint16(spc.readOpcode()) | (uint16(spc.p) << 8)
}

func (spc *SPC) addrIndInd(src *uint16) uint16 {
	*src = uint16(spc.y) | (uint16(spc.p) << 8)
	return uint16(spc.x) | (uint16(spc.p) << 8)
}

func (spc *SPC) addrAbsBit(addr *uint16) byte {
	var addrBit uint16 = spc.readOpcodeWord()
	*addr = addrBit & 0x1fff
	return byte(addrBit >> 13)
}

func (spc *SPC) addrDpWord(low *uint16) uint16 {
	var addr byte = spc.readOpcode()
	*low = uint16(addr) | (uint16(spc.p) << 8)
	return ((uint16(addr) + 1) & 0xff) | (uint16(spc.p) << 8)
}

func (spc *SPC) addrIndP() uint16 {
	v := spc.x
	spc.x++
	return uint16(v) | (uint16(spc.p) << 8)
}
