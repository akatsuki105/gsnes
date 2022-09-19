package core

// TODO: 開発用

type breakpoints struct {
	c     *sfc
	slice []uint32
	last  uint32
}

func newBreakpoints(c *sfc, addrs ...uint32) *breakpoints {
	return &breakpoints{
		c:     c,
		slice: addrs,
		last:  uint32(0xffffffff),
	}
}

func (b *breakpoints) shouldBreak(pc uint32) (isBreak bool) {
	defer func() {
		b.last = pc
	}()
	for _, addr := range b.slice {
		if pc == addr {
			isBreak = true
			break
		}
	}

	isBreak = isBreak && pc != b.last
	return isBreak
}
