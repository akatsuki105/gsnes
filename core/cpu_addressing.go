package core

func (w *w65816) imm(fn func(addr uint24)) {
	fn(READ_PC)
}

func (w *w65816) imm8(fn func(nn uint8)) {
	w.state = CPU_READ_PC
	w.inst = func(w *w65816) {
		fn(w.bus.data)
	}
}

// Read and increment 2 bytes from PC.
func (w *w65816) imm16(fn func(nnnn uint16)) {
	w.imm8(func(lo uint8) {
		w.imm8(func(hi uint8) {
			nnnn := uint16(hi)<<8 | uint16(lo)
			fn(nnnn)
		})
	})
}

func (w *w65816) imm24(fn func(addr uint24)) {
	w.imm16(func(ofs uint16) {
		w.imm8(func(bank uint8) {
			fn(u24(bank, ofs))
		})
	})
}

// ゼロページアドレッシング

// Zero Page:
//
//	Native:  `nn`
//	Nocash:  `[nn]`
//	Cycle:    1 (or 2)
//	Alias:   `Direct`, `d`
//	Example:  0x05
func (w *w65816) zeropage(fn func(addr uint24)) {
	// DO
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO
		}

		addr := u24(0, w.r.d).plus(int(nn)) // 00:(nn+D)
		fn(addr)
	})
}

// Zero Page,X:
//
//	Nocash: `[nn+X]`
//	Cycle:   1 (or 2)
//	Alias:  `Direct, X`
//	Example: 0x15
func (w *w65816) zeropageX(fn func(addr uint24)) {
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO
		}
		addCycle(w.cycles, FAST)

		addr := u24(0, w.r.d).plus(int(nn)).plus(int(w.r.x)) // 00:(nn+D+X)
		fn(addr)
	})
}

// Zero Page,Y:
//
//	Nocash: `[nn+Y]`
//	Cycle:   1 (or 2)
//	Alias:  `Direct, Y`, `d,y`
//	Example: 0xB6
func (w *w65816) zeropageY(fn func(addr uint24)) {
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO
		}
		addCycle(w.cycles, FAST)

		addr := u24(0, w.r.d).plus(int(nn)).plus(int(w.r.y)) // 00:(nn+D+Y)
		fn(addr)
	})
}

// 絶対アドレッシング

// Absolute:
//
//	Native:  `nnnn`
//	Nocash:  `[nnnn]`
//	Cycle:    2
//	Alias:   `a`
//	Example:  0x0D
func (w *w65816) absolute(fn func(addr uint24)) {
	// AAL, AAH
	w.imm16(func(nnnn uint16) {
		addr := u24(w.r.db, nnnn) // DB:nnnn
		fn(addr)
	})
}

// Absolute,X:
//
//	Native:  `nnnn,X`
//	Nocash:  `[nnnn+X]`
//	Alias:   `a,x`
//	Example:  0x1D
func (w *w65816) absoluteX(fn func(addr uint24), rw access) {
	// AAL, AAH
	w.imm16(func(nnnn uint16) {
		// penalty cycle when crossing 8-bit page boundaries:
		if rw == R {
			if !w.r.p.x || (nnnn>>8 != (nnnn+w.r.x)>>8) {
				addCycle(w.cycles, FAST) // 3a
			}
		}
		if rw == W || rw == M {
			addCycle(w.cycles, FAST) // 3a
		}

		addr := u24(w.r.db, nnnn).plus(int(w.r.x)) // DB:(nnnn+X)
		fn(addr)
	})
}

// Absolute,Y:
//
//	Nocash:  `[nnnn+Y]`
//	Alias:   `a,y`
//	Example:  0xB9
func (w *w65816) absoluteY(fn func(addr uint24), rw access) {
	w.imm16(func(nnnn uint16) {
		// penalty cycle when crossing 8-bit page boundaries:
		if rw == R {
			if !w.r.p.x || (nnnn>>8 != (nnnn+w.r.x)>>8) {
				addCycle(w.cycles, FAST) // 3a
			}
		}
		if rw == W || rw == M {
			addCycle(w.cycles, FAST) // 3a
		}

		addr := u24(w.r.db, nnnn).plus(int(w.r.y)) // DB:(nnnn+Y)
		fn(addr)
	})
}

// Absolute Long:
//
//	Nocash:  `[nnnnnn]`
//	Alias:   `al`
//	Example:  0x2F
func (w *w65816) absoluteLong(fn func(addr uint24)) {
	// AAL, AAH
	w.imm16(func(nnnn uint16) {
		// AAB
		w.imm8(func(nn uint8) {
			addr := u24(nn, nnnn) // nnnnnn
			fn(addr)
		})
	})
}

// Absolute Long,X:
//
//	Nocash:  `[nnnnnn+X]`
//	Alias:   `al,x`
//	Example:  0x3F
func (w *w65816) absoluteLongX(fn func(addr uint24)) {
	// AAL, AAH
	w.imm16(func(nnnn uint16) {
		// AAB
		w.imm8(func(nn uint8) {
			addr := u24(nn, nnnn).plus(int(w.r.x)) // nnnnnn+X
			fn(addr)
		})
	})
}

// 間接アドレッシング

// Indirect:
//
//	Nocash:  `(nn)`
//	Native:  `[[nn]]`
//	Alias:   `(d)`, `direct indirect`
//	Example:  0x32
func (w *w65816) indirect(fn func(addr uint24)) {
	// DO(2)
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO(2a)
		}

		// AAL, AAH
		addr := u24(0, w.r.d).plus(int(nn))
		w.read16(addr, func(ofs uint16) {
			addr := u24(w.r.db, ofs) // DB:(WORD[00:(nn+D)])
			fn(addr)
		})
	})
}

// (Indirect,X):
//
//	Nocash:  `[[nn+X]]`
//	Alias:   `(d,x)`, `direct indexed indirect`
//	Example:  0x01
func (w *w65816) indirectX(fn func(addr uint24)) {
	// DO(2)
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO(2a)
		}
		addCycle(w.cycles, FAST)

		// AAL, AAH
		addr := u24(0, w.r.d+uint16(nn)).plus(int(w.r.x))
		w.read16(addr, func(ofs uint16) {
			addr := u24(w.r.db, ofs) // DB:(WORD[00:(nn+D+X)])
			fn(addr)
		})
	})
}

// (Indirect),Y:
//
//	Native:  `(nn),Y`
//	Nocash:  `[[nn]+Y]`
//	Alias:   `(d),y`, `direct indirect indexed`
//	Example:  0x31
func (w *w65816) indirectY(fn func(addr uint24)) {
	// DO(2)
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO(2a)
		}

		// AAL, AAH
		addr := u24(0, w.r.d).plus(int(nn))
		w.read16(addr, func(ofs uint16) {
			addr := u24(w.r.db, ofs).plus(int(w.r.y)) // DB:(WORD[00:(nn+D)]) + Y
			fn(addr)
		})
	})
}

// Indirect Long:
//
//	Native:  `[nn]`
//	Nocash:  `[FAR[nn]]`
//	Alias:   `[d]`, `direct Indirect Long`
//	Example: 0x27
func (w *w65816) indirectLong(fn func(addr uint24)) {
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO(2a)
		}

		addr := u24(0, w.r.d+uint16(nn))
		w.read16(addr, func(ofs uint16) {
			w.bus.addr.offset++
			w.state = CPU_MEMORY_LOAD
			w.inst = func(w *w65816) {
				fn(u24(w.bus.data, ofs))
			}
		})
	})
}

// Direct Indirect Indexed Long:
//
//	Native:  `[nn],y`
//	Nocash:  `[FAR[nn]+Y]`
//	Alias:   `[d],y`, `Direct Indirect Indexed Long`
//	Example: 0x37
func (w *w65816) indirectLongY(fn func(addr uint24)) {
	w.imm8(func(nn uint8) {
		if (w.r.d & 0xFF) != 0 {
			addCycle(w.cycles, FAST) // IO(2a)
		}

		addr := u24(0, w.r.d+uint16(nn))
		w.read16(addr, func(ofs uint16) {
			w.bus.addr.offset++
			w.state = CPU_MEMORY_LOAD
			w.inst = func(w *w65816) {
				fn(u24(w.bus.data, ofs).plus(int(w.r.y)))
			}
		})
	})
}

// Stack Relative:
//
//	Native:  `nn,S`
//	Nocash:  `[nn+S]`
//	Alias:   `d,s`
//	Example:  0x23
func (w *w65816) stackRelative(fn func(addr uint24)) {
	w.imm8(func(nn uint8) {
		fn(u24(0, uint16(nn)).plus(int(w.r.s)))
	})
}

// Stack Relative Indirect Indexed:
//
//	Native:  `(nn,S),Y`
//	Nocash:  `[[nn+S]+Y]`
//	Alias:   `(d,s),y`
//	Example:  0x33
func (w *w65816) stackRelativeY(fn func(addr uint24)) {
	w.imm8(func(nn uint8) {
		addr := u24(0, uint16(nn)).plus(int(w.r.s))
		w.read16(addr, func(ofs uint16) {
			fn(u24(w.r.db, ofs).plus(int(w.r.y)))
		})
	})
}
