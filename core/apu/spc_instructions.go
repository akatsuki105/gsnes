package apu

// opcode functions

func (spc *SPC) doOpcode(opcode byte) {
	switch opcode {
	case 0x00:
		// nop imp
		// no operation
	case 0x01, 0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1, 0xf1:
		// tcall imp
		spc.pushWord(spc.pc)
		var addr uint16 = 0xffde - (2 * (uint16(opcode) >> 4))
		spc.pc = spc.readWord(addr, addr+1)
	case 0x02, 0x22, 0x42, 0x62, 0x82, 0xa2, 0xc2, 0xe2:
		// set1 dp
		var addr uint16 = spc.addrDp()
		spc.Write(addr, spc.Read(addr)|(1<<(opcode>>5)))
	case 0x12, 0x32, 0x52, 0x72, 0x92, 0xb2, 0xd2, 0xf2:
		// clr1 dp
		var addr uint16 = spc.addrDp()
		spc.Write(addr, spc.Read(addr) & ^(1<<(opcode>>5)))
	case 0x03, 0x23, 0x43, 0x63, 0x83, 0xa3, 0xc3, 0xe3:
		// bbs dp, rel
		var val byte = spc.Read(spc.addrDp())
		check := (val & (1 << (opcode >> 5))) > 0
		spc.branch(spc.readOpcode(), check)
	case 0x13, 0x33, 0x53, 0x73, 0x93, 0xb3, 0xd3, 0xf3:
		// bbc dp, rel
		var val byte = spc.Read(spc.addrDp())
		check := (val & (1 << (opcode >> 5))) == 0
		spc.branch(spc.readOpcode(), check)
	case 0x04:
		// or  dp
		spc.or(spc.addrDp())
	case 0x05:
		// or  abs
		spc.or(spc.addrAbs())
	case 0x06:
		// or  ind
		spc.or(spc.addrInd())
	case 0x07:
		// or  idx
		spc.or(spc.addrIdx())
	case 0x08:
		// or  imm
		spc.or(spc.addrImm())
	case 0x09:
		// orm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		spc.orm(dst, src)
	case 0x0a:
		// or1 abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		if (spc.c | ((spc.Read(addr) >> bit) & 1)) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0x0b:
		// asl dp
		spc.asl(spc.addrDp())
	case 0x0c:
		// asl abs
		spc.asl(spc.addrAbs())
	case 0x0d:
		// pushp imp
		spc.pushByte(spc.Flags())
	case 0x0e:
		// tset1 abs
		var addr uint16 = spc.addrAbs()
		var val byte = spc.Read(addr)
		var result byte = spc.a + (val ^ 0xff) + 1
		spc.setZN(result)
		spc.Write(addr, val|spc.a)
	case 0x0f:
		// brk imp
		spc.pushWord(spc.pc)
		spc.pushByte(spc.Flags())
		spc.i = 0x00
		spc.b = 0x01
		spc.pc = spc.readWord(0xffde, 0xffdf)
	case 0x10:
		// bpl rel
		spc.branch(spc.readOpcode(), !spc.CheckFlag(SPCFlagsNegative))
	case 0x14:
		// or  dpx
		spc.or(spc.addrDpx())
	case 0x15:
		// or  abx
		spc.or(spc.addrAbx())
	case 0x16:
		// or  aby
		spc.or(spc.addrAby())
	case 0x17:
		// or  idy
		spc.or(spc.addrIdy())
	case 0x18:
		// orm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		spc.orm(dst, src)
	case 0x19:
		// orm ind, ind
		var src uint16 = 0
		var dst uint16 = spc.addrIndInd(&src)
		spc.orm(dst, src)
	case 0x1a:
		// decw dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		var value uint16 = spc.readWord(low, high) - 1
		if value == 0 {
			spc.z = 0x01
		} else {
			spc.z = 0x00
		}
		if (value & 0x8000) > 0 {
			spc.n = 0x01
		} else {
			spc.n = 0x00
		}
		spc.writeWord(low, high, value)
	case 0x1b:
		// asl dpx
		spc.asl(spc.addrDpx())
	case 0x1c:
		// asla imp
		if (spc.a & 0x80) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		spc.a <<= 1
		spc.setZN(spc.a)
	case 0x1d:
		// decx imp
		spc.x--
		spc.setZN(spc.x)
	case 0x1e:
		// cmpx abs
		spc.cmpx(spc.addrAbs())
	case 0x1f:
		// jmp iax
		var pointer uint16 = spc.readOpcodeWord()
		base := pointer + uint16(spc.x)
		spc.pc = spc.readWord(base&0xffff, (base+1)&0xffff)
	case 0x20:
		// clrp imp
		spc.p = 0x00
	case 0x24:
		// and dp
		spc.and(spc.addrDp())
	case 0x25:
		// and abs
		spc.and(spc.addrAbs())
	case 0x26:
		// and ind
		spc.and(spc.addrInd())
	case 0x27:
		// and idx
		spc.and(spc.addrIdx())
	case 0x28:
		// and imm
		spc.and(spc.addrImm())
	case 0x29:
		// andm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		spc.andm(dst, src)
	case 0x2a:
		// or1n abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		if (spc.c | (^(spc.Read(addr) >> bit) & 1)) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0x2b:
		// rol dp
		spc.rol(spc.addrDp())
	case 0x2c:
		// rol abs
		spc.rol(spc.addrAbs())
	case 0x2d:
		// pusha imp
		spc.pushByte(spc.a)
	case 0x2e:
		// cbne dp, rel
		var val byte = spc.Read(spc.addrDp()) ^ 0xff
		var result byte = spc.a + val + 1
		spc.branch(spc.readOpcode(), result != 0)
	case 0x2f:
		// bra rel
		spc.pc += uint16(int8(spc.readOpcode()))
	case 0x30:
		// bmi rel
		spc.branch(spc.readOpcode(), spc.CheckFlag(SPCFlagsNegative))
	case 0x34:
		// and dpx
		spc.and(spc.addrDpx())
	case 0x35:
		// and abx
		spc.and(spc.addrAbx())
	case 0x36:
		// and aby
		spc.and(spc.addrAby())
	case 0x37:
		// and idy
		spc.and(spc.addrIdy())
	case 0x38:
		// andm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		spc.andm(dst, src)
	case 0x39:
		// andm ind, ind
		var src uint16 = 0
		var dst uint16 = spc.addrIndInd(&src)
		spc.andm(dst, src)
	case 0x3a:
		// incw dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		var value uint16 = spc.readWord(low, high) + 1
		if value == 0 {
			spc.z = 0x01
		} else {
			spc.z = 0x00
		}
		if (value & 0x8000) > 0 {
			spc.n = 0x01
		} else {
			spc.n = 0x00
		}
		spc.writeWord(low, high, value)
	case 0x3b:
		// rol dpx
		spc.rol(spc.addrDpx())
	case 0x3c:
		// rola imp
		var newC bool = (spc.a & 0x80) > 0
		spc.a = (spc.a << 1) | spc.c
		if newC {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		spc.setZN(spc.a)
	case 0x3d:
		// incx imp
		spc.x++
		spc.setZN(spc.x)
	case 0x3e:
		// cmpx dp
		spc.cmpx(spc.addrDp())
	case 0x3f:
		// call abs
		var dst uint16 = spc.readOpcodeWord()
		spc.pushWord(spc.pc)
		spc.pc = dst
	case 0x40:
		// setp imp
		spc.p = 0x01
	case 0x44:
		// eor dp
		spc.eor(spc.addrDp())
	case 0x45:
		// eor abs
		spc.eor(spc.addrAbs())
	case 0x46:
		// eor ind
		spc.eor(spc.addrInd())
	case 0x47:
		// eor idx
		spc.eor(spc.addrIdx())
	case 0x48:
		// eor imm
		spc.eor(spc.addrImm())
	case 0x49:
		// eorm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		spc.eorm(dst, src)
	case 0x4a:
		// and1 abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		if (spc.c & ((spc.Read(addr) >> bit) & 1)) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0x4b:
		// lsr dp
		spc.lsr(spc.addrDp())
	case 0x4c:
		// lsr abs
		spc.lsr(spc.addrAbs())
	case 0x4d:
		// pushx imp
		spc.pushByte(spc.x)
	case 0x4e:
		// tclr1 abs
		var addr uint16 = spc.addrAbs()
		var val byte = spc.Read(addr)
		var result byte = spc.a + (val ^ 0xff) + 1
		spc.setZN(result)
		spc.Write(addr, val & ^spc.a)
	case 0x4f:
		// pcall dp
		var dst byte = spc.readOpcode()
		spc.pushWord(spc.pc)
		spc.pc = uint16(0xff00) | uint16(dst)
	case 0x50:
		// bvc rel
		spc.branch(spc.readOpcode(), !spc.CheckFlag(SPCFlagsOverflow))
	case 0x54:
		// eor dpx
		spc.eor(spc.addrDpx())
	case 0x55:
		// eor abx
		spc.eor(spc.addrAbx())
	case 0x56:
		// eor aby
		spc.eor(spc.addrAby())
	case 0x57:
		// eor idy
		spc.eor(spc.addrIdy())
	case 0x58:
		// eorm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		spc.eorm(dst, src)
	case 0x59:
		// eorm ind, ind
		var src uint16 = 0
		var dst uint16 = spc.addrIndInd(&src)
		spc.eorm(dst, src)
	case 0x5a:
		// cmpw dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		var value uint16 = spc.readWord(low, high) ^ 0xffff
		var ya uint16 = uint16(spc.a) | (uint16(spc.y) << 8)
		var result int = int(ya) + int(value) + 1
		if result > 0xffff {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		if (result & 0xffff) == 0 {
			spc.z = 0x01
		} else {
			spc.z = 0x00
		}
		if (result & 0x8000) > 0 {
			spc.n = 0x01
		} else {
			spc.n = 0x00
		}
	case 0x5b:
		// lsr dpx
		spc.lsr(spc.addrDpx())
	case 0x5c:
		// lsra imp
		if (spc.a & 1) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		spc.a >>= 1
		spc.setZN(spc.a)
	case 0x5d:
		// movxa imp
		spc.x = spc.a
		spc.setZN(spc.x)
	case 0x5e:
		// cmpy abs
		spc.cmpy(spc.addrAbs())
	case 0x5f:
		// jmp abs
		spc.pc = spc.readOpcodeWord()
	case 0x60:
		// clrc imp
		spc.c = 0x00
	case 0x64:
		// cmp dp
		spc.cmp(spc.addrDp())
	case 0x65:
		// cmp abs
		spc.cmp(spc.addrAbs())
	case 0x66:
		// cmp ind
		spc.cmp(spc.addrInd())
	case 0x67:
		// cmp idx
		spc.cmp(spc.addrIdx())
	case 0x68:
		// cmp imm
		spc.cmp(spc.addrImm())
	case 0x69:
		// cmpm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		spc.cmpm(dst, src)
	case 0x6a:
		// and1n abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		if (spc.c & (^(spc.Read(addr) >> bit) & 1)) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0x6b:
		// ror dp
		spc.ror(spc.addrDp())
	case 0x6c:
		// ror abs
		spc.ror(spc.addrAbs())
	case 0x6d:
		// pushy imp
		spc.pushByte(spc.y)
	case 0x6e:
		// dbnz dp, rel
		var addr uint16 = spc.addrDp()
		var result byte = spc.Read(addr) - 1
		spc.Write(addr, result)
		spc.branch(spc.readOpcode(), result != 0)
	case 0x6f:
		// ret imp
		spc.pc = spc.pullWord()
	case 0x70:
		// bvs rel
		spc.branch(spc.readOpcode(), spc.CheckFlag(SPCFlagsOverflow))
	case 0x74:
		// cmp dpx
		spc.cmp(spc.addrDpx())
	case 0x75:
		// cmp abx
		spc.cmp(spc.addrAbx())
	case 0x76:
		// cmp aby
		spc.cmp(spc.addrAby())
	case 0x77:
		// cmp idy
		spc.cmp(spc.addrIdy())
	case 0x78:
		// cmpm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		spc.cmpm(dst, src)
	case 0x79:
		// cmpm ind, ind
		var src uint16 = 0
		var dst uint16 = spc.addrIndInd(&src)
		spc.cmpm(dst, src)
	case 0x7a:
		// addw dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		var value uint16 = spc.readWord(low, high)
		var ya uint16 = uint16(spc.a) | (uint16(spc.y) << 8)
		var result int = int(ya) + int(value)
		if (ya&0x8000) == (value&0x8000) && (value&0x8000) != (uint16(result)&0x8000) {
			spc.v = 0x01
		} else {
			spc.v = 0x00
		}
		if ((ya & 0xfff) + (value & 0xfff) + 1) > 0xfff {
			spc.h = 0x01
		} else {
			spc.h = 0x00
		}
		if result > 0xffff {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		if (result & 0xffff) == 0 {
			spc.z = 0x01
		} else {
			spc.z = 0x00
		}
		if (result & 0x8000) > 0 {
			spc.n = 0x01
		} else {
			spc.n = 0x00
		}
		spc.a = byte(result & 0xff)
		spc.y = byte(result >> 8)
	case 0x7b:
		// ror dpx
		spc.ror(spc.addrDpx())
	case 0x7c:
		// rora imp
		var newC bool = (spc.a & 1) > 0
		spc.a = (spc.a >> 1) | (spc.c << 7)
		if newC {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		spc.setZN(spc.a)
	case 0x7d:
		// movax imp
		spc.a = spc.x
		spc.setZN(spc.a)
	case 0x7e:
		// cmpy dp
		spc.cmpy(spc.addrDp())
	case 0x7f:
		// reti imp
		spc.SetAllFlags(spc.pullByte())
		spc.pc = spc.pullWord()
	case 0x80:
		// setc imp
		spc.c = 0x01
	case 0x84:
		// adc dp
		spc.adc(spc.addrDp())
	case 0x85:
		// adc abs
		spc.adc(spc.addrAbs())
	case 0x86:
		// adc ind
		spc.adc(spc.addrInd())
	case 0x87:
		// adc idx
		spc.adc(spc.addrIdx())
	case 0x88:
		// adc imm
		spc.adc(spc.addrImm())
	case 0x89:
		// adcm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		spc.adcm(dst, src)
	case 0x8a:
		// eor1 abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		if (spc.c ^ ((spc.Read(addr) >> bit) & 1)) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0x8b:
		// dec dp
		spc.dec(spc.addrDp())
	case 0x8c:
		// dec abs
		spc.dec(spc.addrAbs())
	case 0x8d:
		// movy imm
		spc.movy(spc.addrImm())
	case 0x8e:
		// popp imp
		spc.SetAllFlags(spc.pullByte())
	case 0x8f:
		// movm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		var val byte = spc.Read(src)
		spc.Read(dst)
		spc.Write(dst, val)
	case 0x90:
		// bcc rel
		spc.branch(spc.readOpcode(), !spc.CheckFlag(SPCFlagsCarry))
	case 0x94:
		// adc dpx
		spc.adc(spc.addrDpx())
	case 0x95:
		// adc abx
		spc.adc(spc.addrAbx())
	case 0x96:
		// adc aby
		spc.adc(spc.addrAby())
	case 0x97:
		// adc idy
		spc.adc(spc.addrIdy())
	case 0x98:
		// adcm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		spc.adcm(dst, src)
	case 0x99:
		// adcm ind, ind
		var src uint16 = 0
		var dst uint16 = spc.addrIndInd(&src)
		spc.adcm(dst, src)
	case 0x9a:
		// subw dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		var value uint16 = spc.readWord(low, high) ^ 0xffff
		var ya uint16 = uint16(spc.a) | (uint16(spc.y) << 8)
		var result int = int(ya) + int(value) + 1
		if (ya&0x8000) == (value&0x8000) && (value&0x8000) != (uint16(result)&0x8000) {
			spc.v = 0x01
		} else {
			spc.v = 0x00
		}
		if ((ya & 0xfff) + (value & 0xfff) + 1) > 0xfff {
			spc.h = 0x01
		} else {
			spc.h = 0x00
		}
		if result > 0xffff {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
		if (result & 0xffff) == 0 {
			spc.z = 0x01
		} else {
			spc.z = 0x00
		}
		if (result & 0x8000) > 0 {
			spc.n = 0x01
		} else {
			spc.n = 0x00
		}
		spc.a = byte(result & 0xff)
		spc.y = byte(result >> 8)
	case 0x9b:
		// dec dpx
		spc.dec(spc.addrDpx())
	case 0x9c:
		// deca imp
		spc.a--
		spc.setZN(spc.a)
	case 0x9d:
		// movxp imp
		spc.x = spc.sp
		spc.setZN(spc.x)
	case 0x9e:
		// div imp
		// TODO: proper division algorithm
		var value uint16 = uint16(spc.a) | (uint16(spc.y) << 8)
		var result int = 0xffff
		var mod int = int(spc.a)
		if spc.x != 0 {
			result = int(value) / int(spc.x)
			mod = int(value) % int(spc.x)
		}
		if result > 0xff {
			spc.v = 0x01
		} else {
			spc.v = 0x00
		}
		if (spc.x & 0xf) <= (spc.y & 0xf) {
			spc.h = 0x01
		} else {
			spc.h = 0x00
		}
		spc.a = byte(result)
		spc.y = byte(mod)
		spc.setZN(spc.a)
	case 0x9f:
		// xcn imp
		spc.a = (spc.a >> 4) | (spc.a << 4)
		spc.setZN(spc.a)
	case 0xa0:
		// ei  imp
		spc.i = 0x01
	case 0xa4:
		// sbc dp
		spc.sbc(spc.addrDp())
	case 0xa5:
		// sbc abs
		spc.sbc(spc.addrAbs())
	case 0xa6:
		// sbc ind
		spc.sbc(spc.addrInd())
	case 0xa7:
		// sbc idx
		spc.sbc(spc.addrIdx())
	case 0xa8:
		// sbc imm
		spc.sbc(spc.addrImm())
	case 0xa9:
		// sbcm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		spc.sbcm(dst, src)
	case 0xaa:
		// mov1 abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		if ((spc.Read(addr) >> bit) & 1) > 0 {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0xab:
		// inc dp
		spc.inc(spc.addrDp())
	case 0xac:
		// inc abs
		spc.inc(spc.addrAbs())
	case 0xad:
		// cmpy imm
		spc.cmpy(spc.addrImm())
	case 0xae:
		// popa imp
		spc.a = spc.pullByte()
	case 0xaf:
		// movs ind+
		var addr uint16 = spc.addrIndP()
		spc.Write(addr, spc.a)
	case 0xb0:
		// bcs rel
		spc.branch(spc.readOpcode(), spc.CheckFlag(SPCFlagsCarry))
	case 0xb4:
		// sbc dpx
		spc.sbc(spc.addrDpx())
	case 0xb5:
		// sbc abx
		spc.sbc(spc.addrAbx())
	case 0xb6:
		// sbc aby
		spc.sbc(spc.addrAby())
	case 0xb7:
		// sbc idy
		spc.sbc(spc.addrIdy())
	case 0xb8:
		// sbcm dp, imm
		var src uint16 = 0
		var dst uint16 = spc.addrDpImm(&src)
		spc.sbcm(dst, src)
	case 0xb9:
		// sbcm ind, ind
		var src uint16 = 0
		var dst uint16 = spc.addrIndInd(&src)
		spc.sbcm(dst, src)
	case 0xba:
		// movw dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		var val uint16 = spc.readWord(low, high)
		spc.a = byte(val & 0xff)
		spc.y = byte(val >> 8)
		if val == 0 {
			spc.z = 0x01
		} else {
			spc.z = 0x00
		}
		if (val & 0x8000) > 0 {
			spc.n = 0x01
		} else {
			spc.n = 0x00
		}
	case 0xbb:
		// inc dpx
		spc.inc(spc.addrDpx())
	case 0xbc:
		// inca imp
		spc.a++
		spc.setZN(spc.a)
	case 0xbd:
		// movpx imp
		spc.sp = spc.x
	case 0xbe:
		// das imp
		if spc.a > 0x99 || !spc.CheckFlag(SPCFlagsCarry) {
			spc.a -= 0x60
			spc.c = 0x00
		}
		if (spc.a&0xf) > 9 || !spc.CheckFlag(SPCFlagsHalfCarry) {
			spc.a -= 6
		}
		spc.setZN(spc.a)
	case 0xbf:
		// mov ind+
		var addr uint16 = spc.addrIndP()
		spc.a = spc.Read(addr)
		spc.setZN(spc.a)
	case 0xc0:
		// di  imp
		spc.i = 0x00
	case 0xc4:
		// movs dp
		spc.movs(spc.addrDp())
	case 0xc5:
		// movs abs
		spc.movs(spc.addrAbs())
	case 0xc6:
		// movs ind
		spc.movs(spc.addrInd())
	case 0xc7:
		// movs idx
		spc.movs(spc.addrIdx())
	case 0xc8:
		// cmpx imm
		spc.cmpx(spc.addrImm())
	case 0xc9:
		// movsx abs
		spc.movsx(spc.addrAbs())
	case 0xca:
		// mov1s abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		var result byte = (spc.Read(addr) & (^(1 << bit))) | (spc.c << bit)
		spc.Write(addr, result)
	case 0xcb:
		// movsy dp
		spc.movsy(spc.addrDp())
	case 0xcc:
		// movsy abs
		spc.movsy(spc.addrAbs())
	case 0xcd:
		// movx imm
		spc.movx(spc.addrImm())
	case 0xce:
		// popx imp
		spc.x = spc.pullByte()
	case 0xcf:
		// mul imp
		var result uint16 = uint16(spc.a) * uint16(spc.y)
		spc.a = byte(result & 0xff)
		spc.y = byte(result >> 8)
		spc.setZN(spc.y)
	case 0xd0:
		// bne rel
		spc.branch(spc.readOpcode(), !spc.CheckFlag(SPCFlagsZero))
	case 0xd4:
		// movs dpx
		spc.movs(spc.addrDpx())
	case 0xd5:
		// movs abx
		spc.movs(spc.addrAbx())
	case 0xd6:
		// movs aby
		spc.movs(spc.addrAby())
	case 0xd7:
		// movs idy
		spc.movs(spc.addrIdy())
	case 0xd8:
		// movsx dp
		spc.movsx(spc.addrDp())
	case 0xd9:
		// movsx dpy
		spc.movsx(spc.addrDpy())
	case 0xda:
		// movws dp
		var low uint16 = 0
		var high uint16 = spc.addrDpWord(&low)
		spc.Read(low)
		spc.Write(low, spc.a)
		spc.Write(high, spc.y)
	case 0xdb:
		// movsy dpx
		spc.movsy(spc.addrDpx())
	case 0xdc:
		// decy imp
		spc.y--
		spc.setZN(spc.y)
	case 0xdd:
		// movay imp
		spc.a = spc.y
		spc.setZN(spc.a)
	case 0xde:
		// cbne dpx, rel
		var val byte = spc.Read(spc.addrDpx()) ^ 0xff
		var result byte = spc.a + val + 1
		spc.branch(spc.readOpcode(), result != 0)
	case 0xdf:
		// daa imp
		if spc.a > 0x99 || spc.CheckFlag(SPCFlagsCarry) {
			spc.a += 0x60
			spc.c = 0x01 // XXX: what?
		}
		if (spc.a&0xf) > 9 || spc.CheckFlag(SPCFlagsHalfCarry) {
			spc.a += 6
		}
		spc.setZN(spc.a)
	case 0xe0:
		// clrv imp
		spc.v = 0x00
		spc.h = 0x00
	case 0xe4:
		// mov dp
		spc.mov(spc.addrDp())
	case 0xe5:
		// mov abs
		spc.mov(spc.addrAbs())
	case 0xe6:
		// mov ind
		spc.mov(spc.addrInd())
	case 0xe7:
		// mov idx
		spc.mov(spc.addrIdx())
	case 0xe8:
		// mov imm
		spc.mov(spc.addrImm())
	case 0xe9:
		// movx abs
		spc.movx(spc.addrAbs())
	case 0xea:
		// not1 abs.bit
		var addr uint16 = 0
		var bit byte = spc.addrAbsBit(&addr)
		var result byte = spc.Read(addr) ^ (1 << bit)
		spc.Write(addr, result)
	case 0xeb:
		// movy dp
		spc.movy(spc.addrDp())
	case 0xec:
		// movy abs
		spc.movy(spc.addrAbs())
	case 0xed:
		// notc imp
		if !spc.CheckFlag(SPCFlagsCarry) {
			spc.c = 0x01
		} else {
			spc.c = 0x00
		}
	case 0xee:
		// popy imp
		spc.y = spc.pullByte()
	case 0xef:
		// sleep imp
		spc.stopped = true // no interrupts, so sleeping stops as well
	case 0xf0:
		// beq rel
		spc.branch(spc.readOpcode(), spc.CheckFlag(SPCFlagsZero))
	case 0xf4:
		// mov dpx
		spc.mov(spc.addrDpx())
	case 0xf5:
		// mov abx
		spc.mov(spc.addrAbx())
	case 0xf6:
		// mov aby
		spc.mov(spc.addrAby())
	case 0xf7:
		// mov idy
		spc.mov(spc.addrIdy())
	case 0xf8:
		// movx dp
		spc.movx(spc.addrDp())
	case 0xf9:
		// movx dpy
		spc.movx(spc.addrDpy())
	case 0xfa:
		// movm dp, dp
		var src uint16 = 0
		var dst uint16 = spc.addrDpDp(&src)
		var val byte = spc.Read(src)
		spc.Write(dst, val)
	case 0xfb:
		// movy dpx
		spc.movy(spc.addrDpx())
	case 0xfc:
		// incy imp
		spc.y++
		spc.setZN(spc.y)
	case 0xfd:
		// movya imp
		spc.y = spc.a
		spc.setZN(spc.y)
	case 0xfe:
		// dbnzy rel
		spc.y--
		spc.branch(spc.readOpcode(), spc.y != 0)
	case 0xff:
		// stop imp
		spc.stopped = true
	}
}

func (spc *SPC) and(addr uint16) {
	spc.a &= spc.Read(addr)
	spc.setZN(spc.a)
}

func (spc *SPC) andm(dst uint16, src uint16) {
	var value byte = spc.Read(src)
	var result byte = spc.Read(dst) & value
	spc.Write(dst, result)
	spc.setZN(result)
}

func (spc *SPC) or(addr uint16) {
	spc.a |= spc.Read(addr)
	spc.setZN(spc.a)
}

func (spc *SPC) orm(dst uint16, src uint16) {
	var value byte = spc.Read(src)
	var result byte = spc.Read(dst) | value
	spc.Write(dst, result)
	spc.setZN(result)
}

func (spc *SPC) eor(addr uint16) {
	spc.a ^= spc.Read(addr)
	spc.setZN(spc.a)
}

func (spc *SPC) eorm(dst uint16, src uint16) {
	var value byte = spc.Read(src)
	var result byte = spc.Read(dst) ^ value
	spc.Write(dst, result)
	spc.setZN(result)
}

func (spc *SPC) adc(addr uint16) {
	var value byte = spc.Read(addr)
	var result int = int(spc.a) + int(value) + int(spc.c)
	if (spc.a&0x80) == (value&0x80) && (value&0x80) != (byte(result)&0x80) {
		spc.v = 0x01
	} else {
		spc.v = 0x00
	}
	if ((spc.a & 0xf) + (value & 0xf) + spc.c) > 0xf {
		spc.h = 0x01
	} else {
		spc.h = 0x00
	}
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.a = byte(result)
	spc.setZN(spc.a)
}

func (spc *SPC) adcm(dst uint16, src uint16) {
	var value byte = spc.Read(src)
	var applyOn byte = spc.Read(dst)
	var result int = int(applyOn) + int(value) + int(spc.c)
	if (applyOn&0x80) == (value&0x80) && (value&0x80) != (byte(result)&0x80) {
		spc.v = 0x01
	} else {
		spc.v = 0x00
	}
	if ((applyOn & 0xf) + (value & 0xf) + spc.c) > 0xf {
		spc.h = 0x01
	} else {
		spc.h = 0x00
	}
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.Write(dst, byte(result))
	spc.setZN(byte(result))
}

func (spc *SPC) sbc(addr uint16) {
	var value byte = spc.Read(addr) ^ 0xff
	var result int = int(spc.a) + int(value) + int(spc.c)
	if (spc.a&0x80) == (value&0x80) && (value&0x80) != (byte(result)&0x80) {
		spc.v = 0x01
	} else {
		spc.v = 0x00
	}
	if ((spc.a & 0xf) + (value & 0xf) + spc.c) > 0xf {
		spc.h = 0x01
	} else {
		spc.h = 0x00
	}
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.a = byte(result)
	spc.setZN(spc.a)
}

func (spc *SPC) sbcm(dst uint16, src uint16) {
	var value byte = spc.Read(src) ^ 0xff
	var applyOn byte = spc.Read(dst)
	var result int = int(applyOn) + int(value) + int(spc.c)
	if (applyOn&0x80) == (value&0x80) && (value&0x80) != (byte(result)&0x80) {
		spc.v = 0x01
	} else {
		spc.v = 0x00
	}
	if ((applyOn & 0xf) + (value & 0xf) + spc.c) > 0xf {
		spc.h = 0x01
	} else {
		spc.h = 0x00
	}
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.Write(dst, byte(result))
	spc.setZN(byte(result))
}

func (spc *SPC) cmp(addr uint16) {
	var value byte = spc.Read(addr) ^ 0xff
	var result int = int(spc.a) + int(value) + 1
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.setZN(byte(result))
}

func (spc *SPC) cmpx(addr uint16) {
	var value byte = spc.Read(addr) ^ 0xff
	var result int = int(spc.x) + int(value) + 1
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.setZN(byte(result))
}

func (spc *SPC) cmpy(addr uint16) {
	var value byte = spc.Read(addr) ^ 0xff
	var result int = int(spc.y) + int(value) + 1
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.setZN(byte(result))
}

func (spc *SPC) cmpm(dst uint16, src uint16) {
	var value byte = spc.Read(src) ^ 0xff
	var result int = int(spc.Read(dst)) + int(value) + 1
	if result > 0xff {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.setZN(byte(result))
}

func (spc *SPC) mov(addr uint16) {
	spc.a = spc.Read(addr)
	spc.setZN(spc.a)
}

func (spc *SPC) movx(addr uint16) {
	spc.x = spc.Read(addr)
	spc.setZN(spc.x)
}

func (spc *SPC) movy(addr uint16) {
	spc.y = spc.Read(addr)
	spc.setZN(spc.y)
}

func (spc *SPC) movs(addr uint16) {
	spc.Read(addr)
	spc.Write(addr, spc.a)
}

func (spc *SPC) movsx(addr uint16) {
	spc.Read(addr)
	spc.Write(addr, spc.x)
}

func (spc *SPC) movsy(addr uint16) {
	spc.Read(addr)
	spc.Write(addr, spc.y)
}

func (spc *SPC) asl(addr uint16) {
	var val byte = spc.Read(addr)
	if (val & 0x80) > 0 {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	val <<= 1
	spc.Write(addr, val)
	spc.setZN(val)
}

func (spc *SPC) lsr(addr uint16) {
	var val byte = spc.Read(addr)
	if (val & 1) > 0 {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	val >>= 1
	spc.Write(addr, val)
	spc.setZN(val)
}

func (spc *SPC) rol(addr uint16) {
	var val byte = spc.Read(addr)
	var newC bool = (val & 0x80) > 0
	val = (val << 1) | spc.c
	if newC {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.Write(addr, val)
	spc.setZN(val)
}

func (spc *SPC) ror(addr uint16) {
	var val byte = spc.Read(addr)
	var newC bool = (val & 1) > 0
	val = (val >> 1) | (spc.c << 7)
	if newC {
		spc.c = 0x01
	} else {
		spc.c = 0x00
	}
	spc.Write(addr, val)
	spc.setZN(val)
}

func (spc *SPC) inc(addr uint16) {
	var val byte = spc.Read(addr) + 1
	spc.Write(addr, val)
	spc.setZN(val)
}

func (spc *SPC) dec(addr uint16) {
	var val byte = spc.Read(addr) - 1
	spc.Write(addr, val)
	spc.setZN(val)
}
