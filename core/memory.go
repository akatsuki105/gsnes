package core

import (
	"math"
	"strconv"
	"strings"
)

/*
              0x00..3F              0x40..7D         0x7E..7F        0x80..BF           0xC0..FF
0x0000 ┌─────────────────────┬────────────────────┬────────────┬───────────────────┬─────────────────┐
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │       System        │                    │            │      System       │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │     WS1 HiROM      │    WRAM    │                   │    WS2 HiROM    │
0x8000 ├─────────────────────┤                    │            ├───────────────────┤                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │      WS1 LoROM      │                    │            │     WS2 LoROM     │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       │                     │                    │            │                   │                 │
       └─────────────────────┴────────────────────┴────────────┴───────────────────┴─────────────────┘
*/

type memory struct {
	// メモリブロックに読み書きを行う関数(idがインデックス)
	reader [256]func(addr uint, defaultVal uint8) uint8
	writer [256]func(addr uint, val uint8)

	// アドレス -> id
	lookup []uint

	// アドレス -> ミラーとかを考慮したアドレス
	target []uint

	// メモリブロックが持っているアドレス数(idがインデックス)
	counter [256]uint

	before uint
}

func newMemory() *memory {
	m := &memory{
		lookup: make([]uint, 24*MB),
		target: make([]uint, 24*MB),
	}

	m.reader[0] = func(addr uint, defaultVal uint8) uint8 {
		// bank, ofs := m.before>>16, m.before&0xFFFF
		// crash("not mapped address: %02X:%04X", bank, ofs)
		return defaultVal
	}
	m.writer[0] = func(addr uint, val uint8) {}

	return m
}

type _memblock struct {
	// メモリブロックの範囲を表す
	//   "00-3f,80-bf:2180-2183,4016-4017,4200-421f"
	addr string

	read  func(addr uint, defaultVal uint8) uint8
	write func(addr uint, val uint8)
	_mask uint
}

func memblock(addr string, read func(addr uint, defaultVal uint8) uint8, write func(addr uint, val uint8)) *_memblock {
	return &_memblock{
		addr:  addr,
		read:  read,
		write: write,
		_mask: math.MaxUint,
	}
}

func (m *_memblock) mask(mask uint) *_memblock {
	m._mask = mask
	return m
}

func (m *memory) mmap(mb *_memblock) {
	id := uint(1)
	for m.counter[id] > 0 {
		id++
		if id >= 256 {
			crash("too many blocks")
		}
	}

	m.reader[id], m.writer[id] = mb.read, mb.write

	// "00-3f,80-bf:2180-2183,4016-4017,4200-421f" => ["00-3f,80-bf", "2180-2183,4016-4017,4200-421f"]
	p := strings.Split(mb.addr, ":")
	if len(p) != 2 {
		crash("invalid address format: %s", mb.addr)
	}

	// "00-3f,80-bf" => ["00-3f", "80-bf"]
	banks := strings.Split(p[0], ",")
	// "2180-2183,4016-4017,4200-421f" => ["2180-2183", "4016-4017", "4200-421f"]
	addrs := strings.Split(p[1], ",")

	for _, b := range banks {
		for _, a := range addrs {
			// "00-3f" -> ["00", "3f"]
			bankRange := strings.Split(b, "-")
			bankLo, _ := strconv.ParseUint(bankRange[0], 16, 8)
			bankHi, _ := strconv.ParseUint(bankRange[1], 16, 8)

			// "2180-2183" -> ["2180", "2183"]
			addrRange := strings.Split(a, "-")
			addrLo, _ := strconv.ParseUint(addrRange[0], 16, 16)
			addrHi, _ := strconv.ParseUint(addrRange[1], 16, 16)

			// fmt.Printf("%02X-%02X:%04X-%04X\n", uint8(bankLo), uint8(bankHi), uint16(addrLo), uint16(addrHi))

			for bank := uint(bankLo); bank <= uint(bankHi); bank++ {
				for addr := uint(addrLo); addr <= uint(addrHi); addr++ {
					pid := m.lookup[bank<<16|addr]
					if pid > 0 {
						m.counter[pid]--
						if m.counter[pid] == 0 {
							m.reader[pid] = m.reader[0]
							m.writer[pid] = m.writer[0]
						}
					}

					ofs := (bank<<16 | addr) & mb._mask
					m.lookup[bank<<16|addr] = id
					m.target[bank<<16|addr] = ofs
					m.counter[id]++
				}
			}
		}
	}
}
