package core

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/akatsuki105/gsnes/core/scheduler"
	"github.com/akatsuki105/iro"
)

type SuperFamicom interface {
	// Reset core
	Reset() error

	// RunFrame runs emulator until a next frame
	RunFrame()

	// Run emulator for a single instruction
	RunInst()

	// Sends key state into core
	SetKeyInput(key string, press bool)

	// Load game rom.
	LoadROM(romData []byte) error

	// PC returns current Program counter(Prefetch is ignored)
	PC() uint32

	// Display resolution
	Resolution() (w int, h int)

	// Return framebuffer represents game screen
	FrameBuffer() []iro.RGB555

	Pause(p bool)
	Paused() bool

	// Debug feature

	// Replace builtin memory buffer by your buffer.
	MMap(region string, buf []uint8) error

	Debug
}

// 完成時には消す
type Debug interface {
	// Region: "SYSTEM", "CPU", "PPU", "SCREEN", "EVENTS"
	Status(region string) string

	Stack(depth int) (top uint16, stack []uint8)
}

type sfc struct {
	w         *w65816
	ppu       *ppu
	apu       *apu
	s         *scheduler.Scheduler
	frame     int // frame counter
	pause     bool
	dma       *dmaController
	m         *memory
	earlyExit bool
}

func New() SuperFamicom {
	sc := scheduler.New() // in SNES, 1 cycle is 1 master cycle.
	s := &sfc{
		s:   sc,
		apu: newApu(),
		m:   newMemory(),
	}
	s.ppu = newPpu(s)
	s.w = new65816(s, &sc.RelativeCycles, &sc.NextEvent)
	s.dma = newDmaController(s)

	s.m.mmap(memblock("00-3F,80-BF:0000-1FFF", s.w.wram.read, s.w.wram.write).mask(0x1FFF))                 // WRAM(mirror)
	s.m.mmap(memblock("00-3F,80-BF:2100-213F", s.ppu.readIO, s.ppu.writeIO).mask(0x3F))                     // PPU
	s.m.mmap(memblock("00-3F,80-BF:2140-217F", s.apu.readIO, s.apu.writeIO).mask(0x3))                      // APU
	s.m.mmap(memblock("00-3F,80-BF:2180-2183,4016-4017,4200-421F", s.w.readCPU, s.w.writeCPU).mask(0xFFFF)) // CPU
	s.m.mmap(memblock("00-3F,80-BF:4300-437F", s.dma.readIO, s.dma.writeIO).mask(0x7F))                     // DMA
	s.m.mmap(memblock("7E-7F:0000-FFFF", s.w.wram.read, s.w.wram.write).mask(0x1FFFF))                      // WRAM
	return s
}

func (s *sfc) Reset() error {
	s.s.Reset()
	s.w.reset()
	s.ppu.reset()
	s.apu.reset()
	s.dma.reset()
	s.frame = 0
	s.earlyExit = false
	return nil
}

func (s *sfc) LoadROM(romData []uint8) error {
	rom := make([]uint8, len(romData))
	copy(rom, romData)
	c := s.w.cart
	c.loadROM(rom)
	s.Reset()
	return nil
}

// RunFrame runs emulator until a next frame
func (s *sfc) RunFrame() {
	defer s.panicHandler(true)

	const FRAME = SCANLINE * 4 * TOTAL_SCANLINE
	start := s.s.Cycle()

	old := s.frame
	for old == s.frame && s.s.Cycle()-start < FRAME {
		s.run()
		if s.pause {
			break
		}
	}
	s.apu.catchup()
}

// Run emulator for a single instruction
func (s *sfc) RunInst() {
	old := s.pause
	s.pause = false
	defer func() { s.pause = old }()

	if s.w.checkIrq(NMI) || s.w.checkIrq(IRQ) {
		return
	}

	for {
		for s.s.AnyEvent() {
			s.processEvents()
		}
		s.w.step()
		if s.w.state == CPU_FETCH {
			break
		}
	}
}

// Run all scheduled events
func (s *sfc) processEvents() {
	nextEvent := s.s.NextEvent
	for s.s.RelativeCycles >= nextEvent {
		s.s.NextEvent = math.MaxInt64
		nextEvent = 0

		first := true
		for first || (s.w.blocked() && !s.earlyExit) {
			first = false

			cycles := s.s.RelativeCycles
			s.s.RelativeCycles = 0

			if cycles < nextEvent {
				nextEvent = s.s.Add(nextEvent)
			} else {
				nextEvent = s.s.Add(cycles)
			}
		}

		s.s.NextEvent = nextEvent
		if s.w.halted {
			*s.w.cycles = nextEvent
			if s.w.r.p.i || (s.w.nmitimen>>4)&0b11 == 0 {
				break
			}
		}

		if s.earlyExit {
			break
		}
	}

	s.earlyExit = false
	if s.w.blocked() {
		s.s.RelativeCycles = s.s.NextEvent
	}
}

// Run emulator until next event
func (s *sfc) run() {
	running := true
	for running || s.w.state != CPU_FETCH {
		if s.w.state == CPU_FETCH {
			if s.w.checkIrq(NMI) || s.w.checkIrq(IRQ) {
				break
			}
			if s.dma.pending {
				s.dma.initGDMA()
				s.processEvents()
			}
		}

		if s.s.RelativeCycles < s.s.NextEvent {
			running = s.w.step() && running
		} else {
			s.processEvents()
			running = false
		}
	}
}

func (s *sfc) PC() uint32 {
	return s.w.r.pc.u32()
}

// Display resolution
func (s *sfc) Resolution() (w int, h int) {
	return HORIZONTAL, VERTICAL // NTSC
}

func (s *sfc) FrameBuffer() []iro.RGB555 {
	if s.ppu.inFBlank {
		return fblankScreen[:]
	}
	return s.ppu.r.frameBuffer()
}

func (s *sfc) Status(region string) string {
	region = strings.ToUpper(region)
	switch region {
	case "SYSTEM":
		result := `Cycle: %d
GDMA: %s, HDMA: %s`
		return fmt.Sprintf(result, s.s.Cycle(), bitField(s.dma.gdmaen), bitField(s.dma.hdmaen))

	case "CPU":
		return s.w.Status()

	case "PPU", "SCREEN", "OAM":
		return s.ppu.Status(region)

	case "EVENTS":
		return s.s.String()

	default:
		return ""
	}
}

func (s *sfc) Pause(p bool) {
	s.pause = p
}

func (s *sfc) Paused() bool {
	return s.pause
}

func (s *sfc) Stack(depth int) (uint16, []uint8) {
	sp := s.w.r.s
	if depth <= 0 {
		return sp, []uint8{}
	}
	top := sp + 1
	stack := make([]uint8, depth)
	for i := 0; i < depth; i++ {
		stack[i] = s.w.wram.buf[top]
		top++
		if top >= 0x200 {
			break
		}
	}

	return sp, stack
}

func (s *sfc) SetKeyInput(key string, press bool) {
	j := &s.w.joypads[0].buf
	switch key {
	case "A":
		j[7-4] = press
	case "B":
		j[15-4] = press
	case "X":
		j[6-4] = press
	case "Y":
		j[14-4] = press
	case "L":
		j[5-4] = press
	case "R":
		j[4-4] = press
	case "UP":
		j[11-4] = press
	case "DOWN":
		j[10-4] = press
	case "LEFT":
		j[9-4] = press
	case "RIGHT":
		j[8-4] = press
	case "START":
		j[12-4] = press
	case "SELECT":
		j[13-4] = press
	}
}

func (s *sfc) MMap(region string, buf []uint8) error {
	switch region {
	case "WRAM":
		if len(buf) != int(128*KB) {
			return errors.New("WRAM buffer must be 128KB")
		}

		copy(buf, s.w.wram.buf[:])
		s.w.wram.buf = buf

	case "VRAM":
		if len(buf) != int(VRAM_SIZE) {
			return errors.New("VRAM buffer must be 64KB")
		}

		buf16 := (*[VRAM_SIZE / 2]uint16)(unsafe.Pointer(&buf[0]))
		copy(buf16[:], s.ppu.vram.buf[:])
		s.ppu.vram.buf = buf16[:]

	case "PALETTE":
		if len(buf) != int(PAL_SIZE) {
			return errors.New("palette buffer must be 512B")
		}

		buf16 := (*[PAL_SIZE / 2]iro.RGB555)(unsafe.Pointer(&buf[0]))
		copy(buf16[:], s.ppu.pal.buf[:])
		s.ppu.pal.buf = buf16[:]
	}

	return errors.New("invalid region")
}

func (s *sfc) panicHandler(stack bool) {
	if err := recover(); err != nil {
		fmt.Fprintf(os.Stderr, "Panic in %v\n", s.w.lastInstAddr)
		fmt.Fprintf(os.Stderr, "         %s\n", err)

		size := len(histories)
		fmt.Fprintln(os.Stderr, "         History:")
		for i := range histories {
			fmt.Fprintf(os.Stderr, "           %s\n", histories[size-1-i])
		}

		for depth := 3; ; depth++ {
			_, file, line, ok := runtime.Caller(depth)
			if !ok {
				break
			}
			fmt.Fprintf(os.Stderr, "======> %d: %v:%d\n", depth, file, line)
		}
		os.Exit(1)
	}
}
