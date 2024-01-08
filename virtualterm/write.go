package virtualterm

import (
	"log"

	"github.com/lmorg/mxtty/codes"
)

func (term *Term) writeCell(r rune) {
	term.cell().char = r
	term.cell().sgr = term.sgr.Copy()
	term.wrapCursorForwards()
}

// Write multiple characters to the virtual terminal
func (term *Term) printLoop() {
	var r rune

	//term.mutex.Lock()

	for {
		r = term.Pty.ReadRune()
		term.slowBlinkState = true

		switch r {
		case codes.AsciiEscape:
			r = term.Pty.ReadRune()
			switch r {
			case '[': // CSI code
				term.parseCsiCodes()
			case ']': // OSC code
				term.parseOscCodes()
			default:
				term.writeCell(r)
			}

		case codes.AsciiBackspace, codes.IsoBackspace:
			_ = term.moveCursorBackwards(1)

		case codes.AsciiCtrlG: // bell
			// TODO: beep

		case '\t':
			indent := int(4 - (term.curPos.X % term.tabWidth))
			for i := 0; i < indent; i++ {
				term.writeCell(' ')
			}

		case '\r':
			term.curPos.X = 0

		case '\n':
			if term.moveCursorDownwards(1) > 0 {
				term.moveContentsUp()
				term.moveCursorDownwards(1)
			}
			term.curPos.X = 0

		//case ' ':
		//	term.writeCell('·')

		default:
			if r < 32 {
				log.Printf("Unexpected ASCII control character: %d", r)
				//} else {
				//log.Printf("Character code %d (%s)", text[i], string(text[i]))
			}
			term.writeCell(r)
		}
	}

	//term.mutex.Unlock()
}

func multiplyN(n *int32, r rune) {
	if *n < 0 {
		*n = 0
	}

	*n = (*n * 10) + (r - 48)
}
