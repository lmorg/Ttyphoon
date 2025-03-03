package rendersdl

import (
	"fmt"
	"time"

	"github.com/lmorg/mxtty/codes"
	"github.com/lmorg/mxtty/config"
	"github.com/lmorg/mxtty/integrations"
	"github.com/lmorg/mxtty/types"
	"github.com/veandco/go-sdl2/sdl"
)

type termWidgetT struct{}

func (tw *termWidgetT) eventTextInput(sr *sdlRender, evt *sdl.TextInputEvent) {
	sr.footerText = ""
	b := []byte(evt.GetText())

	if len(b) == 1 {
		if (b[0] >= '0' && b[0] <= '9') || // keypad numeric
			b[0] == '+' || b[0] == '-' || b[0] == '*' || b[0] == '/' || b[0] == '.' || // keypad syms
			b[0] == '`' { // ctrl+\ in tmux

			go func() {
				select {
				case ignore := <-sr.keyIgnore:
					if ignore {
						return
					}
					sr.term.Reply(b)

				case <-time.After(5 * time.Millisecond):
					sr.term.Reply(b)
				}

			}()
			return
		}
	}

	sr.term.Reply(b)
}

func (tw *termWidgetT) eventKeyPress(sr *sdlRender, evt *sdl.KeyboardEvent) {
	go func() {
		// basically this just tells tw.eventTextInput() to ignore input
		if (evt.Keysym.Sym >= sdl.K_KP_DIVIDE && evt.Keysym.Sym <= sdl.K_KP_PERIOD) ||
			(evt.Keysym.Sym == '\\' && (evt.Keysym.Mod == sdl.KMOD_LCTRL || evt.Keysym.Mod == sdl.KMOD_RCTRL)) {
			go func() {
				sr.keyIgnore <- true
			}()
		}

	}()
	tw._eventKeyPress(sr, evt)
}

func (tw *termWidgetT) _eventKeyPress(sr *sdlRender, evt *sdl.KeyboardEvent) {
	sr.footerText = ""
	sr.keyModifier = evt.Keysym.Mod

	switch evt.Keysym.Sym {
	case sdl.K_LSHIFT, sdl.K_RSHIFT, sdl.K_LALT, sdl.K_RALT,
		sdl.K_LCTRL, sdl.K_RCTRL, sdl.K_LGUI, sdl.K_RGUI,
		sdl.K_CAPSLOCK, sdl.K_NUMLOCKCLEAR, sdl.K_SCROLLLOCK, sdl.K_SPACE:
		// modifier keys pressed on their own shouldn't trigger anything
		return
	}

	mod := keyEventModToCodesModifier(evt.Keysym.Mod)

	if (evt.Keysym.Sym > ' ' && evt.Keysym.Sym < 127) &&
		(mod == 0 || mod == codes.MOD_SHIFT) {
		// lets let eventTextInput() handle this so we don't need to think about
		// keyboard layouts and shift chars like whether shift+'2' == '@' or '"'
		return
	}

	switch {
	case evt.Keysym.Sym == sdl.K_F3 && mod == 0:
		sr.term.Search()
		return
	case evt.Keysym.Sym == 'v' && mod == codes.MOD_META:
		sr.clipboardPasteText()
		return
	}

	keyCode := sr.keyCodeLookup(evt.Keysym.Sym)
	b := codes.GetAnsiEscSeq(sr.keyboardMode.Get(), keyCode, mod)
	if len(b) > 0 {
		sr.term.Reply(b)
	}
}

// SDL doesn't name these, so lets name it ourselves for convenience
const (
	_MOUSE_BUTTON_LEFT = 1 + iota
	_MOUSE_BUTTON_MIDDLE
	_MOUSE_BUTTON_RIGHT
	_MOUSE_BUTTON_X1
	_MOUSE_BUTTON_X2
)

func (tw *termWidgetT) eventMouseButton(sr *sdlRender, evt *sdl.MouseButtonEvent) {
	posCell := sr.convertPxToCellXY(evt.X, evt.Y)

	if config.Config.Tmux.Enabled && sr.windowTabs != nil &&
		(evt.Y-_PANE_TOP_MARGIN)/sr.glyphSize.Y == sr.term.GetSize().Y+sr.footer-1 {
		// window tab bar
		if evt.State == sdl.PRESSED {
			return
		}

		x := ((evt.X - _PANE_LEFT_MARGIN) / sr.glyphSize.X) - sr.windowTabs.offset.X
		for i := range sr.windowTabs.boundaries {
			if x < sr.windowTabs.boundaries[i] {
				switch evt.Clicks {
				case 1:
					if i == 0 {
						return
					}
					sr.selectWindow(i - 1)

				default: // 2 or more
					if i == 0 {
						return
					}
					sr.DisplayInputBox("Please enter a new name for this window", sr.windowTabs.windows[i-1].Name, func(name string) {
						err := sr.windowTabs.windows[i-1].Rename(name)
						if err != nil {
							sr.DisplayNotification(types.NOTIFY_ERROR, err.Error())
						}
					})
				}
				return
			}
		}
		if evt.Clicks == 2 {
			sr.tmux.NewWindow()
		}
		return
	}

	state := evt.State == sdl.PRESSED

	if evt.X <= _PANE_LEFT_MARGIN {
		posCell.X = -1

		sr.term.MouseClick(posCell, uint8(evt.Button), evt.Clicks, state, func() {})
		return
	}

	switch evt.Button {
	case _MOUSE_BUTTON_LEFT:
		sr.term.MouseClick(posCell, uint8(evt.Button), evt.Clicks, state, func() {
			if evt.State == sdl.PRESSED {
				highlighterStart(sr, uint8(evt.Button), evt.X, evt.Y)
				sr.highlighter.setMode(_HIGHLIGHT_MODE_LINE_RANGE)
			}
		})

	case _MOUSE_BUTTON_MIDDLE:
		if evt.State == sdl.PRESSED {
			sr.term.MouseClick(posCell, uint8(evt.Button), evt.Clicks, state, sr.clipboardPasteText)
		}

	case _MOUSE_BUTTON_RIGHT:
		sr.contextMenu = make(contextMenuT, 0) // empty the context menu
		sr.term.MouseClick(posCell, uint8(evt.Button), evt.Clicks, state, func() {
			if evt.State == sdl.RELEASED {
				tw._eventMouseButtonRightClick(sr, posCell)
			}
		})

	case _MOUSE_BUTTON_X1:
		sr.term.MouseClick(posCell, uint8(evt.Button), evt.Clicks, state, func() {})
	}
}

func (tw *termWidgetT) _eventMouseButtonRightClick(sr *sdlRender, posCell *types.XY) {
	menu := contextMenuT{
		{
			Title: fmt.Sprintf("Paste text from clipboard [%s+v]", types.KEY_STR_META),
			Fn:    sr.clipboardPasteText,
		},
		{
			Title: MENU_SEPARATOR,
		},
		{
			Title: "Fold on indentation",
			Fn: func() {
				err := sr.term.FoldAtIndent(posCell)
				if err != nil {
					sr.DisplayNotification(types.NOTIFY_WARN, err.Error())
				}
			},
		},
		//{
		//	Title: "Match bracket",
		//	Fn:    func() { sr.term.Match(posCell) },
		//},
		{
			Title: "Search text [F3]",
			Fn:    sr.term.Search,
		},
	}

	if sr.tmux != nil {
		menu = append(menu,
			types.MenuItem{
				Title: MENU_SEPARATOR,
			},
			types.MenuItem{
				Title: "List tmux hotkeys",
				Fn:    sr.tmux.ListKeyBindings,
			},
		)
	}

	if len(sr.contextMenu) > 0 {
		menu = append(menu, types.MenuItem{Title: MENU_SEPARATOR})
		menu = append(menu, sr.contextMenu...)
	}

	menu = append(menu,
		types.MenuItem{
			Title: MENU_SEPARATOR,
		},
		types.MenuItem{
			Title: "Bash integration (pasted into shell)",
			Fn:    func() { sr.term.Reply(integrations.Get("shell.bash")) },
		},
		types.MenuItem{
			Title: "Zsh integration (pasted into shell)",
			Fn:    func() { sr.term.Reply(integrations.Get("shell.zsh")) },
		},
	)

	sr.DisplayMenuUnderCursor("Select an action", menu.Options(), nil, menu.Callback, nil)
}

var _highlighterStartFooterText = fmt.Sprintf(
	"Copy to clipboard: [%s] Square region  |  [%s] Text region  |  [%s] Entire line(s)  |  [%s] PNG",
	types.KEY_STR_CTRL, types.KEY_STR_SHIFT, types.KEY_STR_ALT, types.KEY_STR_META,
)

func highlighterStart(sr *sdlRender, button uint8, x, y int32) {
	sr.footerText = _highlighterStartFooterText

	sr.highlighter = &highlightWidgetT{
		button: button,
		rect:   &sdl.Rect{X: x, Y: y},
	}
	if sr.keyModifier != 0 {
		sr.highlighter.modifier(sr.keyModifier)
	}
	sr.keyModifier = 0
}

func (tw *termWidgetT) eventMouseWheel(sr *sdlRender, evt *sdl.MouseWheelEvent) {
	mouseX, mouseY, _ := sdl.GetMouseState()

	if evt.Direction == sdl.MOUSEWHEEL_FLIPPED {
		sr.term.MouseWheel(sr.convertPxToCellXY(mouseX, mouseY), &types.XY{X: evt.X, Y: -evt.Y})
	} else {
		sr.term.MouseWheel(sr.convertPxToCellXY(mouseX, mouseY), &types.XY{X: evt.X, Y: evt.Y})
	}
}

func (tw *termWidgetT) eventMouseMotion(sr *sdlRender, evt *sdl.MouseMotionEvent) {
	if config.Config.Tmux.Enabled && sr.windowTabs != nil {

		if (evt.Y-_PANE_TOP_MARGIN)/sr.glyphSize.Y == sr.term.GetSize().Y+sr.footer-1 {
			x := ((evt.X - _PANE_LEFT_MARGIN) / sr.glyphSize.X) - sr.windowTabs.offset.X
			for i := range sr.windowTabs.boundaries {
				if x >= 0 && x < sr.windowTabs.boundaries[i] {
					sr.windowTabs.mouseOver = i - 1
					sr.footerText = fmt.Sprintf("[Click]  Switch to window '%s' (%s)", sr.windowTabs.windows[i-1].Name, sr.windowTabs.windows[i-1].Id)
					return
				}
			}
			sr.footerText = "[2x Click]  Start new window"
			sr.windowTabs.mouseOver = -1
			return
		}

		sr.windowTabs.mouseOver = -1
		sr.footerText = ""
	}

	var callback = sr._termMouseMotionCallback
	if evt.State != 0 {
		callback = func() {
			switch evt.State {
			case _MOUSE_BUTTON_LEFT:
				highlighterStart(sr, uint8(evt.State), evt.X-evt.XRel, evt.Y-evt.YRel)
				sr.highlighter.setMode(_HIGHLIGHT_MODE_LINE_RANGE)
			}
		}
	}

	sr.term.MouseMotion(
		sr.convertPxToCellXYNegX(evt.X, evt.Y),
		&types.XY{
			X: evt.XRel / sr.glyphSize.X,
			Y: evt.YRel / sr.glyphSize.Y,
		},
		callback,
	)
}

func (sr *sdlRender) _termMouseMotionCallback() {
	sr.footerText = "[Left Click] Copy  |  [Right Click] Menu  |  [Wheel] Scrollback buffer"
}

func (sr *sdlRender) selectWindow(winIndex int) {
	if winIndex < 0 || winIndex >= len(sr.windowTabs.windows) {
		return
	}

	winId := sr.windowTabs.windows[winIndex].Id
	err := sr.tmux.SelectWindow(winId)
	if err != nil {
		sr.DisplayNotification(types.NOTIFY_ERROR, err.Error())
	}
}

func (sr *sdlRender) RefreshWindowList() {
	if sr.tmux == nil {
		return
	}

	sr.limiter.Lock()

	sr.windowTabs = nil
	sr.term = sr.tmux.ActivePane().Term()

	sr.limiter.Unlock()
}
