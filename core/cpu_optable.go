package core

type opcode = func(w *w65816)

var opTable = [256]opcode{
	op00, op01, op02, op03, op04, op05, op06, op07, op08, op09, op0A, op0B, op0C, op0D, op0E, op0F,
	op10, op11, op12, op13, op14, op15, op16, op17, op18, op19, op1A, op1B, op1C, op1D, op1E, op1F,
	op20, op21, op22, op23, op24, op25, op26, op27, op28, op29, op2A, op2B, op2C, op2D, op2E, op2F,
	op30, op31, op32, op33, op34, op35, op36, op37, op38, op39, op3A, op3B, op3C, op3D, op3E, op3F,
	op40, op41, op42, op43, op44, op45, op46, op47, op48, op49, op4A, op4B, op4C, op4D, op4E, op4F,
	op50, op51, op52, op53, op54, op55, op56, op57, op58, op59, op5A, op5B, op5C, op5D, op5E, op5F,
	op60, op61, op62, op63, op64, op65, op66, op67, op68, op69, op6A, op6B, op6C, op6D, op6E, op6F,
	op70, op71, op72, op73, op74, op75, op76, op77, op78, op79, op7A, op7B, op7C, op7D, op7E, op7F,
	op80, op81, op82, op83, op84, op85, op86, op87, op88, op89, op8A, op8B, op8C, op8D, op8E, op8F,
	op90, op91, op92, op93, op94, op95, op96, op97, op98, op99, op9A, op9B, op9C, op9D, op9E, op9F,
	opA0, opA1, opA2, opA3, opA4, opA5, opA6, opA7, opA8, opA9, opAA, opAB, opAC, opAD, opAE, opAF,
	opB0, opB1, opB2, opB3, opB4, opB5, opB6, opB7, opB8, opB9, opBA, opBB, opBC, opBD, opBE, opBF,
	opC0, opC1, opC2, opC3, opC4, opC5, opC6, opC7, opC8, opC9, opCA, opCB, opCC, opCD, opCE, opCF,
	opD0, opD1, opD2, opD3, opD4, opD5, opD6, opD7, opD8, opD9, opDA, opDB, opDC, opDD, opDE, opDF,
	opE0, opE1, opE2, opE3, opE4, opE5, opE6, opE7, opE8, opE9, opEA, opEB, opEC, opED, opEE, opEF,
	opF0, opF1, opF2, opF3, opF4, opF5, opF6, opF7, opF8, opF9, opFA, opFB, opFC, opFD, opFE, opFF,
}

func op00(w *w65816) {
	w.exception(BRK)
}

func op01(w *w65816) {
	w.indirectX(w.ORA)
}

func op02(w *w65816) {
	w.exception(COP)
}

func op03(w *w65816) {
	w.stackRelative(w.ORA)
}

func op04(w *w65816) {
	w.zeropage(w.TSB)
}

func op05(w *w65816) {
	w.zeropage(w.ORA)
}

func op06(w *w65816) {
	w.zeropage(w.ASL)
}

func op07(w *w65816) {
	w.indirectLong(w.ORA)
}

// PHP, `PUSH P`
func op08(w *w65816) {
	w.PUSH8(w.r.p.pack(), func(w *w65816) {})
}

func op09(w *w65816) {
	w.imm(w.ORA)
}

// ASL A
func op0A(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.m {
			w.r.p.c = bit(w.r.a, 7)
			w.r.l(&w.r.a, uint8(w.r.a)<<1)
			w.r.p.setFlags(zn(w.r.a, 8))
			return
		}

		w.r.p.c = bit(w.r.a, 15)
		w.r.a <<= 1
		w.r.p.setFlags(zn(w.r.a, 16))
	}
}

func op0B(w *w65816) {
	w.PUSH(&w.r.d)
}

func op0C(w *w65816) {
	w.absolute(w.TSB)
}

// OR A,[nnnn]
func op0D(w *w65816) {
	w.absolute(w.ORA)
}

func op0E(w *w65816) {
	w.absolute(w.ASL)
}

func op0F(w *w65816) {
	w.absoluteLong(w.ORA)
}

// BPL
func op10(w *w65816) {
	w.JMP8(func() bool { return !w.r.p.n })
}

func op11(w *w65816) {
	w.indirectY(w.ORA)
}

func op12(w *w65816) {
	w.indirect(w.ORA)
}

func op13(w *w65816) {
	w.stackRelativeY(w.ORA)
}

func op14(w *w65816) {
	w.zeropage(w.TRB)
}

func op15(w *w65816) {
	w.zeropageX(w.ORA)
}

func op16(w *w65816) {
	w.zeropageX(w.ASL)
}

func op17(w *w65816) {
	w.indirectLongY(w.ORA)
}

// CLC
func op18(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.c = false }
}

func op19(w *w65816) {
	w.absoluteY(w.ORA)
}

// 0x1A:
//
//	Native: `INA`
//	Nocash: `INC A`
func op1A(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.m {
			w.INC8(&w.r.a)
			return
		}
		w.INC16(&w.r.a)
	}
}

func op1B(w *w65816) {
	w.MOV(&w.r.s, &w.r.a)
}

func op1C(w *w65816) {
	w.absolute(w.TRB)
}

func op1D(w *w65816) {
	w.absoluteX(w.ORA)
}

func op1E(w *w65816) {
	w.absoluteX(w.ASL)
}

func op1F(w *w65816) {
	w.absoluteLongX(w.ORA)
}

// 0x20:
//
//	Effect: `[S]=PC+2,PC=nnnn`
//	Native: `JSR nnnn`
//	Nocash: `CALL nnnn`
func op20(w *w65816) {
	// 2,3
	w.imm16(func(pc uint16) {
		w.state = CPU_DUMMY_READ // 4(IO)
		w.inst = func(w *w65816) {
			// 5,6
			w.PUSH16(w.r.pc.offset-1, func(w *w65816) {
				w.r.pc = u24(w.r.pc.bank, pc)
			})
		}
	})
}

func op21(w *w65816) {
	w.indirectX(w.AND)
}

func op22(w *w65816) {
	// 2, 3
	w.imm16(func(pc uint16) {
		// 4
		w.PUSH8(w.r.pc.bank, func(w *w65816) {
			// 5
			w.state = CPU_DUMMY_READ
			w.inst = func(w *w65816) {
				// 6
				w.imm8(func(pb uint8) {
					// 7, 8
					w.PUSH16(w.r.pc.offset-1, func(w *w65816) {
						w.r.pc = u24(pb, pc)
					})
				})
			}
		})
	})
}

func op23(w *w65816) {
	w.stackRelative(w.AND)
}

func op24(w *w65816) {
	w.zeropage(w.BIT)
}

func op25(w *w65816) {
	w.zeropage(w.AND)
}

func op26(w *w65816) {
	w.zeropage(w.ROL)
}

func op27(w *w65816) {
	w.indirectLong(w.AND)
}

// PLP, `POP P`
func op28(w *w65816) {
	// 2
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		// 3
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			// 4
			w.POP8(func(val uint8) {
				w.r.p.setPacked(val)
			})
		}
	}
}

func op29(w *w65816) {
	w.imm(w.AND)
}

// ROL A
func op2A(w *w65816) {
	c := w.r.p.c

	if w.r.emulation || w.r.p.m {
		w.r.p.c = bit(w.r.a, 7)
		a := uint8(w.r.a) << 1
		a = setBit(a, 0, c)
		w.r.l(&w.r.a, a)
		w.r.p.setFlags(zn(w.r.a, 8))
		return
	}

	w.r.p.c = bit(w.r.a, 15)
	w.r.a <<= 1
	w.r.a = setBit(w.r.a, 0, c)
	w.r.p.setFlags(zn(w.r.a, 16))
}

func op2B(w *w65816) {
	w.POP(&w.r.d)
}

func op2C(w *w65816) {
	w.absolute(w.BIT)
}

func op2D(w *w65816) {
	w.absolute(w.AND)
}

func op2E(w *w65816) {
	w.absolute(w.ROL)
}

func op2F(w *w65816) {
	w.absoluteLong(w.AND)
}

// BMI
func op30(w *w65816) {
	w.JMP8(func() bool { return w.r.p.n })
}

func op31(w *w65816) {
	w.indirectY(w.AND)
}

func op32(w *w65816) {
	w.indirect(w.AND)
}

func op33(w *w65816) {
	w.stackRelativeY(w.AND)
}

func op34(w *w65816) {
	w.zeropageX(w.BIT)
}

func op35(w *w65816) {
	w.zeropageX(w.AND)
}

func op36(w *w65816) {
	w.zeropageX(w.ROL)
}

func op37(w *w65816) {
	w.indirectLongY(w.AND)
}

// SEC, STC (C=1)
func op38(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.c = true }
}

func op39(w *w65816) {
	w.absoluteY(w.AND)
}

// DEA, `DEC A`
func op3A(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.m {
			w.DEC8(&w.r.a)
			return
		}
		w.DEC16(&w.r.a)
	}
}

func op3B(w *w65816) {
	w.MOV(&w.r.a, &w.r.s)
}

func op3C(w *w65816) {
	w.absoluteX(w.BIT)
}

func op3D(w *w65816) {
	w.absoluteX(w.AND)
}

func op3E(w *w65816) {
	w.absoluteX(w.ROL)
}

func op3F(w *w65816) {
	w.absoluteLongX(w.AND)
}

// RTI
func op40(w *w65816) {
	// 2
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		// 3
		w.POP8(func(p uint8) {
			w.r.p.setPacked(p)
			// 4
			w.POP16(func(pc uint16) {
				// 6
				if w.r.emulation {
					w.r.pc = u24(w.r.pc.bank, pc)
					return
				}

				w.POP8(func(pb uint8) {
					w.r.pc = u24(pb, pc)
				})
			})
		})
	}
}

func op41(w *w65816) {
	w.indirectX(w.EOR)
}

// WDM (Reserved for future use)
func op42(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.pc = w.r.pc.plus(1) }
}

func op43(w *w65816) {
	w.stackRelative(w.EOR)
}

// 0x44:
//
//	Effect: Copy A+1 bytes X to Y (Decrement)
//	Native: `MVP ss,dd` // MVN=Block Move Positive
//	Nocash: `LDDR [dd:Y],[ss:X],A+1`
func op44(w *w65816) {
	// 2, 3
	w.imm16(func(nnnn uint16) {
		ss, dd := uint8(nnnn>>8), uint8(nnnn)

		// 4
		w.read8(u24(ss, w.r.x), func(val uint8) {
			w.write8(u24(dd, w.r.y), val, func() {
				w.r.x--
				w.r.y--
				w.r.a--
				if w.r.a != 0xffff {
					w.r.pc.offset -= 3 // まだ転送が終わってないのでもう一度実行させる
				}

				addCycle(w.cycles, FAST*2)
			})
		})
	})
}

func op45(w *w65816) {
	w.zeropage(w.EOR)
}

func op46(w *w65816) {
	w.zeropage(w.LSR)
}

func op47(w *w65816) {
	w.indirectLong(w.EOR)
}

func op48(w *w65816) {
	w.PUSH(&w.r.a)
}

func op49(w *w65816) {
	w.imm(w.EOR)
}

// 0x4A:
//
//	Effect: Shift one bit right
//	Native: `LSR A`
//	Nocash: `SHR A`
func op4A(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.m {
			w.LSR8(&w.r.a)
			return
		}

		w.LSR16(&w.r.a)
	}
}

func op4B(w *w65816) {
	w.PUSH8(w.r.pc.bank, func(w *w65816) {})
}

// JMP nnnn (PC=nnnn)
func op4C(w *w65816) {
	w.imm16(func(val uint16) { w.r.pc = u24(w.r.pc.bank, val) })
}

func op4D(w *w65816) {
	w.absolute(w.EOR)
}

func op4E(w *w65816) {
	w.absolute(w.LSR)
}

func op4F(w *w65816) {
	w.absoluteLong(w.EOR)
}

// BVC
func op50(w *w65816) {
	w.JMP8(func() bool { return !w.r.p.v })
}

func op51(w *w65816) {
	w.indirectY(w.EOR)
}

func op52(w *w65816) {
	w.indirect(w.EOR)
}

func op53(w *w65816) {
	w.stackRelativeY(w.EOR)
}

// 0x54:
//
//	Effect: Copy A+1 bytes X to Y (Increment)
//	Native: `MVN ss,dd` // MVN=Block Move Negative
//	Nocash: `LDIR [dd:Y],[ss:X],A+1`
//	Alias:  `MVN xyc`
func op54(w *w65816) {
	w.imm16(func(nnnn uint16) {
		ss, dd := uint8(nnnn>>8), uint8(nnnn)

		w.read8(u24(ss, w.r.x), func(val uint8) {
			w.write8(u24(dd, w.r.y), val, func() {
				w.r.x++
				w.r.y++
				w.r.a--
				if w.r.a != 0xffff {
					w.r.pc.offset -= 3 // まだ転送が終わってないのでもう一度実行させる
				}

				addCycle(w.cycles, FAST*2)
			})
		})
	})
}

func op55(w *w65816) {
	w.zeropageX(w.EOR)
}

func op56(w *w65816) {
	w.zeropageX(w.LSR)
}

func op57(w *w65816) {
	w.indirectLongY(w.EOR)
}

// EI (I=0)
func op58(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.i = false }
}

func op59(w *w65816) {
	w.absoluteY(w.EOR)
}

func op5A(w *w65816) {
	w.PUSH(&w.r.y)
}

func op5B(w *w65816) {
	w.MOV(&w.r.d, &w.r.a)
}

// JMP nnnnnn (PB:PC=nnnnnn)
func op5C(w *w65816) {
	w.imm24(func(pc uint24) { w.r.pc = pc })
}

func op5D(w *w65816) {
	w.absoluteX(w.EOR)
}

func op5E(w *w65816) {
	w.absoluteX(w.LSR)
}

func op5F(w *w65816) {
	w.absoluteLongX(w.EOR)
}

// RTS, RET (PC=[S+1]+1, S=S+2)
func op60(w *w65816) {
	// 2(IO)
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		// 3(IO)
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			// 4, 5
			w.POP16(func(pc uint16) {
				pc += 1

				// 6
				w.state = CPU_DUMMY_READ
				w.inst = func(w *w65816) { w.r.pc = u24(w.r.pc.bank, pc) }
			})
		}
	}
}

func op61(w *w65816) {
	w.indirectX(w.ADC)
}

// PER rel16 (`PUSH disp16`)
func op62(w *w65816) {
	w.imm16(func(nnnn uint16) {
		disp := int(int16(nnnn))
		val := w.r.pc.plus(disp)
		w.PUSH16(val.offset, func(w *w65816) {})
	})
}

func op63(w *w65816) {
	w.stackRelative(w.ADC)
}

func op64(w *w65816) {
	w.zeropage(w.STN(nil))
}

func op65(w *w65816) {
	w.zeropage(w.ADC)
}

func op66(w *w65816) {
	w.zeropage(w.ROR)
}

func op67(w *w65816) {
	w.indirectLong(w.ADC)
}

func op68(w *w65816) {
	w.POP(&w.r.a)
}

func op69(w *w65816) {
	w.imm(w.ADC)
}

// ROR A
func op6A(w *w65816) {
	c := w.r.p.c

	if w.r.emulation || w.r.p.m {
		w.r.p.c = bit(w.r.a, 0)
		a := uint8(w.r.a) >> 1
		a = setBit(a, 7, c)
		w.r.l(&w.r.a, a)
		w.r.p.setFlags(zn(w.r.a, 8))
		return
	}

	w.r.p.c = bit(w.r.a, 0)
	w.r.a >>= 1
	w.r.a = setBit(w.r.a, 15, c)
	w.r.p.setFlags(zn(w.r.a, 16))
}

// RTL
func op6B(w *w65816) {
	// 2
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		// 3
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			w.POP16(func(pc uint16) {
				w.POP8(func(pb uint8) {
					w.r.pc = u24(pb, pc).plus(1)
				})
			})
		}
	}
}

// JMP (nnnn)
func op6C(w *w65816) {
	// 2, 3
	w.imm16(func(nnnn uint16) {
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			addr := u24(0, nnnn)
			w.read16(addr, func(pc uint16) {
				w.r.pc = u24(w.r.pc.bank, pc)
			})
		}
	})
}

func op6D(w *w65816) {
	w.absolute(w.ADC)
}

func op6E(w *w65816) {
	w.absolute(w.ROR)
}

func op6F(w *w65816) {
	w.absoluteLong(w.ADC)
}

// BVS
func op70(w *w65816) {
	w.JMP8(func() bool { return w.r.p.v })
}

func op71(w *w65816) {
	w.indirectY(w.ADC)
}

func op72(w *w65816) {
	w.indirect(w.ADC)
}

func op73(w *w65816) {
	w.stackRelativeY(w.ADC)
}

func op74(w *w65816) {
	w.zeropageX(w.STN(nil))
}

func op75(w *w65816) {
	w.zeropageX(w.ADC)
}

func op76(w *w65816) {
	w.zeropageX(w.ROR)
}

func op77(w *w65816) {
	w.indirectLongY(w.ADC)
}

// DI (I=1)
func op78(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.i = true }
}

func op79(w *w65816) {
	w.absoluteY(w.ADC)
}

func op7A(w *w65816) {
	w.POP(&w.r.y)
}

func op7B(w *w65816) {
	w.MOV(&w.r.a, &w.r.d)
}

func op7C(w *w65816) {
	// 2, 3
	w.imm16(func(nnnn uint16) {
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			addr := u24(w.r.pc.bank, nnnn).plus(int(w.r.x))
			w.read16(addr, func(pc uint16) {
				w.r.pc = u24(w.r.pc.bank, pc)
			})
		}
	})
}

func op7D(w *w65816) {
	w.absoluteX(w.ADC)
}

func op7E(w *w65816) {
	w.absoluteX(w.ROR)
}

func op7F(w *w65816) {
	w.absoluteLongX(w.ADC)
}

func op80(w *w65816) {
	w.JMP8(func() bool { return true })
}

func op81(w *w65816) {
	w.indirectX(w.STN(&w.r.a))
}

// BRL disp16 (PC=PC+/-disp16)
func op82(w *w65816) {
	w.imm16(func(val uint16) {
		disp16 := int(int16(val))
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			w.r.pc = w.r.pc.plus(disp16)
		}
	})
}

func op83(w *w65816) {
	w.stackRelative(w.STN(&w.r.a))
}

func op84(w *w65816) {
	w.zeropage(w.STN(&w.r.y))
}

func op85(w *w65816) {
	w.zeropage(w.STN(&w.r.a))
}

func op86(w *w65816) {
	w.zeropage(w.STN(&w.r.x))
}

func op87(w *w65816) {
	w.indirectLong(w.STN(&w.r.a))
}

// DEY (Y=Y-1)
func op88(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.x {
			w.DEC8(&w.r.y)
			return
		}

		w.DEC16(&w.r.y)
	}
}

// `BIT #nn`, `TEST A,nn`
// Zフラグしか変更しないことに注意
func op89(w *w65816) {
	if w.r.emulation || w.r.p.m {
		w.imm8(func(nn uint8) {
			z := uint8(w.r.a)&nn == 0
			w.r.p.setFlags(flag{'z', z})
		})
		return
	}

	w.imm16(func(nnnn uint16) {
		z := w.r.a&nnnn == 0
		w.r.p.setFlags(flag{'z', z})
	})
}

// TXA (A=X)
func op8A(w *w65816) {
	w.MOV(&w.r.a, &w.r.x)
}

func op8B(w *w65816) {
	w.PUSH8(w.r.db, func(w *w65816) {})
}

func op8C(w *w65816) {
	w.absolute(w.STN(&w.r.y))
}

func op8D(w *w65816) {
	w.absolute(w.STN(&w.r.a))
}

func op8E(w *w65816) {
	w.absolute(w.STN(&w.r.x))
}

func op8F(w *w65816) {
	w.absoluteLong(w.STN(&w.r.a))
}

// BCC(BLT)
func op90(w *w65816) {
	w.JMP8(func() bool { return !w.r.p.c })
}

func op91(w *w65816) {
	w.indirectY(w.STN(&w.r.a))
}

func op92(w *w65816) {
	w.indirect(w.STN(&w.r.a))
}

func op93(w *w65816) {
	w.stackRelativeY(w.STN(&w.r.a))
}

func op94(w *w65816) {
	w.zeropageX(w.STN(&w.r.y))
}

func op95(w *w65816) {
	w.zeropageX(w.STN(&w.r.a))
}

func op96(w *w65816) {
	w.zeropageY(w.STN(&w.r.x))
}

func op97(w *w65816) {
	w.indirectLongY(w.STN(&w.r.a))
}

func op98(w *w65816) {
	w.MOV(&w.r.a, &w.r.y)
}

func op99(w *w65816) {
	w.absoluteY(w.STN(&w.r.a))
}

func op9A(w *w65816) {
	w.MOV(&w.r.s, &w.r.x)
}

func op9B(w *w65816) {
	w.MOV(&w.r.y, &w.r.x)
}

func op9C(w *w65816) {
	w.absolute(w.STN(nil))
}

func op9D(w *w65816) {
	w.absoluteX(w.STN(&w.r.a))
}

func op9E(w *w65816) {
	w.absoluteX(w.STN(nil))
}

func op9F(w *w65816) {
	w.absoluteLongX(w.STN(&w.r.a))
}

func opA0(w *w65816) {
	w.imm(w.LDNm(&w.r.y))
}

// LDA (nn, X), `MOV A,[[nn+X]]`
func opA1(w *w65816) {
	w.indirectX(w.LDNm(&w.r.a))
}

// LDX #nn (X=nn)
func opA2(w *w65816) {
	w.imm(w.LDNm(&w.r.x))
}

// LDA nn,S
func opA3(w *w65816) {
	w.stackRelative(w.LDNm(&w.r.a))
}

func opA4(w *w65816) {
	w.zeropage(w.LDNm(&w.r.y))
}

func opA5(w *w65816) {
	w.zeropage(w.LDNm(&w.r.a))
}

func opA6(w *w65816) {
	w.zeropage(w.LDNm(&w.r.x))
}

// LDA [nn] (`MOV A,[FAR[nn]]`)
func opA7(w *w65816) {
	w.indirectLong(w.LDNm(&w.r.a))
}

func opA8(w *w65816) {
	w.MOV(&w.r.y, &w.r.a)
}

func opA9(w *w65816) {
	w.imm(w.LDNm(&w.r.a))
}

func opAA(w *w65816) {
	w.MOV(&w.r.x, &w.r.a)
}

// PLB, `POP DB`
func opAB(w *w65816) {
	// 2
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		// 3
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			// 4
			w.POP8(func(val uint8) {
				w.r.db = val
				w.r.p.setFlags(zn(uint16(w.r.db), 8))
			})
		}
	}
}

// MOV Y,[nnnn]
func opAC(w *w65816) {
	w.absolute(w.LDNm(&w.r.y))
}

// MOV A,[nnnn]
func opAD(w *w65816) {
	w.absolute(w.LDNm(&w.r.a))
}

func opAE(w *w65816) {
	w.absolute(w.LDNm(&w.r.x))
}

func opAF(w *w65816) {
	w.absoluteLong(w.LDNm(&w.r.a))
}

// BCS(BGE)
func opB0(w *w65816) {
	w.JMP8(func() bool { return w.r.p.c })
}

func opB1(w *w65816) {
	w.indirectY(w.LDNm(&w.r.a))
}

func opB2(w *w65816) {
	w.indirect(w.LDNm(&w.r.a))
}

func opB3(w *w65816) {
	w.stackRelativeY(w.LDNm(&w.r.a))
}

func opB4(w *w65816) {
	w.zeropageX(w.LDNm(&w.r.y))
}

// LDA nn,X
func opB5(w *w65816) {
	w.zeropageX(w.LDNm(&w.r.a))
}

// LDX nn,Y
func opB6(w *w65816) {
	w.zeropageY(w.LDNm(&w.r.x))
}

// LDA [nn],y
func opB7(w *w65816) {
	w.indirectLongY(w.LDNm(&w.r.a))
}

// CLV (V=0)
func opB8(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.v = false }
}

func opB9(w *w65816) {
	w.absoluteY(w.LDNm(&w.r.a))
}

func opBA(w *w65816) {
	w.MOV(&w.r.x, &w.r.s)
}

func opBB(w *w65816) {
	w.MOV(&w.r.x, &w.r.y)
}

func opBC(w *w65816) {
	w.absoluteX(w.LDNm(&w.r.y))
}

func opBD(w *w65816) {
	w.absoluteX(w.LDNm(&w.r.a))
}

// LDX nnnn,Y (`MOV X,[nnnn+Y]`)
func opBE(w *w65816) {
	w.absoluteY(w.LDNm(&w.r.x))
}

func opBF(w *w65816) {
	w.absoluteLongX(w.LDNm(&w.r.a))
}

func opC0(w *w65816) {
	w.imm(w.CMP(&w.r.y))
}

func opC1(w *w65816) {
	w.indirectX(w.CMP(&w.r.a))
}

// REP, `CLR P,nn` (`P=P AND NOT nn`)
func opC2(w *w65816) {
	// 2
	w.imm8(func(nn uint8) {
		old := w.r.p.pack()
		p := old & ^nn
		// fmt.Printf("REP #%02x: 0x%02x -> 0x%02x in %s\n", nn, old, p, w.lastInstAddr)
		w.r.p.setPacked(p)
		addCycle(w.cycles, FAST)
	})
}

func opC3(w *w65816) {
	w.stackRelative(w.CMP(&w.r.a))
}

func opC4(w *w65816) {
	w.zeropage(w.CMP(&w.r.y))
}

func opC5(w *w65816) {
	w.zeropage(w.CMP(&w.r.a))
}

// 0xC6:
//
//	Effect: `[D+nn]=[D+nn]-1`
//	Native: `DEC nn`
//	Nocash: `DEC [nn]`
func opC6(w *w65816) {
	// 2, 2a
	w.zeropage(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			// 3
			w.read8(addr, func(val uint8) {
				val--
				// 4
				w.state = CPU_DUMMY_READ
				w.inst = func(w *w65816) {
					// 5
					w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
				}
			})
			return
		}

		// 3,3a
		w.read16(addr, func(val uint16) {
			val--
			// 5a, 5
			w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
		})
	})
}

func opC7(w *w65816) {
	w.indirectLong(w.CMP(&w.r.a))
}

// INY (Y=Y+1)
func opC8(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.x {
			w.INC8(&w.r.y)
			return
		}
		w.INC16(&w.r.y)
	}
}

func opC9(w *w65816) {
	w.imm(w.CMP(&w.r.a))
}

// DEX (X=X-1)
func opCA(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.x {
			w.DEC8(&w.r.x)
			return
		}

		w.DEC16(&w.r.x)
	}
}

// WAI
func opCB(w *w65816) {
	if w.irqPending && w.r.p.i {
		return
	}
	w.halted = true
	addCycle(w.cycles, *w.nextEvent)
}

func opCC(w *w65816) {
	w.absolute(w.CMP(&w.r.y))
}

func opCD(w *w65816) {
	w.absolute(w.CMP(&w.r.a))
}

// 0xCE:
//
//	Effect: `[DB:nnnn]=[DB:nnnn]-1`
//	Native: `DEC nnnn`
//	Nocash: `DEC [nnnn]`
func opCE(w *w65816) {
	w.absolute(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val -= 1
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})
			return
		}

		w.read16(addr, func(val uint16) {
			val -= 1
			w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
		})
	})
}

func opCF(w *w65816) {
	w.absoluteLong(w.CMP(&w.r.a))
}

func opD0(w *w65816) {
	w.JMP8(func() bool { return !w.r.p.z })
}

func opD1(w *w65816) {
	w.indirectY(w.CMP(&w.r.a))
}

func opD2(w *w65816) {
	w.indirect(w.CMP(&w.r.a))
}

func opD3(w *w65816) {
	w.stackRelativeY(w.CMP(&w.r.a))
}

// PEI
func opD4(w *w65816) {
	w.zeropage(func(addr uint24) {
		w.read16(addr, func(val uint16) {
			w.PUSH16(val, func(w *w65816) {})
		})
	})
}

func opD5(w *w65816) {
	w.zeropageX(w.CMP(&w.r.a))
}

// 0xD6:
//
//	Effect: `[D+nn+X]=[D+nn+X]-1`
//	Native: `DEC nn,X`
//	Nocash: `DEC [nn+X]`
func opD6(w *w65816) {
	w.zeropageX(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val -= 1
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})
			return
		}

		// 3,3a
		w.read16(addr, func(val uint16) {
			val -= 1
			w.write16(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 16)) })
		})
	})
}

func opD7(w *w65816) {
	w.indirectLongY(w.CMP(&w.r.a))
}

// CLD (D=0)
func opD8(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.d = false }
}

func opD9(w *w65816) {
	w.absoluteY(w.CMP(&w.r.a))
}

// PHX, `PUSH X`
func opDA(w *w65816) {
	if w.r.emulation || w.r.p.x {
		w.PUSH8(uint8(w.r.x), func(w *w65816) {})
		return
	}
	w.PUSH16(w.r.x, func(w *w65816) {})
}

// STP
func opDB(w *w65816) {}

// JML(Jump Long, PB:PC=[00:nnnn])
func opDC(w *w65816) {
	w.imm16(func(nnnn uint16) {
		addr := u24(0, nnnn)
		w.read16(addr, func(ofs uint16) {
			w.read8(addr.plus(2), func(bank uint8) {
				w.r.pc = u24(bank, ofs)
			})
		})
	})
}

func opDD(w *w65816) {
	w.absoluteX(w.CMP(&w.r.a))
}

// 0xDE:
//
//	Effect: `[DB:nnnn+X]=[DB:nnnn+X]-1`
//	Native: `DEC nnnn,X`
//	Nocash: `DEC [nnnn+X]`
func opDE(w *w65816) {
	w.absoluteX(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val -= 1
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})
			return
		}

		w.read16(addr, func(val uint16) {
			val -= 1
			w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
		})
	})
}

func opDF(w *w65816) {
	w.absoluteLongX(w.CMP(&w.r.a))
}

func opE0(w *w65816) {
	w.imm(w.CMP(&w.r.x))
}

func opE1(w *w65816) {
	w.indirectX(w.SBC)
}

// SEP, `SET P,nn` (`P = (P OR nn)`)
func opE2(w *w65816) {
	w.imm8(func(nn uint8) {
		old := w.r.p.pack()
		p := old | nn
		// fmt.Printf("SEP #%02x: 0x%02x -> 0x%02x in %s\n", nn, old, p, w.lastInstAddr)
		w.r.p.setPacked(p)

		w.dummyRead(func(w *w65816) {})
	})
}

func opE3(w *w65816) {
	w.stackRelative(w.SBC)
}

func opE4(w *w65816) {
	w.zeropage(w.CMP(&w.r.x))
}

func opE5(w *w65816) {
	w.zeropage(w.SBC)
}

// 0xE6:
//
//	Effect: `[D+nn]=[D+nn]+1`
//	Native: `INC nn`
//	Nocash: `INC [nn]`
func opE6(w *w65816) {
	w.zeropage(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val++
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})

			return
		}

		// 3,3a
		w.read16(addr, func(val uint16) {
			val += 1
			w.write16(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 16)) })
		})
	})
}

func opE7(w *w65816) {
	w.indirectLong(w.SBC)
}

// INX (X=X+1)
func opE8(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if w.r.emulation || w.r.p.x {
			w.INC8(&w.r.x)
			return
		}

		w.INC16(&w.r.x)
	}
}

func opE9(w *w65816) {
	w.imm(w.SBC)
}

// NOP
func opEA(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {}
}

// XBA (Swap low and high nibble in A)
func opEB(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			w.r.a = (w.r.a >> 8) | (w.r.a << 8)
			w.r.p.setFlags(zn(w.r.a, 8))
		}
	}
}

func opEC(w *w65816) {
	w.absolute(w.CMP(&w.r.x))
}

func opED(w *w65816) {
	w.absolute(w.SBC)
}

// INC nnnn
func opEE(w *w65816) {
	w.absolute(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val++
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})
			return
		}

		w.read16(addr, func(val uint16) {
			val++
			w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
		})
	})
}

func opEF(w *w65816) {
	w.absoluteLong(w.SBC)
}

func opF0(w *w65816) {
	w.JMP8(func() bool { return w.r.p.z })
}

func opF1(w *w65816) {
	w.indirectY(w.SBC)
}

func opF2(w *w65816) {
	w.indirect(w.SBC)
}

func opF3(w *w65816) {
	w.stackRelativeY(w.SBC)
}

// PEA
func opF4(w *w65816) {
	w.imm16(func(nnnn uint16) {
		w.PUSH16(nnnn, func(w *w65816) {})
	})
}

func opF5(w *w65816) {
	w.zeropageX(w.SBC)
}

// 0xF6:
//
//	Effect: `[D+nn+X]=[D+nn+X]+1`
//	Native: `INC nn,X`
//	Nocash: `INC [nn+X]`
func opF6(w *w65816) {
	w.zeropageX(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val++
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})
			return
		}

		// 3,3a
		w.read16(addr, func(val uint16) {
			val += 1
			w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
		})
	})
}

func opF7(w *w65816) {
	w.indirectLongY(w.SBC)
}

// SED (D=1)
func opF8(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) { w.r.p.d = true }
}

func opF9(w *w65816) {
	w.absoluteY(w.SBC)
}

func opFA(w *w65816) {
	w.POP(&w.r.x)
}

// XCE (C=E, E=C)
func opFB(w *w65816) {
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		c, e := w.r.p.c, w.r.emulation
		w.r.p.c = e
		w.r.setEmulation(c)
	}
}

// JSR
func opFC(w *w65816) {
	// 2(AAL)
	w.imm8(func(lo uint8) {
		// 3, 4
		w.PUSH16(w.r.pc.offset, func(w *w65816) {
			// 5
			w.imm8(func(hi uint8) {
				// 6
				w.state = CPU_DUMMY_READ
				w.inst = func(w *w65816) {
					nnnn := uint16(hi)<<8 | uint16(lo)
					addr := u24(w.r.pc.bank, nnnn).plus(int(w.r.x))
					w.read16(addr, func(pc uint16) {
						w.r.pc = u24(w.r.pc.bank, pc)
					})
				}
			})
		})
	})
}

func opFD(w *w65816) {
	w.absoluteX(w.SBC)
}

// 0xFE:
//
//	Effect: `[DB:nnnn+X]=[DB:nnnn+X]+1`
//	Native: `INC nnnn,X`
//	Nocash: `INC [nnnn+X]`
func opFE(w *w65816) {
	w.absoluteX(func(addr uint24) {
		if w.r.emulation || w.r.p.m {
			w.read8(addr, func(val uint8) {
				val++
				w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
			})
			return
		}

		w.read16(addr, func(val uint16) {
			val++
			w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
		})
	})
}

func opFF(w *w65816) {
	w.absoluteLongX(w.SBC)
}
