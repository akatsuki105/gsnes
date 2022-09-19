package apu

// 0x00F0
type test struct {
	a *apu
}

func (r *test) Read() uint8 {
	return 0xFF
}

func (r *test) Write(val uint8) {}

// 0x00F1
type control struct {
	a *apu
}

func (r *control) Read() uint8 {
	return 0xFF
}

func (r *control) Write(val uint8) {}

// 0x00F2
type dspaddr struct {
	a *apu
}

func (r *dspaddr) Read() uint8 {
	return 0xFF
}

func (r *dspaddr) Write(val uint8) {}

// 0x00F3
type dspdata struct {
	a *apu
}

func (r *dspdata) Read() uint8 {
	return 0xFF
}

func (r *dspdata) Write(val uint8) {}

// 0x00F4,0x00F5,0x00F6,0x00F7
type cpuio struct {
	a    *apu
	port int // 0,1,2,3
}

func (r *cpuio) Read() uint8 {
	return r.a.ports[r.port].toApu
}

func (r *cpuio) Write(val uint8) {
	r.a.ports[r.port].fromApu = val
}

type auxio struct {
	a    *apu
	port int // 4 or 5
}

func (r *auxio) Read() uint8 {
	return 0xFF
}

func (r *auxio) Write(val uint8) {}

type tdiv struct {
}

func (r *tdiv) Read() uint8 {
	return 0xFF
}

func (r *tdiv) Write(val uint8) {}

type tout struct {
}

func (r *tout) Read() uint8 {
	return 0xFF
}

func (r *tout) Write(val uint8) {}
