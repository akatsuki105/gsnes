package core

import (
	a "github.com/akatsuki105/gsnes/core/apu"
)

type apu struct {
	a.APU
	cycles int64 // Clock is 21.47727MHz
}

func newApu() *apu {
	a := &apu{
		APU: a.New(),
	}
	return a
}

func (a *apu) reset() {
	a.APU.Reset()
}

// addr: 0 or 1 or 2 or 3
func (a *apu) readIO(addr uint, defaultVal uint8) uint8 {
	a.catchup()
	return a.Read(int(addr))
}

// addr: 0 or 1 or 2 or 3
func (a *apu) writeIO(addr uint, val uint8) {
	a.catchup()
	a.Write(int(addr), val)
}

func (a *apu) catchup() {
	apuCycles := int(a.cycles) / 21
	for i := 0; i < apuCycles; i++ {
		a.APU.Cycle()
		a.cycles -= 21
	}
}
