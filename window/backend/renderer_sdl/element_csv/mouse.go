package elementCsv

import (
	"fmt"
	"strings"

	"github.com/lmorg/mxtty/codes"
	"github.com/lmorg/mxtty/config"
	"github.com/lmorg/mxtty/types"
	"github.com/lmorg/mxtty/window/backend/cursor"
	"golang.design/x/clipboard"
)

func (el *ElementCsv) MouseClick(pos *types.XY, button uint8, clicks uint8, pressed bool, callback types.EventIgnoredCallback) {
	pos.X -= el.renderOffset

	if pos.Y != 0 {
		if button != 1 {
			callback()
			return
		}
		switch clicks {
		case 1:
			if int(pos.Y) > len(el.table) {
				callback()
				return
			}
			for i := range el.boundaries {
				if pos.X <= el.boundaries[i] {
					var start int32
					if i != 0 {
						start = el.boundaries[i-1]
					}
					cell := string(el.table[pos.Y-1][start:el.boundaries[i]])
					clipboard.Write(clipboard.FmtText, []byte(strings.TrimSpace(cell)))
					el.renderer.DisplayNotification(types.NOTIFY_INFO, "Cell copied to clipboard")
					return
				}
			}
			callback()
			return

		case 2:
			el.renderer.DisplayInputBox(fmt.Sprintf("SELECT * FROM '%s' WHERE ... (empty query to reset view)", el.name), el.filter, func(filter string) {
				el.renderOffset = 0
				el.limitOffset = 0
				el.filter = filter
				err := el.runQuery()
				if err != nil {
					el.renderer.DisplayNotification(types.NOTIFY_ERROR, "Cannot sort table: "+err.Error())
				}
			})

		default:
			callback()
			return
		}

		return
	}

	var column int
	for column = range el.boundaries {
		if int(pos.X) < int(el.boundaries[column]) {
			break
		}
	}

	column++ // columns count from 1 because of rowid

	switch button {
	case 1:
		if el.orderByIndex == column {
			el.orderDesc = !el.orderDesc
		} else {
			el.orderByIndex = column
			el.orderDesc = false
		}

	case 3:
		el.orderByIndex = 0
	}

	err := el.runQuery()
	if err != nil {
		el.renderer.DisplayNotification(types.NOTIFY_ERROR, "Cannot sort table: "+err.Error())
	}
}

func (el *ElementCsv) MouseWheel(_ *types.XY, movement *types.XY, callback types.EventIgnoredCallback) {
	termX := el.renderer.GetTermSize().X
	width := el.boundaries[len(el.boundaries)-1]
	mod := codes.Modifier(el.renderer.GetKeyboardModifier())

	if mod == 0 {
		callback()
		return
	}

	if width > termX && movement.X != 0 {

		el.renderOffset += -movement.X * config.Config.Terminal.Widgets.Table.ScrollMultiplierX

		if el.renderOffset > 0 {
			el.renderOffset = 0
		}

		if el.renderOffset < -(width - termX) {
			el.renderOffset = -(width - termX)
		}
	}

	if el.lines >= el.size.Y && movement.Y != 0 {

		el.limitOffset += -movement.Y * config.Config.Terminal.Widgets.Table.ScrollMultiplierY

		if el.limitOffset < 0 {
			el.limitOffset = 0
		}

		if el.limitOffset > el.lines-el.size.Y {
			el.limitOffset = el.lines - el.size.Y
		}

		err := el.runQuery()
		if err != nil {
			el.renderer.DisplayNotification(types.NOTIFY_ERROR, err.Error())
		}
	}
}

func (el *ElementCsv) MouseMotion(pos *types.XY, move *types.XY, callback types.EventIgnoredCallback) {
	switch {
	case pos.Y == 0:
		cursor.Hand()
		el.renderer.StatusBarText("[Left Click] Sort row (ASC|DESC)  |  [Right Click] Remove sort  |  [Ctrl+Scroll] Scroll table")

	case int(pos.Y) <= len(el.table):
		el.renderer.StatusBarText("[Click] Copy cell text to clipboard  |  [2x Click] Filter table (SQL)  |  [Ctrl+Scroll] Scroll table")

	default:
		el.renderer.StatusBarText("")
	}

	if pos.Y < 1 || int(pos.Y) > len(el.table) || pos.X > el.boundaries[len(el.boundaries)-1] {
		el.highlight = nil
		return
	}

	el.highlight = pos
	el.renderer.TriggerRedraw()
}

func (el *ElementCsv) MouseOut() {
	el.renderer.StatusBarText("")
	el.highlight = nil
	el.renderer.TriggerRedraw()
}
