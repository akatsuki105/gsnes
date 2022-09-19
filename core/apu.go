package core

import (
	a "github.com/pokemium/gsnes/core/apu"
)

func toApuCycles(masterCycles int64) float64 {
	return float64(masterCycles*2) / 21
}

type apu struct {
	a.APU
	cycles float64
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
	cycles := int(a.cycles)
	for i := 0; i < cycles; i++ {
		a.APU.Cycle()
	}
	a.cycles -= float64(cycles)
}
