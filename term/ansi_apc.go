package virtualterm

import (
	"fmt"

	"github.com/lmorg/mxtty/codes"
	"github.com/lmorg/mxtty/types"
)

/*
	Reference documentation used:
	- xterm: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Application-Program-Command-functions
*/

func (term *Term) parseApcCodes() {
	var (
		r    rune
		err  error
		text []rune
	)

	for {
		r, err = term.Pty.Read()
		if err != nil {
			return
		}
		text = append(text, r)
		if r == codes.AsciiEscape {
			r, err = term.Pty.Read()
			if err != nil {
				return
			}
			if r == '\\' { // ST (APC terminator)
				text = text[:len(text)-1]
				break
			}
			text = append(text, r)
			continue
		}
		if r == codes.AsciiCtrlG { // bell (xterm OSC terminator)
			text = text[:len(text)-1]
			break
		}
	}

	apc := types.NewApcSlice(text)

	switch apc.Index(0) {
	case "begin":
		switch apc.Index(1) {
		case "csv":
			term.mxapcBegin(types.ELEMENT_ID_CSV, apc)

		case "output-block":
			term.mxapcBeginOutputBlock(apc)

		default:
			term.renderer.DisplayNotification(types.NOTIFY_DEBUG,
				fmt.Sprintf("Unknown mxAPC code %s: %s", apc.Index(1), string(text[:len(text)-1])))
		}

	case "end":
		switch apc.Index(1) {
		case "csv":
			term.mxapcEnd(types.ELEMENT_ID_CSV, apc)

		case "output-block":
			term.mxapcEndOutputBlock(apc)

		default:
			term.renderer.DisplayNotification(types.NOTIFY_DEBUG,
				fmt.Sprintf("Unknown mxAPC code %s: %s", apc.Index(1), string(text[:len(text)-1])))
		}

	case "insert":
		switch apc.Index(1) {
		case "image":
			term.mxapcInsert(types.ELEMENT_ID_IMAGE, apc)
		}

	default:
		term.renderer.DisplayNotification(types.NOTIFY_DEBUG,
			fmt.Sprintf("Unknown mxAPC code %s: %s", apc.Index(1), string(text[:len(text)-1])))
	}
}
