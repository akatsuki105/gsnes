package core

import (
	"fmt"
)

type w65816 struct {
	c            *sfc
	r            reg
	lastInstAddr uint24 // cur instruction ($)
	state        cpustate
	inst         opcode
	bus          bus
	cycles       *int64
	nextEvent    *int64
	halted       bool // by WAI
	wram
	cart *cartridge

	nmiPending              bool // internal NMI flag
	nmitimen, rdnmi, timeup uint8
	wrio                    uint8
	ajr                     bool  // HVBJOY.0
	ws2                     int64 // MEMSEL

	joypads [4]joypad

	lock int8 // CPU blocked by DMA and DRAM refresh

	bkpts breakpoints

	mul struct {
		a      uint8  // WRMPYA
		result uint16 // RDMPY
	}
	div struct {
		dividend uint16 // WRDIV(L/H)
		result   uint16 // RDDIV
	}
}

type joypad struct {
	idx int // 1,2,3,4
	buf [12]bool
	val uint16
}

type bus struct {
	addr uint24
	data uint8
}

func new65816(c *sfc, cycles, nextEvent *int64) *w65816 {
	if c.ppu == nil {
		crash("PPU must be initialized befor CPU")
	}

	w := &w65816{
		c:         c,
		cycles:    cycles,
		nextEvent: nextEvent,
		cart:      newCartridge(c),
		bkpts:     *newBreakpoints(c),
		wram:      *newWram(),
		joypads:   [4]joypad{{idx: 1}, {idx: 2}, {idx: 3}, {idx: 4}},
		ws2:       MEDIUM,
		rdnmi:     2, // CPU Version
	}
	w.r.p.r = &w.r
	return w
}

func (w *w65816) reset() {
	addCycle(w.cycles, INIT_CYCLE)
	w.lastInstAddr = u24(0, 0)
	w.r.reset()
	entry := w.load16(w.r.vector(RESET), nil)
	w.r.pc = u24(0, entry)
	w.wram.reset()
	w.state = CPU_FETCH
	w.halted = false
	w.lock = 0
}

func (w *w65816) step() (running bool) {
	prev := *w.cycles

	switch w.state {
	case CPU_FETCH:
		w.lastInstAddr = w.r.pc
		addCycle(w.cycles, w.wait(w.r.pc))
		opcode := w.load8(w.r.pc)
		pushHistory(opcode, w.r.pc)

		if pc := w.lastInstAddr.u32(); w.bkpts.shouldBreak(pc) {
			for i := range histories {
				fmt.Println(i, histories[i])
			}
			w.c.pause = true
			return false
		}

		w.r.pc.offset++
		w.inst = opTable[opcode]

	case CPU_READ_PC:
		addCycle(w.cycles, w.wait(w.r.pc))
		val := w.load8(w.r.pc)
		w.bus.data = val
		w.r.pc.offset++

	case CPU_DUMMY_READ:
		addCycle(w.cycles, FAST)

	case CPU_MEMORY_LOAD:
		addCycle(w.cycles, w.wait(w.bus.addr))
		w.bus.data = w.load8(w.bus.addr)

	case CPU_MEMORY_STORE:
		w.store8(w.bus.addr, w.bus.data, w.cycles)
	}

	w.state = CPU_FETCH
	w.inst(w)

	w.c.apu.cycles += toApuCycles(*w.cycles - prev)
	return true
}

// check NMI and IRQ
func (w *w65816) checkIrq(e exception) (interrupted bool) {
	switch e {
	case NMI:
		if !w.nmiPending {
			return false
		}
		w.nmiPending = false

	case IRQ:
		requested := bit(w.timeup, 7)
		if requested {
			w.halted = false
		}

		if w.r.p.i || !requested {
			return false
		}
	}

	fn := func(w *w65816) {
		w.PUSH16(w.r.pc.offset, func(w *w65816) {
			w.PUSH8(w.r.p.pack(), func(w *w65816) {
				w.r.p.d, w.r.p.i = false, true
				vec := w.r.vector(e)
				w.read16(vec, func(offset uint16) {
					w.r.pc = u24(0, offset)
				})
			})
		})
	}

	w.halted = false
	if w.r.emulation {
		w.r.p.x = false
		fn(w)
	} else {
		w.PUSH8(w.r.pc.bank, fn)
	}
	return true
}

func (w *w65816) wait(addr uint24) int64 {
	addr32 := addr.u32()
	bank := addr.bank

	// 00-3f,80-bf:8000-ffff; 40-7f,c0-ff:0000-ffff
	if (addr32 & 0x408000) != 0 {
		if bank&0x80 != 0 {
			return w.ws2
		}
		return MEDIUM
	}

	// 00-3f,80-bf:0000-1fff,6000-7fff
	if (addr32+0x6000)&0x4000 != 0 {
		return MEDIUM
	}

	if (addr32-0x4000)&0x7e00 != 0 {
		return FAST
	}

	// 00-3f,80-bf:4000-41ff
	return SLOW
}

// Load a byte from memory.
func (w *w65816) load8(addr uint24) uint8 {
	m := w.c.m
	m.before = uint(addr.u32())
	return m.reader[m.lookup[addr.u32()]](m.target[addr.u32()], w.bus.data)
}

// Load 2 bytes from memory as Little endian.
func (w *w65816) load16(addr uint24, cycles *int64) uint16 {
	addCycle(cycles, w.wait(addr))
	lo := uint16(w.load8(addr))

	addr = addr.plus(1)
	addCycle(cycles, w.wait(addr))
	hi := uint16(w.load8(addr))

	return (hi << 8) | lo
}

func (w *w65816) store8(addr uint24, val uint8, cycles *int64) {
	addCycle(cycles, w.wait(addr))
	m := w.c.m
	m.before = uint(addr.u32())
	m.writer[m.lookup[addr.u32()]](m.target[addr.u32()], val)
}

func (w *w65816) read8(addr uint24, fn func(uint8)) {
	if addr == READ_PC {
		w.imm8(fn)
		return
	}

	w.bus.addr = addr
	w.state = CPU_MEMORY_LOAD
	w.inst = func(w *w65816) {
		fn(w.bus.data)
	}
}

func (w *w65816) read16(addr uint24, fn func(val uint16)) {
	if addr == READ_PC {
		w.imm16(fn)
		return
	}

	w.bus.addr = addr
	w.state = CPU_MEMORY_LOAD
	w.inst = func(w *w65816) {
		lo := w.bus.data

		w.bus.addr = addr.plus(1)
		w.state = CPU_MEMORY_LOAD
		w.inst = func(w *w65816) {
			hi := w.bus.data
			fn(uint16(hi)<<8 | uint16(lo))
		}
	}
}

func (w *w65816) write8(addr uint24, val uint8, fn func()) {
	w.bus.addr, w.bus.data = addr, val
	w.state = CPU_MEMORY_STORE
	w.inst = func(w *w65816) {
		if fn != nil {
			fn()
		}
	}
}

func (w *w65816) write16(addr uint24, val uint16, fn func()) {
	lo, hi := uint8(val), uint8(val>>8)
	w.write8(addr, lo, func() {
		w.write8(addr.plus(1), hi, fn)
	})
}

func (w *w65816) carry() uint16 {
	if w.r.p.c {
		return 1
	}
	return 0
}

// nサイクルだけCPUを停止する(Haltではない)
func (w *w65816) block(block int8, n, cyclesLate int64) {
	val := int8(1) << block
	w.lock += val
	schedule("Unblock", w.c.s, func(_ int64) { w.lock -= val }, n-cyclesLate)
}

func (w *w65816) blocked() bool {
	if w.lock < 0 {
		crash("w.lock cannot be negative")
	}
	return w.lock > 0
}

func (w *w65816) Status() string {
	r := &w.r
	return fmt.Sprintf(`A: %04X,  X: %04X,  Y: %04X
S: %04X,  D: %04X, DB: %04X, PC: %v
P: %s, E: %v
Halt: %v, Blocked: %v`,
		r.a, r.x, r.y,
		r.s, r.d, r.db, r.pc,
		&r.p, r.emulation,
		w.halted, w.blocked())
}
