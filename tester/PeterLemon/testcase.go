package main

import (
	"fmt"

	"github.com/akatsuki105/gsnes/tester"
)

var max = 0

var tests = [...]testcase{
	{
		name: "HELLO",
		fn:   testHelloWorld,
	},

	// Bank
	{
		name: "BANK/WRAM",
		fn:   testBank("WRAM"),
	},
	{
		name: "BANK/LoROMFastROM",
		fn:   testBank("LoROMFastROM"),
	},
	{
		name: "BANK/LoROMSlowROM",
		fn:   testBank("LoROMSlowROM"),
	},

	// CPU
	{
		name: "CPU/ADC",
		fn:   testCpu("ADC"),
	},
	{
		name: "CPU/AND",
		fn:   testCpu("AND"),
	},
	{
		name: "CPU/ASL",
		fn:   testCpu("ASL"),
	},
	{
		name: "CPU/BIT",
		fn:   testCpu("BIT"),
	},
	{
		name: "CPU/BRA",
		fn:   testCpu("BRA"),
	},
	{
		name: "CPU/CMP",
		fn:   testCpu("CMP"),
	},
	{
		name: "CPU/DEC",
		fn:   testCpu("DEC"),
	},
	{
		name: "CPU/EOR",
		fn:   testCpu("EOR"),
	},
	{
		name: "CPU/INC",
		fn:   testCpu("INC"),
	},
	{
		name: "CPU/JMP",
		fn:   testCpu("JMP"),
	},
	{
		name: "CPU/LDR",
		fn:   testCpu("LDR"),
	},
	{
		name: "CPU/LSR",
		fn:   testCpu("LSR"),
	},
	{
		name: "CPU/MOV",
		fn:   testCpu("MOV"),
	},
	{
		name: "CPU/MSC",
		fn:   testCpu("MSC"),
	},
	{
		name: "CPU/ORA",
		fn:   testCpu("ORA"),
	},
	{
		name: "CPU/PHL",
		fn:   testCpu("PHL"),
	},
	{
		name: "CPU/PSR",
		fn:   testCpu("PSR"),
	},
	{
		name: "CPU/RET",
		fn:   testCpu("RET"),
	},
	{
		name: "CPU/ROL",
		fn:   testCpu("ROL"),
	},
	{
		name: "CPU/ROR",
		fn:   testCpu("ROR"),
	},
	{
		name: "CPU/SBC",
		fn:   testCpu("SBC"),
	},
	{
		name: "CPU/STR",
		fn:   testCpu("STR"),
	},
	{
		name: "CPU/TRN",
		fn:   testCpu("TRN"),
	},

	// PPU/BGMAP
	{
		name: "BGMAP/2BPP/BG1",
		fn:   testBGMap("2BPP", "8x8BG1Map2BPP32x328PAL"),
	},
	{
		name: "BGMAP/2BPP/BG2",
		fn:   testBGMap("2BPP", "8x8BG2Map2BPP32x328PAL"),
	},
	{
		name: "BGMAP/2BPP/BG3",
		fn:   testBGMap("2BPP", "8x8BG3Map2BPP32x328PAL"),
	},
	{
		name: "BGMAP/2BPP/BG4",
		fn:   testBGMap("2BPP", "8x8BG4Map2BPP32x328PAL"),
	},
	{
		name: "BGMAP/4BPP",
		fn:   testBGMap("4BPP", "8x8BGMap4BPP32x328PAL"),
	},
}

// 結果表示で列をそろえるためのパディング用
func init() {
	for i := range tests {
		length := len(tests[i].name)
		if length > max {
			max = length
		}
	}
	max++
}

type testcase struct {
	name string
	fn   func() error
	err  error
}

func (t *testcase) String() string {
	s := t.name + ":"
	padsize := max + 1 - len(s)
	for i := 0; i < padsize; i++ {
		s += " "
	}

	if t.err != nil {
		s += tester.NG + fmt.Sprintf("(%v)", t.err)
	} else {
		s += tester.OK
	}
	return s
}
