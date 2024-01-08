package virtualterm

import "log"

func isCsiTerminator(r rune) bool {
	return r >= 0x40 && r <= 0x7E
}

func parseCsiCodes(term *Term, text []rune) int {
	i := 2

	var (
		n     int32 = -1 // default value
		stack []int32
	)

	for ; i < len(text); i++ {
		if text[i] >= '0' && '9' >= text[i] {
			multiplyN(&n, text[i])
			continue
		}

		switch text[i] {
		case 'A', 'E': // moveCursorUp
			term.moveCursorUpwards(n)

		case 'B', 'F': // moveCursorDown
			term.moveCursorDownwards(n)

		case 'C': // moveCursorForwards
			term.moveCursorForwards(n)

		case 'D': // moveCursorBackwards
			term.moveCursorBackwards(n)

		case 'm': // SGR
			lookupSgr(term.sgr, n)

		case 'H': // moveCursor
			if len(stack) != 2 {
				term.curPos = xy{}
			} else {
				term.curPos = xy{
					X: stack[0] + 1,
					Y: stack[1] + 1,
				}
			}

		case 'J': // eraseDisplay...
			switch n {
			case -1, 0:
				term.eraseDisplayAfter()
			case 1:
				term.eraseDisplayBefore()
			case 2, 3:
				term.eraseDisplay() // TODO: 3 should erase scrollback buffer
			}

		case 'K': // clearLine...
			switch n {
			case -1, 0:
				term.eraseLineAfter()
			case 1:
				term.eraseLineBefore()
			case 2:
				term.eraseLine()
			}

		case '?': // private codes
			adjust, n, r := parseNumericAlphaCodes(i, text)
			log.Printf("CSI private code gobbled: '[?%d%s'", n, string(r))
			return i + adjust - 3

		case ';':
			stack = append(stack, n)
			//log.Printf("Unhandled CSI parameter: '%d;'", n)

		default:
			log.Printf("Unknown CSI code: '%d%s'", n, string(text[i]))
		}

		if isCsiTerminator(text[i]) {
			return i - 1
		}
	}
	return i
}

func parseNumericAlphaCodes(i int, text []rune) (int, int32, rune) {
	i++
	var n int32 = -1 // default value

	for ; i < len(text); i++ {
		if text[i] >= '0' && '9' >= text[i] {
			//n = (n * 10) + (text[i] - 48)
			multiplyN(&n, text[i])
			continue
		}

		if isCsiTerminator(text[i]) {
			return i, n, text[i]
		}

		log.Printf("Unexpected character in private CSI sequence: %s", string(text[i]))
		return i, n, text[i]
	}
	return i, n, 0
}

func lookupSgr(sgr *sgr, n int32) {
	switch n {
	case 0: // reset / normal
		sgr.sgrReset()

	case 1: // bold
		sgr.Set(sgrBold)

	case 4: // underscore
		sgr.Set(sgrUnderscore)

	case 5: // blink
		sgr.Set(sgrBlink)

	case 7: // invert
		sgr.Set(sgrInvert)

		//
		// 4bit foreground colour:
		//

	case 30: // fg black
		sgr.fg = sgrColour4Black

	case 31: // fg red
		sgr.fg = sgrColour4Red

	case 32: // fg green
		sgr.fg = sgrColour4Green

	case 33: // fg yellow
		sgr.fg = sgrColour4Yellow

	case 34: // fg blue
		sgr.fg = sgrColour4Blue

	case 35: // fg magenta
		sgr.fg = sgrColour4Magenta

	case 36: // fg cyan
		sgr.fg = sgrColour4Cyan

	case 37: // fg white
		sgr.fg = sgrColour4White

	default:
		log.Printf("Unknown SGR code: %d", n)
	}
}
