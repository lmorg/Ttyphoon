package virtualterm

import (
	"fmt"
	"log"

	"github.com/lmorg/mxtty/codes"
)

func (term *Term) csiCursorPosSave() {
	term._savedCurPos = term.curPos
}

func (term *Term) csiCursorPosRestore() {
	term.curPos = term._savedCurPos
}

func (term *Term) csiScreenBufferAlternative() {
	term.cells = &term._altBuf
}

func (term *Term) csiScreenBufferNormal() {
	term.cells = &term._normBuf
	for i := range term._altBuf {
		term._altBuf[i] = make([]cell, term.size.X)
	}
}

func (term *Term) csiCursorHide() {
	term._hideCursor = true
}

func (term *Term) csiCursorShow() {
	term._hideCursor = false
}

func (term *Term) csiSetScrollingRegion(region []int32) {
	term._scrollRegion = &scrollRegionT{
		Top:    region[0],
		Bottom: region[1],
	}
}

func (term *Term) csiWindowTitleStackSaveTo() {
	term._windowTitleStack = append(term._windowTitleStack, term.renderer.GetWindowTitle())
}

func (term *Term) csiWindowTitleStackRestoreFrom() {
	title := term._windowTitleStack[len(term._windowTitleStack)-1]
	term.renderer.SetWindowTitle(title)
	term._windowTitleStack = term._windowTitleStack[:len(term._windowTitleStack)-1]
}

func (term *Term) csiCallback(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	_, err := term.Pty.Secondary.WriteString(codes.Csi + msg)
	if err != nil {
		log.Printf("ERROR: writing callback message '%s': %s", msg, err.Error())
	}
}

func (term *Term) csiRepeatPreceding(n int32) {
	if n < 1 {
		n = 1
	}
	cell, _ := term.previousCell()
	for i := int32(0); i < n; i++ {
		term.writeCell(cell.char)
	}
}
