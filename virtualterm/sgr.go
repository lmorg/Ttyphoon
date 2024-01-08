package virtualterm

type sgr struct {
	bitwise sgrFlag
	fg      *rgb
	bg      *rgb
}

type rgb struct {
	Red, Green, Blue byte
}

func (sgr *sgr) Is(flag sgrFlag) bool {
	return sgr.bitwise&flag != 0
}

func (sgr *sgr) sgrReset() {
	sgr.bitwise = 0
	sgr.fg = SGR_DEFAULT.fg
	sgr.bg = SGR_DEFAULT.bg
}

func (sgr *sgr) Set(flag sgrFlag) {
	sgr.bitwise |= flag
}

func (c *cell) clear() {
	c.char = 0
	c.sgr = &sgr{}
}
