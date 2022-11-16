package core

type maskLogic uint8

const (
	OR maskLogic = iota
	AND
	XOR
	XNOR
)

type windowSystem struct {
	win1, win2 window
	logic      [6]maskLogic // WBGLOG/WOBJLOG
}

// WxxSEL
type maskSetting struct {
	enable bool // bit0
	invert bool // bit1
}

func windowMask(val uint8) maskSetting {
	return maskSetting{
		enable: bit(val, 0),
		invert: bit(val, 1),
	}
}

type window struct {
	left, right uint8
	mask        [6]maskSetting // WxxSEL
}
