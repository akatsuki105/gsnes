package core

import (
	"fmt"
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/pokemium/gsnes/core/scheduler"
	"github.com/pokemium/iro"
)

type SuperFamicom interface {
	// Reset core
	Reset() error

	// RunFrame runs emulator until a next frame
	RunFrame()

	// Run emulator for a single step
	Run()

	// Run emulator until next event or for max cycles
	RunLoop(max int64)

	// SetKeyInput sends key state into core
	SetKeyInput(key string, press bool)

	// LoadROM loads game rom
	//
	// It assumes an environment with enough memory, so it is necessary to pass the complete ROM data in advance.
	//
	// NOTE: romData is mutable(not copied).
	LoadROM(romData []byte) error

	// PC returns current Program counter.
	//
	// NOTE: Prefetch is ignored.
	PC() uint32

	// Display resolution
	Resolution() (w int, h int)

	// FrameBuffer return framebuffer represents game screen
	FrameBuffer() []iro.RGB555

	Pause(p bool)
	Paused() bool

	debug
}

type debug interface {
	// Get raw memory array(editable)
	//  Region: "ROM", "WRAM", "VRAM", "PALETTE"
	MemoryBuffer(region string) (buffer unsafe.Pointer, length int)

	// Region: "SYSTEM", "CPU", "PPU", "SCREEN", "EVENTS"
	Status(region string) string

	Stack(depth int) (top uint16, stack []uint8)
}

type sfc struct {
	w     *w65816
	ppu   *ppu
	apu   *apu
	s     *scheduler.Scheduler
	frame int // frame counter
	pause bool
	dma   *dmaController
	m     *memory
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
	return nil
}

func (s *sfc) LoadROM(romData []uint8) error {
	c := s.w.cart
	c.loadROM(romData)
	s.Reset()
	return nil
}

// RunFrame runs emulator until a next frame
func (s *sfc) RunFrame() {
	const FRAME = SCANLINE * 4 * TOTAL_SCANLINE
	start := s.s.Cycle()

	old := s.frame
	for old == s.frame && s.s.Cycle()-start < FRAME {
		s.RunLoop(math.MaxInt64)
		if s.pause {
			break
		}
	}
	s.apu.catchup()
}

// Run emulator for a single instruction
func (s *sfc) Run() {
	old := s.pause
	s.pause = false
	for {
		for s.s.AnyEvent() {
			s.processEvents()
		}
		s.w.step()
		if s.w.state == CPU_FETCH {
			break
		}
	}
	s.pause = old
}

// Run all scheduled events
func (s *sfc) processEvents() {
	nextEvent := s.s.NextEvent
	for s.s.RelativeCycles >= nextEvent {
		s.s.NextEvent = math.MaxInt64
		nextEvent = 0

		first := true
		for first || s.w.blocked {
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
			break
		}
	}

	if s.w.blocked {
		s.s.RelativeCycles = s.s.NextEvent
	}
}

// Run emulator until next event or for max cycles
func (s *sfc) RunLoop(max int64) {
	start := s.s.Cycle()
	running := true

	for running || s.w.state != CPU_FETCH {
		if s.w.state == CPU_FETCH {
			if s.w.checkIrq(NMI) || s.w.checkIrq(IRQ) {
				break
			}
		}
		if s.s.RelativeCycles < s.s.NextEvent {
			running = s.w.step() && running
		} else {
			s.processEvents()
			running = false
		}

		now := s.s.Cycle()
		if now-start >= max {
			break
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

func (s *sfc) MemoryBuffer(region string) (buffer unsafe.Pointer, length int) {
	switch region {
	case "ROM":
		return unsafe.Pointer(&s.w.cart.rom[0]), len(s.w.cart.rom)
	case "WRAM":
		return unsafe.Pointer(&s.w.wram.buf[0]), len(s.w.wram.buf)
	case "VRAM":
		return unsafe.Pointer(&s.ppu.vram.buf[0]), len(s.ppu.vram.buf) * 2
	case "PALETTE":
		return unsafe.Pointer(&s.ppu.pal.buf[0]), len(s.ppu.pal.buf) * 2
	default:
		fmt.Fprintf(os.Stderr, "Unknown region: %s", region)
		return nil, 0
	}
}

func (s *sfc) Status(region string) string {
	region = strings.ToUpper(region)
	switch region {
	case "SYSTEM":
		return fmt.Sprintf("Cycle: %d", s.s.Cycle())
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
