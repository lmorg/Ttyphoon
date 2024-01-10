package virtualterm

import (
	"log"
	"sync"

	"github.com/lmorg/mxtty/types"
)

// Term is the display state of the virtual term
type Term struct {
	size     *types.XY
	curPos   types.XY
	sgr      *sgr
	renderer types.Renderer
	Pty      types.Pty
	_mutex   sync.Mutex

	_slowBlinkState bool

	cells    *[][]cell
	_normBuf [][]cell
	_altBuf  [][]cell

	// CSI states
	_tabWidth         int32
	_hideCursor       bool
	_savedCurPos      types.XY
	_scrollRegion     *scrollRegionT
	_windowTitleStack []string

	// line feed redraw
	_lfEnabled   bool
	_lfNum       int32
	_lfFrequency int32

	// state
	//_altBufActive bool
	_activeElement mxapc
}

func (term *Term) lfRedraw() {
	if !term._lfEnabled {
		return
	}

	term._lfNum++
	if term._lfNum >= term._lfFrequency {
		term._lfNum = 0
		term.renderer.TriggerRedraw()
	}
}

type cell struct {
	char rune
	// 0: empty
	// 1: render element
	// 2: element child

	sgr *sgr

	element types.Element
}

const (
	CELL_NULL          = 0
	CELL_ELEMENT_START = 1
	CELL_ELEMENT_FILL  = 2
)

/*
Types of elements:
- image rendering
- json tree
- table sorting
*/

// NewTerminal creates a new virtual term
func NewTerminal(renderer types.Renderer) *Term {
	size := renderer.Size()

	normBuf := make([][]cell, size.Y)
	for i := range normBuf {
		normBuf[i] = make([]cell, size.X)
	}
	altBuf := make([][]cell, size.Y)
	for i := range altBuf {
		altBuf[i] = make([]cell, size.X)
	}

	term := &Term{
		renderer:  renderer,
		_normBuf:  normBuf,
		_altBuf:   altBuf,
		size:      size,
		sgr:       SGR_DEFAULT.Copy(),
		_tabWidth: 8,
	}

	term.cells = &term._normBuf

	term._lfFrequency = 2
	term._lfEnabled = true

	return term
}

func (term *Term) GetSize() *types.XY {
	return term.size
}

func (term *Term) cell() *cell {
	if term.curPos.X < 0 {
		log.Printf("ERROR: term.curPos.X < 0(returning first cell) TODO fixme")
		term.curPos.X = 0
	}

	if term.curPos.Y < 0 {
		log.Printf("ERROR: term.curPos.Y < 0 (returning first cell) TODO fixme")
		term.curPos.Y = 0
	}

	if term.curPos.X >= term.size.X {
		log.Printf("ERROR: term.curPos.X >= term.size.X (returning last cell) TODO fixme")
		term.curPos.X = term.size.X - 1
	}

	if term.curPos.Y >= term.size.Y {
		log.Printf("ERROR: term.curPos.Y >= term.size.Y (returning last cell) TODO fixme")
		term.curPos.Y = term.size.Y - 1
	}

	return &(*term.cells)[term.curPos.Y][term.curPos.X]
}

func (term *Term) previousCell() (*cell, *types.XY) {
	pos := term.curPos
	pos.X--

	if pos.X < 0 {
		pos.X = 0
		pos.Y--
	}

	if pos.Y < 0 {
		pos.Y = 0
	}

	return &(*term.cells)[pos.Y][pos.X], &pos
}

type scrollRegionT struct {
	Top    int32
	Bottom int32
}

func (term *Term) getScrollRegion() (top int32, bottom int32) {
	if term._scrollRegion == nil {
		top = 0
		bottom = term.size.Y - 1
	} else {
		top = term._scrollRegion.Top - 1
		bottom = term._scrollRegion.Bottom - 1
	}

	return
}

func (term *Term) Reply(b []byte) error {
	return term.Pty.Write(b)
}

func (term *Term) Bg() *types.Colour {
	return SGR_DEFAULT.bg
}
