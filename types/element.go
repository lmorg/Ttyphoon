package types

type Element interface {
	Generate(*ApcSlice, *Sgr) error
	Write(rune) error
	Rune(*XY) rune
	Size() *XY
	Draw(*XY, *XY)
	MouseClick(*XY, MouseButtonT, uint8, ButtonStateT, EventIgnoredCallback)
	MouseWheel(*XY, *XY, EventIgnoredCallback)
	MouseMotion(*XY, *XY, EventIgnoredCallback)
	MouseOut()
}

type ElementID int

const (
	ELEMENT_ID_IMAGE ElementID = iota
	ELEMENT_ID_SIXEL
	ELEMENT_ID_CSV
	ELEMENT_ID_HYPERLINK
)
