package cartridge

func HaveSRAM(h *Header) bool {
	c := h.Chipset.lists
	for i := range c {
		if c[i] == "Battery" {
			return true
		}
	}
	return false
}
