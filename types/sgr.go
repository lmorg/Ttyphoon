package types

type Sgr struct {
	Bitwise SgrFlag
	Fg      *Colour
	Bg      *Colour
}

func (s *Sgr) Reset() {
	s.Bitwise = 0
	s.Fg = SGR_DEFAULT.Fg
	s.Bg = SGR_DEFAULT.Bg
}

func (s *Sgr) Copy() *Sgr {
	if s == nil {
		return SGR_DEFAULT.Copy()
	}
	return &Sgr{
		Fg:      s.Fg,
		Bg:      s.Bg,
		Bitwise: s.Bitwise,
	}
}

func (s *Sgr) HashValue() uint64 {
	return (uint64(s.Bitwise) << 48) | (uint64(s.Fg.RGB24()) << 24) | uint64(s.Bg.RGB24())
}

type SgrFlag uint16

// Flags
const (
	SGR_NORMAL SgrFlag = 0
	SGR_BOLD   SgrFlag = 1 << iota
	SGR_ITALIC
	SGR_UNDERLINE
	SGR_STRIKETHROUGH
	SGR_SLOW_BLINK
	SGR_INVERT

	SGR_HIGHLIGHT_SEARCH_RESULT

	// _SGR_PLACEHOLDER1
	// _SGR_PLACEHOLDER2
	// _SGR_PLACEHOLDER3
	// _SGR_PLACEHOLDER4
	// _SGR_PLACEHOLDER5
	// _SGR_PLACEHOLDER6
	// _SGR_PLACEHOLDER7
	// _SGR_PLACEHOLDER8
)

func (f SgrFlag) Is(flag SgrFlag) bool {
	return f&flag != 0
}

func (f *SgrFlag) Set(flag SgrFlag) {
	*f |= flag
}

func (f *SgrFlag) Unset(flag SgrFlag) {
	*f &^= flag
}
