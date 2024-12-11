package rendersdl

import (
	"time"

	"github.com/lmorg/mxtty/codes"
	"github.com/lmorg/mxtty/types"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var (
	questionColor       = &types.Colour{0x00, 0x77, 0x00}
	questionColorBorder = &types.Colour{0x00, 0xff, 0x00}
)

type inputBoxCallbackT func(string)

type inputBoxT struct {
	title      string
	callback   inputBoxCallbackT
	value      string
	blinkState bool
}

func (sr *sdlRender) DisplayInputBox(title string, defaultValue string, callback func(string)) {
	sr.inputBox = &inputBoxT{
		title:    title,
		value:    defaultValue,
		callback: callback,
	}

	sr.footerText = "[Return] Ok  |  [Esc] Cancel  |  [Ctrl+u] Clear text"
	sr.term.ShowCursor(false)
	go sr.inputBox.inputBoxCursorBlink(sr)
}

func (sr *sdlRender) closeInputBox() {
	sr.footerText = ""
	sr.inputBox = nil
	sr.term.ShowCursor(true)
}

func (inputBox *inputBoxT) eventTextInput(sr *sdlRender, evt *sdl.TextInputEvent) {
	inputBox.value += evt.GetText()
}

func (inputBox *inputBoxT) eventKeyPress(sr *sdlRender, evt *sdl.KeyboardEvent) {
	mod := keyEventModToCodesModifier(evt.Keysym.Mod)
	switch evt.Keysym.Sym {
	case sdl.K_ESCAPE:
		sr.closeInputBox()
	case sdl.K_RETURN:
		sr.closeInputBox()
		inputBox.callback(inputBox.value)
	case sdl.K_BACKSPACE:
		if inputBox.value != "" {
			inputBox.value = inputBox.value[:len(inputBox.value)-1]
		} else {
			sr.Bell()
		}
	case sdl.K_u:
		if mod == codes.MOD_CTRL {
			inputBox.value = ""
		}
	}
}

func (inputBox *inputBoxT) eventMouseButton(sr *sdlRender, evt *sdl.MouseButtonEvent) {
	// do nothing
}

func (inputBox *inputBoxT) eventMouseWheel(sr *sdlRender, evt *sdl.MouseWheelEvent) {
	// do nothing
}

func (inputBox *inputBoxT) eventMouseMotion(sr *sdlRender, evt *sdl.MouseMotionEvent) {
	// do nothing
}

const _INPUTBOX_MAX_CHARS = int32(75)

func (sr *sdlRender) renderInputBox(windowRect *sdl.Rect) {
	surface, err := sdl.CreateRGBSurfaceWithFormat(0, windowRect.W, windowRect.H, 32, uint32(sdl.PIXELFORMAT_RGBA32))
	if err != nil {
		panic(err) //TODO: don't panic!
	}
	defer surface.Free()

	sr.font.SetStyle(ttf.STYLE_BOLD)

	text, err := sr.font.RenderUTF8BlendedWrapped(sr.inputBox.title, sdl.Color{R: 200, G: 200, B: 200, A: 255}, int(surface.W-sr.notifyIconSize.X))
	if err != nil {
		panic(err) // TODO: don't panic!
	}
	defer text.Free()

	textShadow, err := sr.font.RenderUTF8BlendedWrapped(sr.inputBox.title, sdl.Color{R: 0, G: 0, B: 0, A: 150}, int(surface.W-sr.notifyIconSize.X))
	if err != nil {
		panic(err) // TODO: don't panic!
	}
	defer textShadow.Free()

	/*
		FRAME
	*/

	padding := int32(10)
	height := text.H + (padding * 3) + sr.glyphSize.Y
	maxLen := int32(len(sr.inputBox.title))
	if maxLen < _INPUTBOX_MAX_CHARS {
		maxLen = _INPUTBOX_MAX_CHARS
	}
	width := sr.glyphSize.X*maxLen + sr.notifyIconSize.X + padding
	offsetH := (surface.H / 2) - (height / 2)
	offsetY := (surface.W - width) / 2

	// draw border
	sr.renderer.SetDrawColor(questionColorBorder.Red, questionColorBorder.Green, questionColorBorder.Blue, notificationAlpha)
	rect := sdl.Rect{
		X: offsetY - 1,
		Y: offsetH - 1,
		W: width + 2,
		H: height + 2,
	}
	sr.renderer.DrawRect(&rect)
	rect = sdl.Rect{
		X: offsetY,
		Y: offsetH,
		W: width,
		H: height,
	}
	sr.renderer.DrawRect(&rect)

	// fill background
	sr.renderer.SetDrawColor(questionColor.Red, questionColor.Green, questionColor.Blue, notificationAlpha)
	rect = sdl.Rect{
		X: offsetY + 1,
		Y: 1 + offsetH,
		W: width - 2,
		H: height - 2,
	}
	sr.renderer.FillRect(&rect)

	// render shadow
	rect = sdl.Rect{
		X: offsetY + padding + sr.notifyIconSize.X + 2,
		Y: sr.border + offsetH + 2,
		W: width - sr.notifyIconSize.X,
		H: text.H + padding - 2,
	}
	_ = textShadow.Blit(nil, surface, &rect)
	sr._renderNotificationSurface(surface, &rect)

	// render text
	rect = sdl.Rect{
		X: offsetY + padding + sr.notifyIconSize.X,
		Y: sr.border + offsetH,
		W: width - sr.notifyIconSize.X,
		H: text.H + padding - 2,
	}
	err = text.Blit(nil, surface, &rect)
	if err != nil {
		panic(err) // TODO: don't panic!
	}
	sr._renderNotificationSurface(surface, &rect)

	/*
		TEXT FIELD
	*/

	height = sr.glyphSize.Y + (sr.border * 2)
	offsetH += text.H + sr.border + sr.border
	var textWidth int32

	// draw border
	sr.renderer.SetDrawColor(255, 255, 255, 150)
	rect = sdl.Rect{
		X: offsetY + sr.notifyIconSize.X + padding - 1,
		Y: offsetH - 1,
		W: width - sr.notifyIconSize.X - padding - padding + 2,
		H: height + 2,
	}
	sr.renderer.DrawRect(&rect)
	rect = sdl.Rect{
		X: offsetY + sr.notifyIconSize.X + padding,
		Y: offsetH,
		W: width - sr.notifyIconSize.X - padding - padding,
		H: height,
	}
	sr.renderer.DrawRect(&rect)

	// fill background
	sr.renderer.SetDrawColor(0, 0, 0, 150)
	rect = sdl.Rect{
		X: offsetY + sr.notifyIconSize.X + padding + 1,
		Y: 1 + offsetH,
		W: width - sr.notifyIconSize.X - padding - padding - 2,
		H: height - 2,
	}
	sr.renderer.FillRect(&rect)

	// value
	if len(sr.inputBox.value) > 0 {
		textValue, err := sr.font.RenderUTF8Blended(sr.inputBox.value, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		defer textValue.Free()

		rect = sdl.Rect{
			X: offsetY + padding + sr.notifyIconSize.X + sr.border,
			Y: sr.border + offsetH,
			W: surface.W - sr.notifyIconSize.X,
			H: textValue.H + padding - 2,
		}
		err = textValue.Blit(nil, surface, &rect)
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		sr._renderNotificationSurface(surface, &rect)
		textWidth = textValue.W
	}

	if surface, ok := sr.notifyIcon[types.NOTIFY_QUESTION].Asset().(*sdl.Surface); ok {
		srcRect := &sdl.Rect{
			X: 0,
			Y: 0,
			W: width,
			H: surface.H,
		}

		dstRect := &sdl.Rect{
			X: offsetY + (padding / 2),
			Y: offsetH + text.H - sr.notifyIconSize.Y,
			W: sr.notifyIconSize.X,
			H: sr.notifyIconSize.X,
		}

		texture, err := sr.renderer.CreateTextureFromSurface(surface)
		if err != nil {
			panic(err) //TODO: don't panic!
		}
		defer texture.Destroy()

		err = sr.renderer.Copy(texture, srcRect, dstRect)
		if err != nil {
			panic(err) //TODO: don't panic!
		}
	}

	if sr.inputBox.blinkState {
		rect = sdl.Rect{
			X: offsetY + padding + sr.notifyIconSize.X + sr.border + textWidth,
			Y: sr.border + offsetH,
			W: sr.glyphSize.X,
			H: sr.glyphSize.Y,
		}
		sr._drawHighlightRect(&rect, 255, 200)
	}

	texture, err := sr.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err) // TODO: better error handling please!
	}
	defer texture.Destroy()

	err = sr.renderer.Copy(texture, windowRect, windowRect)
	if err != nil {
		panic(err) // TODO: better error handling please!
	}
}

func (inputBox *inputBoxT) inputBoxCursorBlink(sr *sdlRender) {
	for {
		time.Sleep(500 * time.Millisecond)
		inputBox.blinkState = !inputBox.blinkState
		if sr.inputBox == nil {
			return
		}
	}
}
