package apu

type opcode = func(c *spc700)

var opTable = [256]opcode{
	opE4,
}

func opE4(c *spc700) {
	addCycle(c.cycles, 3)
	c.MOV(c.zeropage)
}

func (c *spc700) MOV(addressing func() uint16) {
	addr := addressing()
	c.r.a = c.load8(addr, nil)
	c.r.p.z, c.r.p.n = zn(c.r.a)
}

func (c *spc700) zeropage() uint16 {
	nn := c.load8(c.r.pc, nil)
	c.r.pc++
	return c.load16(uint16(nn), nil)
}
