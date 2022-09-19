package core

type cpustate int

const (
	CPU_FETCH cpustate = iota
	CPU_READ_PC
	CPU_MEMORY_LOAD
	CPU_MEMORY_STORE
	CPU_DUMMY_READ
)

type exception int

const (
	COP exception = iota
	BRK
	ABORT
	NMI
	IRQ
	RESET
)

const (
	MAX_ROM_SIZE   = 8 * MB
	WRAM_BANK_SIZE = 64 * KB
	WRAM_SIZE      = 2 * WRAM_BANK_SIZE
	SRAM_SIZE      = 512 * KB
	VRAM_SIZE      = 64 * KB
	ROM_SIZE       = MAX_ROM_SIZE + 0x200 + 0x8000
)

const (
	FAST   = 6  // 6 * (3.58/3.58)
	MEDIUM = 8  // 6 * (3.58/2.68)
	SLOW   = 12 // 6 * (3.58/1.78)
)

const (
	// PPU
	INIDISP                    = 0x2100
	OBSEL                      = 0x2101
	OAMADDL, OAMADDH           = 0x2102, 0x2103
	OAMDATA                    = 0x2104
	BGMODE                     = 0x2105
	MOSAIC                     = 0x2106
	BG1SC, BG2SC, BG3SC, BG4SC = 0x2107, 0x2108, 0x2109, 0x210A
	BG12NBA                    = 0x210B
	BG34NBA                    = 0x210C
	BG1HOFS, BG1VOFS           = 0x210D, 0x210E
	BG2HOFS, BG2VOFS           = 0x210F, 0x2110
	BG3HOFS, BG3VOFS           = 0x2111, 0x2112
	BG4HOFS, BG4VOFS           = 0x2113, 0x2114
	VMAIN                      = 0x2115
	VMADDL, VMADDH             = 0x2116, 0x2117
	VMDATAL, VMDATAH           = 0x2118, 0x2119
	CGADD                      = 0x2121
	CGDATA                     = 0x2122
	TM, TS                     = 0x212C, 0x212D
	TMW, TSW                   = 0x212E, 0x212F
	CGWSEL                     = 0x2130
	CGADSUB                    = 0x2131
	COLDATA                    = 0x2132
	SETINI                     = 0x2133
	MPYL, MPYM, MPYH           = 0x2134, 0x2135, 0x2136
	SLHV                       = 0x2137
	RDOAM                      = 0x2138
	RDVRAML, RDVRAMH, RDCGRAM  = 0x2139, 0x213A, 0x213B
	OPHCT, OPVCT               = 0x213C, 0x213D
	STAT77, STAT78             = 0x213E, 0x213F

	// APU
	APUI00 = 0x2140
	APUI01 = 0x2141
	APUI02 = 0x2142
	APUI03 = 0x2143

	// WRAM
	WMDATA = 0x2180
	WMADDL = 0x2181
	WMADDM = 0x2182
	WMADDH = 0x2183

	// CPU(Slow)
	JOYWR = 0x4016 // Write
	JOYA  = 0x4016 // Read
	JOYB  = 0x4017

	// CPU(Fast)
	NMITIMEN       = 0x4200
	WRIO           = 0x4201
	HTIMEL, HTIMEH = 0x4207, 0x4208
	VTIMEL, VTIMEH = 0x4209, 0x420A
	MDMAEN         = 0x420B
	HDMAEN         = 0x420C
	RDNMI          = 0x4210
	TIMEUP         = 0x4211
	HVBJOY         = 0x4212
	JOY1L, JOY1H   = 0x4218, 0x4219
	JOY2L, JOY2H   = 0x421A, 0x421B
	JOY3L, JOY3H   = 0x421C, 0x421D
	JOY4L, JOY4H   = 0x421E, 0x421F

	// DMA
	DMAP0, DMAP1, DMAP2, DMAP3, DMAP4, DMAP5, DMAP6, DMAP7 = 0x4300, 0x4310, 0x4320, 0x4330, 0x4340, 0x4350, 0x4360, 0x4370
	BBAD0, BBAD1, BBAD2, BBAD3, BBAD4, BBAD5, BBAD6, BBAD7 = 0x4301, 0x4311, 0x4321, 0x4331, 0x4341, 0x4351, 0x4361, 0x4371
	A1T0L, A1T1L, A1T2L, A1T3L, A1T4L, A1T5L, A1T6L, A1T7L = 0x4302, 0x4312, 0x4322, 0x4332, 0x4342, 0x4352, 0x4362, 0x4372
	A1T0H, A1T1H, A1T2H, A1T3H, A1T4H, A1T5H, A1T6H, A1T7H = 0x4303, 0x4313, 0x4323, 0x4333, 0x4343, 0x4353, 0x4363, 0x4373
	A1B0, A1B1, A1B2, A1B3, A1B4, A1B5, A1B6, A1B7         = 0x4304, 0x4314, 0x4324, 0x4334, 0x4344, 0x4354, 0x4364, 0x4374
	DAS0L, DAS1L, DAS2L, DAS3L, DAS4L, DAS5L, DAS6L, DAS7L = 0x4305, 0x4315, 0x4325, 0x4335, 0x4345, 0x4355, 0x4365, 0x4375
	DAS0H, DAS1H, DAS2H, DAS3H, DAS4H, DAS5H, DAS6H, DAS7H = 0x4306, 0x4316, 0x4326, 0x4336, 0x4346, 0x4356, 0x4366, 0x4376
	DASB0, DASB1, DASB2, DASB3, DASB4, DASB5, DASB6, DASB7 = 0x4307, 0x4317, 0x4327, 0x4337, 0x4347, 0x4357, 0x4367, 0x4377
)

// dots span
const (
	SCANLINE = 340 // H=0..339
)

// pixel
const (
	HORIZONTAL     = 256
	VERTICAL       = 224
	TOTAL_SCANLINE = 262
)

const (
	DISABLED = iota
	COLOR_2BPP
	COLOR_4BPP
	COLOR_7BPP
	COLOR_8BPP
)

const (
	EVENT_VIDEO  = "EvtVideo"
	EVENT_IRQ    = "EvtIRQ"
	EVENT_HCOUNT = "EvtHDot"
	EVENT_DMA    = "EvtDma"
)

const (
	EVENT_IRQ_PRIO   = 0
	EVENT_VIDEO_PRIO = 1
	EVENT_DMA_PRIO   = 0x10
)

const (
	INIT_CYCLE = 182
)
