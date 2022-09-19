package core

// TODO: 開発用

import "fmt"

type history struct {
	opcode uint8
	addr   uint24
}

func (h history) String() string {
	return fmt.Sprintf("%s(opcode: %02Xh)", h.addr, h.opcode)
}

var histories = [5]history{}

func pushHistory(opcode uint8, addr uint24) {
	for i := range histories {
		if i+1 <= len(histories)-1 {
			histories[i] = histories[i+1]
		}
	}
	histories[len(histories)-1] = history{
		opcode: opcode,
		addr:   addr,
	}
}
