package types

type Renderer struct {
	Size           *Rect
	Close          func()
	Update         func() error
	PrintRuneColor func(r rune, posX, posY int32, fg *Colour, bg *Colour) error
	PrintBlink     func(state bool, posX, posY int32) error
}

type Colour struct {
	Red   byte
	Green byte
	Blue  byte
}
