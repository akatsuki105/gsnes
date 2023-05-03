package core

func (w *w65816) dummyRead(inst opcode) {
	w.state = CPU_DUMMY_READ
	w.inst = inst
}

func (w *w65816) exception(e exception) {
	vec := w.r.vector(e)

	switch e {
	case BRK, COP:
		fn := func(w *w65816) {
			// 4,5
			w.PUSH16(w.r.pc.offset+1, func(w *w65816) {
				// 6
				w.PUSH8(w.r.p.pack(), func(w *w65816) {
					// 7, 8
					w.read16(vec, func(pc uint16) {
						w.r.p.setFlags(flag{'d', false}, flag{'i', true})
						w.r.pc = u24(0x00, pc)
					})
				})
			})
		}

		if w.r.emulation {
			w.r.p.setFlags(flag{'x', true})
			fn(w)
		} else {
			w.PUSH8(w.r.pc.bank, fn) // 3
		}
	default:
		crash("invalid exception")
	}
}

func (w *w65816) AND(addr uint24) {
	and8 := func(val uint8) {
		w.r.l(&w.r.a, uint8(w.r.a)&val)
		w.r.p.setFlags(zn(w.r.a, 8))
	}
	and16 := func(val uint16) {
		w.r.a &= val
		w.r.p.setFlags(zn(w.r.a, 16))
	}

	if w.r.emulation || w.r.p.m {
		w.read8(addr, and8)
		return
	}

	w.read16(addr, and16)
}

// Cycle: 1 or 2(E=0)
func (w *w65816) ORA(addr uint24) {
	if w.r.emulation || w.r.p.m {
		ora8 := func(val uint8) {
			w.r.a |= uint16(val)
			w.r.p.setFlags(zn(w.r.a, 8))
		}

		// 1
		w.read8(addr, ora8)
		return
	}

	ora16 := func(val uint16) {
		w.r.a |= val
		w.r.p.setFlags(zn(w.r.a, 16))
	}

	// 2
	w.read16(addr, ora16)
}

// Cycle: 1 or 2(E=0)
func (w *w65816) LDNm(r *uint16) func(addr uint24) {
	p := &w.r.p.x
	if r == &w.r.a {
		p = &w.r.p.m
	}

	return func(addr uint24) {
		if w.r.emulation || *p {
			// 1
			w.read8(addr, func(val uint8) {
				w.r.l(r, val)
				w.r.p.setFlags(zn(*r, 8))
			})
			return
		}

		// 2
		w.read16(addr, func(val uint16) {
			*r = val
			w.r.p.setFlags(zn(*r, 16))
		})
	}
}

// STZ, STA, STX, STY
func (w *w65816) STN(r *uint16) func(addr uint24) {
	p := &w.r.p.x
	if r == nil || r == &w.r.a {
		p = &w.r.p.m
	}

	return func(addr uint24) {
		val := uint16(0)
		if r != nil {
			val = *r
		}

		if w.r.emulation || *p {
			w.write8(addr, uint8(val), nil)
			return
		}

		w.write16(addr, val, nil)
	}
}

// Cycle: 1
func (w *w65816) MOV(dst, src *uint16) {
	is16bit := false

	p := &w.r.p.x
	switch dst {
	case &w.r.a:
		p = &w.r.p.m
		switch src {
		case &w.r.d, &w.r.s:
			is16bit = true
		}
	case &w.r.s:
		p = &w.r.emulation
	case &w.r.d:
		is16bit = true
	}

	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		if !is16bit && (w.r.emulation || *p) {
			w.r.l(dst, uint8(*src))
			if src != &w.r.s && dst != &w.r.s {
				w.r.p.setFlags(zn(*dst, 8))
			}
			return
		}

		// 16bit
		*dst = *src
		if src != &w.r.s && dst != &w.r.s {
			w.r.p.setFlags(zn(*dst, 16))
		}
	}
}

func (w *w65816) INC8(r *uint16) {
	w.r.l(r, uint8(*r)+1)
	w.r.p.setFlags(zn(*r, 8))
}

func (w *w65816) INC16(r *uint16) {
	*r = *r + 1
	w.r.p.setFlags(zn(*r, 16))
}

func (w *w65816) DEC8(r *uint16) {
	w.r.l(r, uint8(*r)-1)
	w.r.p.setFlags(zn(*r, 8))
}

func (w *w65816) DEC16(r *uint16) {
	*r = *r - 1
	w.r.p.setFlags(zn(*r, 16))
}

// PHA, PHX, PHY, PHD
func (w *w65816) PUSH(r *uint16) {
	// 2
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {
		is16Bit := false
		p := &w.r.p.x
		switch r {
		case &w.r.a:
			p = &w.r.p.m
		case &w.r.d:
			is16Bit = true
		}

		if !is16Bit && (w.r.emulation || *p) {
			// 3
			w.PUSH8(uint8(*r), func(w *w65816) {})
			return
		}
		// 3, 3a
		w.PUSH16(*r, func(w *w65816) {})
	}
}

func (w *w65816) PUSH8(val uint8, fn opcode) {
	w.bus.data = val
	w.state = CPU_MEMORY_STORE
	w.bus.addr = u24(0, w.r.s)
	w.r.s--
	w.inst = fn
}

func (w *w65816) PUSH16(val uint16, fn opcode) {
	hi, lo := uint8(val>>8), uint8(val)
	w.PUSH8(hi, func(w *w65816) {
		w.PUSH8(lo, fn)
	})
}

// PLA, PLX, PLY, PLD
func (w *w65816) POP(r *uint16) {
	is16Bit := false
	p := &w.r.p.x
	switch r {
	case &w.r.a:
		p = &w.r.p.m
	case &w.r.d:
		is16Bit = true
	}

	// 2(IO)
	w.state = CPU_DUMMY_READ
	w.inst = func(w *w65816) {

		// 3(IO)
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			// 4(REG Low)
			w.POP8(func(lo uint8) {
				if !is16Bit && (w.r.emulation || *p) {
					w.r.l(r, lo)
					w.r.p.setFlags(zn(*r, 8))
					return
				}

				// 4a(REG High)
				w.POP8(func(hi uint8) {
					*r = uint16(hi)<<8 | uint16(lo)
					w.r.p.setFlags(zn(*r, 16))
				})
			})
		}
	}
}

func (w *w65816) POP8(fn func(val uint8)) {
	w.r.s++
	w.read8(u24(0, w.r.s), fn)
}

func (w *w65816) POP16(fn func(val uint16)) {
	w.POP8(func(lo uint8) {
		w.POP8(func(hi uint8) {
			val := uint16(hi)<<8 | uint16(lo)
			fn(val)
		})
	})
}

func (w *w65816) BIT(addr uint24) {
	if w.r.emulation || w.r.p.m {
		bit8 := func(val uint8) {
			v := val&0x40 != 0
			n := val&0x80 != 0
			z := val&uint8(w.r.a) == 0
			w.r.p.setFlags(flag{'v', v}, flag{'n', n}, flag{'z', z})
		}

		w.read8(addr, bit8)
		return
	}

	bit16 := func(val uint16) {
		v := val&0x4000 != 0
		n := val&0x8000 != 0
		z := val&w.r.a == 0
		w.r.p.setFlags(flag{'v', v}, flag{'n', n}, flag{'z', z})
	}

	w.read16(addr, bit16)
}

// CMP, CPX, CPY
func (w *w65816) CMP(r *uint16) func(addr uint24) {
	p := &w.r.p.x
	if r == &w.r.a {
		p = &w.r.p.m
	}

	return func(addr uint24) {
		if w.r.emulation || *p {
			cmp8 := func(op uint8) {
				i16 := int16(*r&0xFF) - int16(op)
				c := i16 >= 0
				z, n := zn(uint16(uint8(i16)), 8)
				w.r.p.setFlags(z, n, flag{'c', c})
			}

			w.read8(addr, cmp8)
			return
		}

		cmp16 := func(op uint16) {
			i32 := int32(*r) - int32(op)
			c := i32 >= 0
			z, n := zn(uint16(i32), 16)
			w.r.p.setFlags(z, n, flag{'c', c})
		}
		w.read16(addr, cmp16)
	}
}

func (w *w65816) ADC(addr uint24) {
	if w.r.emulation || w.r.p.m {
		adc8 := func(val uint8) {
			if w.r.p.d {
				result := (w.r.a & 0x0F) + uint16(val&0x0F) + w.carry()
				if result > 0x09 {
					result += 0x06
				}
				carry := btou16(result > 0x0F)
				result = (w.r.a & 0xF0) + uint16(val&0xF0) + (result & 0x0F) + carry*0x10

				v := uint8(w.r.a)&0x80 == (val&0x80) && (w.r.a&0x80) != (result&0x80)

				if result > 0x9F {
					result += 0x60
				}

				c := result > 0xFF

				w.r.l(&w.r.a, uint8(result))
				z, n := zn(w.r.a, 8)
				w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})
				return
			}

			u16 := (w.r.a & 0xFF) + uint16(val) + w.carry()
			c := u16 >= 0x100
			v := (^(uint8(w.r.a) ^ val) & (val ^ uint8(u16)) & 0x80) != 0
			w.r.l(&w.r.a, uint8(u16))

			z, n := zn(w.r.a, 8)
			w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})
		}

		w.read8(addr, adc8)
		return
	}

	adc16 := func(val uint16) {
		if w.r.p.d {
			a := uint32(w.r.a)
			result := (a & 0x000F) + uint32(val&0x000F) + uint32(w.carry())
			if result > 0x0009 {
				result += 0x0006
			}
			carry := btou32(result > 0x000F)

			result = (a & 0x00F0) + uint32(val&0x00F0) + (result & 0x000F) + carry*0x10
			if result > 0x009F {
				result += 0x0060
			}
			carry = btou32(result > 0x00FF)

			result = (a & 0x0F00) + uint32(val&0x0F00) + (result & 0x00FF) + carry*0x100
			if result > 0x09FF {
				result += 0x0600
			}
			carry = btou32(result > 0x0FFF)

			result = (a & 0xF000) + uint32(val&0xF000) + (result & 0x0FFF) + carry*0x1000

			v := w.r.a&0x8000 == (val&0x8000) && (a&0x8000) != (result&0x8000)

			if result > 0x9FFF {
				result += 0x6000
			}

			c := result > 0xFFFF

			w.r.a = uint16(result)
			z, n := zn(w.r.a, 16)
			w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})
			return
		}

		u32 := uint32(w.r.a) + uint32(val) + uint32(w.carry())
		c := u32 >= 0x10000
		v := (^(w.r.a ^ val) & (val ^ uint16(u32)) & 0x8000) != 0
		w.r.a = uint16(u32)

		z, n := zn(w.r.a, 16)
		w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})
	}
	w.read16(addr, adc16)
}

func (w *w65816) SBC(addr uint24) {
	if w.r.emulation || w.r.p.m {
		sbc8 := func(val uint8) {
			if w.r.p.d {
				val ^= 0xFF
				a := uint8(w.r.a)

				result := int((a & 0x0F) + (val & 0x0F) + uint8(w.carry()))
				if result < 0x10 {
					result -= 0x06
				}
				carry := btou8(result > 0x0F)

				result = int(a&0xF0) + int(val&0xF0) + (result & 0x0F) + int(carry*0x10)

				v := a&0x80 == val&0x80 && a&0x80 != uint8(result)&0x80

				if result < 0x100 {
					result -= 0x60
				}

				c := result > 0xFF

				w.r.l(&w.r.a, uint8(result))
				z, n := zn(w.r.a, 8)
				w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})

				return
			}

			carry := int16(w.carry())
			i16 := int16(w.r.a&0xFF) + carry - int16(val) - 1

			c := i16 >= 0
			v := ((uint8(w.r.a) ^ val) & (uint8(w.r.a) ^ uint8(i16)) & 0x80) != 0

			w.r.l(&w.r.a, uint8(i16))
			z, n := zn(w.r.a, 8)

			w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})
		}

		w.read8(addr, sbc8)
		return
	}

	sbc16 := func(val uint16) {
		if w.r.p.d {
			val ^= 0xFFFF
			a := int(w.r.a)

			carry := int(w.carry())
			result := 0
			c, v := false, false
			for i := 0; i < 4; i++ {
				d := 0xF << (4 * i)
				carry *= (0x1 << (4 * i))
				max := (0x1 << (4 * (i + 1))) - 1

				result = (a & d) + (int(val) & d) + (result & (max >> 4)) + carry
				v = (a^int(val)&0x8000) == 0 && ((a^result)&0x8000) != 0
				if result < (0x10 << (4 * i)) {
					result -= (0x6 << (4 * i))
				}
				c = result > max
				carry = btoi(c)
			}
			w.r.a = uint16(result)
			z, n := zn(w.r.a, 16)
			w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})

			return
		}

		carry := int32(w.carry())
		i32 := int32(w.r.a) + carry - int32(val) - 1

		c := i32 >= 0
		v := ((w.r.a ^ val) & (w.r.a ^ uint16(i32)) & 0x8000) != 0

		w.r.a = uint16(i32)
		z, n := zn(w.r.a, 16)

		w.r.p.setFlags(z, n, flag{'c', c}, flag{'v', v})
	}
	w.read16(addr, sbc16)
}

func (w *w65816) TSB(addr uint24) {
	if w.r.emulation || w.r.p.m {
		w.read8(addr, func(val uint8) {
			w.r.p.setFlags(flag{'z', uint8(w.r.a)&val == 0})
			val |= uint8(w.r.a)
			w.write8(addr, val, nil)
		})
		return
	}

	w.read16(addr, func(val uint16) {
		w.r.p.setFlags(flag{'z', w.r.a&val == 0})
		val |= w.r.a
		w.write16(addr, val, nil)
	})
}

func (w *w65816) TRB(addr uint24) {
	if w.r.emulation || w.r.p.m {
		w.read8(addr, func(val uint8) {
			w.r.p.setFlags(flag{'z', uint8(w.r.a)&val == 0})
			val &= ^uint8(w.r.a)
			w.write8(addr, val, nil)
		})
		return
	}

	w.read16(addr, func(val uint16) {
		w.r.p.setFlags(flag{'z', w.r.a&val == 0})
		val &= ^w.r.a
		w.write16(addr, val, nil)
	})
}

// If cond is ok, jump +/- dips8
func (w *w65816) JMP8(cond func() bool) {
	// 2(Offset)
	w.imm8(func(val uint8) {
		dd := int(int8(val))
		if !cond() {
			return
		}

		// 3(IO)
		pc := uint16(int(w.r.pc.offset) + dd)
		w.state = CPU_DUMMY_READ
		w.inst = func(w *w65816) {
			fn := func(w *w65816) { w.r.pc = u24(w.r.pc.bank, pc) }

			// 4(IO)
			// Add 1 cycle if branch is taken across page boundaries in 6502 emulation mode (E=1).
			if w.r.emulation && (w.r.pc.offset>>8 != pc>>8) {
				w.state = CPU_DUMMY_READ
				w.inst = fn
				return
			}
			fn(w)
		}
	})
}

// 0x66, 0x76, 0x6E, 0x7E
// Rotate right with carry. (C -> MSB, LSB -> C)
func (w *w65816) ROR(addr uint24) {
	c := w.carry()

	if w.r.emulation || w.r.p.m {
		w.read8(addr, func(val uint8) {
			addCycle(w.cycles, FAST)
			u16 := uint16(val) | (uint16(c) << 8)
			c := u16&0b1 != 0
			u16 >>= 1
			w.write8(addr, uint8(u16), func() {
				z, n := zn(u16, 8)
				w.r.p.setFlags(z, n, flag{'c', c})
			})
		})

		return
	}

	w.read16(addr, func(val uint16) {
		addCycle(w.cycles, FAST)
		u32 := uint32(val) | (uint32(c) << 16)
		c := u32&0b1 != 0
		u32 >>= 1
		w.write16(addr, uint16(u32), func() {
			z, n := zn(uint16(u32), 16)
			w.r.p.setFlags(z, n, flag{'c', c})
		})
	})
}

// 0x26, 0x36, 0x2E, 0x3E
// Rotate left with carry. (C -> LSB, MSB -> C)
func (w *w65816) ROL(addr uint24) {
	c := w.r.p.c

	if w.r.emulation || w.r.p.m {
		w.read8(addr, func(val uint8) {
			addCycle(w.cycles, FAST)
			w.r.p.c = bit(val, 7)
			val <<= 1
			val = setBit(val, 0, c)
			w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
		})

		return
	}

	w.read16(addr, func(val uint16) {
		addCycle(w.cycles, FAST)
		w.r.p.c = bit(val, 15)
		val <<= 1
		val = setBit(val, 0, c)
		w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
	})
}

// 0x46, 0x4E, 0x56, 0x5E
func (w *w65816) LSR(addr uint24) {
	if w.r.emulation || w.r.p.m {
		w.read8(addr, func(val uint8) {
			addCycle(w.cycles, FAST)
			w.r.p.c = bit(val, 0)
			val >>= 1
			w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
		})
		return
	}

	w.read16(addr, func(val uint16) {
		addCycle(w.cycles, FAST)
		w.r.p.c = bit(val, 0)
		val >>= 1
		w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
	})
}

func (w *w65816) LSR8(r *uint16) {
	msb := uint8((*r & 0b1) << 7)
	val := uint8(*r) >> 1
	w.r.l(r, val)
	w.r.p.setFlags(flag{'z', val == 0}, flag{'n', false}, flag{'c', msb != 0})
}

func (w *w65816) LSR16(r *uint16) {
	msb := *r << 15
	*r = (*r >> 1)
	w.r.p.setFlags(flag{'z', *r == 0}, flag{'n', false}, flag{'c', msb != 0})
}

// Shift One Bit Left
// 0x06, 0x0E, 0x16, 0x1E
func (w *w65816) ASL(addr uint24) {
	if w.r.emulation || w.r.p.m {
		w.read8(addr, func(val uint8) {
			addCycle(w.cycles, FAST)
			w.r.p.c = bit(val, 7)
			val <<= 1
			w.write8(addr, val, func() { w.r.p.setFlags(zn(uint16(val), 8)) })
		})
		return
	}

	w.read16(addr, func(val uint16) {
		addCycle(w.cycles, FAST)
		w.r.p.c = bit(val, 15)
		val <<= 1
		w.write16(addr, val, func() { w.r.p.setFlags(zn(val, 16)) })
	})
}

// a.k.a XOR
func (w *w65816) EOR(addr uint24) {
	if w.r.emulation || w.r.p.m {
		eor8 := func(val uint8) {
			w.r.a ^= uint16(val)
			w.r.p.setFlags(zn(w.r.a, 8))
		}

		w.read8(addr, eor8)
		return
	}

	eor16 := func(val uint16) {
		w.r.a ^= val
		w.r.p.setFlags(zn(w.r.a, 16))
	}

	w.read16(addr, eor16)
}
