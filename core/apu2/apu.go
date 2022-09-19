package apu

import "github.com/pokemium/gsnes/core/scheduler"

type Apu interface {
	Reset()

	// Read 1 byte from APU. idx is 0,1,2,3.
	Read(port int) uint8

	// Write 1 byte into APU. idx is 0,1,2,3.
	Write(port int, val uint8)
}

func New() Apu {
	a := &apu{
		s: scheduler.New(),
	}
	a.c = newSpc700(a)

	a.io[0] = &test{a}
	a.io[1] = &control{a}
	a.io[2], a.io[3] = &dspaddr{a}, &dspdata{a}
	a.io[4], a.io[5], a.io[6], a.io[7] = &cpuio{a, 0}, &cpuio{a, 1}, &cpuio{a, 2}, &cpuio{a, 3}
	a.io[8], a.io[9] = &auxio{a, 4}, &auxio{a, 5}
	a.io[10], a.io[11], a.io[12] = &tdiv{}, &tdiv{}, &tdiv{}
	a.io[13], a.io[14], a.io[15] = &tout{}, &tout{}, &tout{}

	a.Reset()

	return a
}

type apu struct {
	c     *spc700
	ports [4]port
	io    [16]io8
	s     *scheduler.Scheduler
}

type port struct {
	fromApu, toApu uint8
}

func (a *apu) Reset() {
	for i := range a.ports {
		a.ports[i].fromApu = 0x00
		a.ports[i].toApu = 0x00
	}
}

func (a *apu) Read(port int) uint8 {
	return a.ports[port].fromApu
}

func (a *apu) Write(port int, val uint8) {
	a.ports[port].toApu = val
}
