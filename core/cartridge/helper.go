package cartridge

import (
	"fmt"
)

const (
	_       = iota
	KB uint = 1 << (10 * iota)
	MB
	GB
)

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
