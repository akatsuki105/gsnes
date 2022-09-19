package cartridge

import (
	"fmt"
	"strings"
)

// FFD6h
type Chipset struct {
	Val   uint8
	lists []string
}

func chipset(val uint8) *Chipset {
	c := &Chipset{
		Val:   val,
		lists: []string{"Unknown"},
	}

	switch val {
	case 0x00:
		c.lists = []string{"ROM"}
	case 0x01:
		c.lists = []string{"ROM", "RAM"}
	case 0x02:
		c.lists = []string{"ROM", "RAM", "Battery"}
	case 0x03:
		c.lists = []string{"ROM", "DSP"}
	case 0x04:
		c.lists = []string{"ROM", "DSP", "RAM"}
	case 0x05:
		c.lists = []string{"ROM", "DSP", "RAM", "Battery"}
	case 0x14:
		c.lists = []string{"ROM", "GSU", "RAM"}
	case 0x15:
		c.lists = []string{"ROM", "GSU", "RAM", "Battery"}
	case 0x25:
		c.lists = []string{"ROM", "OBC1", "RAM", "Battery"}
	case 0x32, 0x35:
		c.lists = []string{"ROM", "SA1", "RAM", "Battery"}
	case 0x34:
		c.lists = []string{"ROM", "SA1", "RAM"}
	case 0x43:
		c.lists = []string{"ROM", "S-DD1"}
	case 0x45:
		c.lists = []string{"ROM", "S-DD1", "RAM", "Battery"}
	case 0x55:
		c.lists = []string{"ROM", "S-RTC", "RAM", "Battery"}
	}

	return c
}

func (c *Chipset) String() string {
	s := fmt.Sprintf("%02Xh (%s)", c.Val, strings.Join(c.lists, "+"))
	return s
}
