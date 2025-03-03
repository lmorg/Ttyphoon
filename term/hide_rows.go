package virtualterm

import (
	"errors"
	"fmt"

	"github.com/lmorg/mxtty/debug"
	"github.com/lmorg/mxtty/types"
)

func (term *Term) HideRows(start int32, end int32) error {
	if term.IsAltBuf() {
		return errors.New("this feature is not supported in alt buffer")
	}

	term._mutex.Lock()
	defer term._mutex.Unlock()

	newBuf := term._scrollBuf
	newBuf = append(newBuf, term._normBuf...)

	if len(newBuf[start-1].Hidden) != 0 {
		return errors.New("this row already contains hidden rows")
	}

	newBuf[start-1].Hidden = clone(newBuf[start:end])
	debug.Log(newBuf[start-1].Hidden.String())
	length := len(newBuf[start-1].Hidden)
	newBuf = append(newBuf[:start], newBuf[end:]...)

	if len(newBuf) < int(term.size.Y) {
		newBuf = append(term.makeScreen(), newBuf...)
	}

	if term._scrollOffset > 0 {
		term._scrollOffset -= int(end - start)
	}
	term.updateScrollback()

	term._normBuf = clone(newBuf[len(newBuf)-int(term.size.Y):])
	term._scrollBuf = clone(newBuf[:len(newBuf)-int(term.size.Y)])

	term.renderer.DisplayNotification(types.NOTIFY_INFO, fmt.Sprintf("%d rows have been hidden", length))

	return nil
}

func (term *Term) UnhideRows(pos int32) error {
	if term.IsAltBuf() {
		return errors.New("this feature is not supported in alt buffer")
	}

	term._mutex.Lock()
	defer term._mutex.Unlock()

	tmp := term._scrollBuf
	tmp = append(tmp, term._normBuf...)

	length := len(tmp[pos].Hidden)
	debug.Log(tmp[pos].Hidden.String())
	newBuf := append(clone(tmp[:pos+1]), tmp[pos].Hidden...)
	tmp[pos].Hidden = nil
	newBuf = clone(append(newBuf, tmp[pos+1:]...))

	term._normBuf = clone(newBuf[len(newBuf)-int(term.size.Y):])
	term._scrollBuf = clone(newBuf[:len(newBuf)-int(term.size.Y)])

	term.renderer.DisplayNotification(types.NOTIFY_INFO, fmt.Sprintf("%d rows have been unhidden", length))

	return nil
}

func (term *Term) FoldAtIndent(pos *types.XY) error {
	if term.IsAltBuf() {
		return errors.New("folding is not supported in alt buffer")
	}

	row := term.visibleScreen()[pos.Y]
	screen := append(term._scrollBuf, term._normBuf...)
	absPos := term.convertRelPosToAbsPos(pos)

	for i := range row.Cells {
		if row.Cells[i].Rune() != ' ' {
			absPos.X = int32(i)
			absPos.Y++
			_, err := outputBlockFoldIndent(term, screen, absPos, true)
			return err
		}
	}

	return errors.New("cannot fold from an empty line")
}

func outputBlockFoldIndent(term *Term, screen types.Screen, absPos *types.XY, hide bool) (int32, error) {
	var x, y int32
	for y = absPos.Y + 1; int(y) < len(screen); y++ {
		if screen[y].Meta.Is(types.ROW_OUTPUT_BLOCK_END) || screen[y].Meta.Is(types.ROW_OUTPUT_BLOCK_ERROR) {
			goto fold
		}

		for x = int32(0); x <= absPos.X && int(x) < len(*screen[y].Phrase); x++ {

			if (*screen[y].Phrase)[x] == ' ' {
				// next column
				continue
			}

			goto fold
		}
	}

fold:
	if absPos.Y == y-1 {
		return 0, errors.New("nothing to fold")
	}

	if hide {
		term.HideRows(absPos.Y, y)
	}
	return y, nil
}
