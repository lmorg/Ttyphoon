package virtualterm

import (
	"log"
	"regexp"

	"github.com/lmorg/mxtty/codes"
)

func (term *Term) writeCell(r rune) {
	term.cell().char = r
	term.cell().sgr = term.sgr.Copy()
	term.wrapCursorForwards()
}

var rxLazyCsiCheck = regexp.MustCompile(`\[[;\?a-zA-Z0-9]+`)

// Write multiple characters to the virtual terminal
func (term *Term) write() {
	var (
		escape bool
		i      int
		p      = make([]byte, 10*1024)
		text   []rune
	)

	for {
		n, err := term.Pty.Read(p)
		if err != nil {
			log.Printf("error reading from buffer (%d bytes dropped): %s", n, err.Error())
			continue
		}
		r := []rune(string(p[:n]))

		if escape {
			text = append(text, r...)
		} else {
			text = r
			i = 0
		}

		//term.mutex.Lock()

		for ; i < len(text); i++ {
			switch text[i] {
			case codes.AsciiEscape:
				escape = true
				continue

			case codes.AsciiBackspace, codes.IsoBackspace:
				_ = term.moveCursorBackwards(1)

			case codes.AsciiCtrlG: // bell
				// TODO: beep

			case '[':
				if !escape {
					term.writeCell(text[i])
					continue
				}
				escape = false
				start := i
				i += parseCsiCodes(term, text[i-1:])
				if i >= len(text) {
					i = len(text) - 1
					log.Printf("Parsing error: i>len(text)")
				}
				if !rxLazyCsiCheck.MatchString(string(text[start : i+1])) {
					log.Printf("Invalid CSI code parsed: %v (%s)", []byte(string(text[start:i+1])), string(text[start:i+1]))
					//} else {
					//	log.Printf("Valid CSI code parsed: %v (%s)", []byte(string(text[start:i+1])), string(text[start:i+1]))
				}

			case ']': // TODO: OSC
				if !escape {
					term.writeCell(text[i])
					continue
				}
				start := i
				for ; i < len(text); i++ {
					if text[i] == 'S' && i < len(text) && text[i+1] == 'T' { // ST  (OSC terminator)
						i += 2
						break
					}
					if text[i] == codes.AsciiCtrlG { // bell (xterm OSC terminator)
						i++
						break
					}
				}
				log.Printf("TODO: OSC sequences: '%s'", string(text[start:i]))

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
				if text[i] < 32 {
					log.Printf("Unexpected ASCII control character: %d", text[i])
				} else {
					//log.Printf("Character code %d (%s)", text[i], string(text[i]))
				}
				term.writeCell(text[i])
			}

			escape = false
		}

		//term.mutex.Unlock()
		term.ExportMxTTY()
	}
}

func multiplyN(n *int32, r rune) {
	if *n < 0 {
		*n = 0
	}

	*n = (*n * 10) + (r - 48)
}
