package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokemium/gsnes/core/scheduler"
)

const (
	_       = iota
	KB uint = 1 << (10 * iota)
	MB
	GB
)

func addCycle(c *int64, masterCycles int64) {
	if c != nil {
		*c += masterCycles
	}
}

func schedule(name string, s *scheduler.Scheduler, fn func(cyclesLate int64), after int64) {
	e := scheduler.NewEvent(scheduler.EventName(name), fn, EVENT_TMP_PRIO)
	s.Schedule(e, after)
}

// FormatSize convert 1024 into "1KB"
func formatSize(s uint) string {
	switch {
	case s < KB:
		return fmt.Sprintf("%dB", s)
	case s < MB:
		return fmt.Sprintf("%dKB", s/KB)
	case s < GB:
		return fmt.Sprintf("%dMB", s/MB)
	default:
		return fmt.Sprintf("%dB", s)
	}
}

func crash(msg string, a ...any) {
	msg = fmt.Sprintf(msg, a...)
	msg = strings.TrimSuffix(msg, "\n")
	fmt.Fprintln(os.Stderr, msg+"\n")
	panic(msg)
}

// Bit check val's idx bit
func bit[V uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64](val V, idx int) bool {
	if idx < 0 || idx > 63 {
		return false
	}
	return (val & (1 << idx)) != 0
}

func setBit[V uint | uint8 | uint16 | uint32 | int8](val V, idx int, b bool) V {
	old := val
	if b {
		val = old | (1 << idx)
	} else {
		val = old & ^(1 << idx)
	}
	return val
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func btou8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func btou16(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}

func btou32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func addInt16(u uint16, i int16) uint16 {
	if i > 0 {
		u += uint16(i)
	} else {
		u -= uint16(-i)
	}
	return u
}

func btos(b bool) string {
	if b {
		return "OK"
	}
	return "NG"
}

func bitField(val interface{}) string {
	result := ""
	b := map[bool]string{true: "1", false: "0"}
	switch val := val.(type) {
	case uint8:
		for i := 7; i >= 0; i-- {
			result += b[bit(val, i)]
		}
	case uint16:
		for i := 16; i >= 0; i-- {
			result += b[bit(val, i)]
		}
	case uint32:
		for i := 32; i >= 0; i-- {
			result += b[bit(val, i)]
		}
	case uint64:
		for i := 64; i >= 0; i-- {
			result += b[bit(val, i)]
		}
	case uint:
		for i := 64; i >= 0; i-- {
			result += b[bit(val, i)]
		}
	default:
		panic("invalid type")
	}
	return result
}
