package cartridge

import "fmt"

type destination uint8

func (d destination) String() string {
	switch d {
	case destination(0x00):
		return "J (Japan)"
	case destination(0x01):
		return "E (America, Canada)"
	case destination(0x02):
		return "P (Europe)"
	case destination(0x03):
		return "W (Sweden, Scandinavia)"
	case destination(0x04):
		return "- (Finland)"
	case destination(0x05):
		return "- (Danmark)"
	case destination(0x06):
		return "F (France)"
	case destination(0x07):
		return "H (Nederland)"
	case destination(0x08):
		return "S (Spain)"
	case destination(0x09):
		return "D (Germany)"
	case destination(0x0a):
		return "I (Italy)"
	case destination(0x0b):
		return "C (China)"
	case destination(0x0c):
		return "- (Indonesia)"
	case destination(0x0d):
		return "K (Korea)"
	case destination(0x0f):
		return "N (Canada)"
	case destination(0x10):
		return "B (Brazil)"
	}
	return fmt.Sprintf("Unknown(%d)", int(d))
}
