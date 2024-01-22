package virtualterm

import (
	"log"
	"strings"

	"github.com/lmorg/mxtty/codes"
)

/*
	Reference documentation used:
	- https://en.wikipedia.org/wiki/ANSI_escape_code#OSC_(Operating_System_Command)_sequences
	- https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
	- ChatGPT (when the documentation above was unclear)
*/

func (term *Term) parseOscCodes() {
	var (
		r    rune
		text []rune
	)

	for {
		r = term.Pty.Read()
		text = append(text, r)
		switch r {

		case codes.AsciiEscape:
			r = term.Pty.Read()
			if r == '\\' { // ST (OSC terminator)
				goto parsed
			}
			text = append(text, r)
			continue

		case codes.AsciiCtrlG: // bell (xterm OSC terminator)
			goto parsed

		}

	}
parsed:
	text = text[:len(text)-1]

	stack := strings.Split(string(text), ";")

	switch stack[0] {
	case "0":
		// change icon and window title
		term.renderer.SetWindowTitle(stack[1])

	case "2":
		// change window title
		term.renderer.SetWindowTitle(stack[1])

	case "1337":
		//$(osc)1337;File=inline=1:${base64 -i $file -o -}

	default:
		log.Printf("WARNING: Unknown OSC code %s: %s", stack[0], string(text[:len(text)-1]))
	}
}
