package rendersdl

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lmorg/mxtty/assets"
	"github.com/lmorg/mxtty/types"
	"github.com/veandco/go-sdl2/sdl"
)

var notifyColour = map[int]*types.Colour{
	types.NOTIFY_DEBUG: {0x1c, 0x3e, 0x64},
	types.NOTIFY_INFO:  {0x31, 0x6d, 0xb0},
	types.NOTIFY_WARN:  {0x74, 0x58, 0x10},
	types.NOTIFY_ERROR: {0x66, 0x16, 0x1a},
}

var notifyBorderColour = map[int]*types.Colour{
	types.NOTIFY_DEBUG: {0x31, 0x6d, 0xb0},
	types.NOTIFY_INFO:  {0x99, 0xc0, 0xd3},
	types.NOTIFY_WARN:  {0xf2, 0xb7, 0x1f},
	types.NOTIFY_ERROR: {0xde, 0x33, 0x3b},
}

func (sr *sdlRender) preloadNotificationGlyphs() {
	var err error
	sr.notifyIcon = map[int]types.Image{
		types.NOTIFY_DEBUG: nil,
		types.NOTIFY_INFO:  nil,
		types.NOTIFY_WARN:  nil,
		types.NOTIFY_ERROR: nil,
	}
	sr.notifyIconSize = &types.XY{
		X: sr.glyphSize.Y + (sr.border * 4),
		Y: sr.glyphSize.Y + (sr.border * 4),
	}

	sr.notifyIcon[types.NOTIFY_DEBUG], err = sr.loadImage(assets.Get(assets.ICON_DEBUG), sr.notifyIconSize)
	if err != nil {
		panic(err)
	}

	sr.notifyIcon[types.NOTIFY_INFO], err = sr.loadImage(assets.Get(assets.ICON_INFO), sr.notifyIconSize)
	if err != nil {
		panic(err)
	}

	sr.notifyIcon[types.NOTIFY_WARN], err = sr.loadImage(assets.Get(assets.ICON_WARN), sr.notifyIconSize)
	if err != nil {
		panic(err)
	}

	sr.notifyIcon[types.NOTIFY_ERROR], err = sr.loadImage(assets.Get(assets.ICON_ERROR), sr.notifyIconSize)
	if err != nil {
		panic(err)
	}
}

type notifyT struct {
	stack []*notificationT
	mutex sync.Mutex
}

type notificationT struct {
	Type    types.NotificationType
	Message string
	wait    <-chan time.Time
	end     time.Time
}

func (n *notifyT) _wait() {
	for {
		if len(n.stack) == 0 {
			return
		}

		<-n.stack[0].wait
		n.remove()
	}
}

func (n *notifyT) add(notification *notificationT) {
	d := 5 * time.Second
	notification.end = time.Now().Add(d)
	notification.wait = time.After(d)

	n.mutex.Lock()
	n.stack = append(n.stack, notification)

	if len(n.stack) > 0 {
		go n._wait()
	}
	n.mutex.Unlock()

	log.Printf("NOTIFICATION: %s", notification.Message)
}

func (n *notifyT) remove() {
	n.mutex.Lock()
	n.stack = n.stack[1:]
	n.mutex.Unlock()
}

func (n *notifyT) get() []*notificationT {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if len(n.stack) == 0 {
		return nil
	}

	notifications := make([]*notificationT, len(n.stack))
	copy(notifications, n.stack)

	return notifications
}

func (sr *sdlRender) DisplayNotification(notificationType types.NotificationType, message string) {
	notification := &notificationT{
		Type:    notificationType,
		Message: message,
	}
	sr.notifications.add(notification)
}

func (sr *sdlRender) renderNotification(windowRect *sdl.Rect) {
	notifications := sr.notifications.get()
	if notifications == nil {
		return
	}

	surface, err := sdl.CreateRGBSurfaceWithFormat(0, windowRect.W, windowRect.H, 32, uint32(sdl.PIXELFORMAT_RGBA32))
	if err != nil {
		panic(err) //TODO: don't panic!
	}
	defer surface.Free()

	sr.setFontStyle(types.SGR_BOLD)

	padding := sr.border * 2
	var offset int32
	for _, notification := range notifications {
		// generate text
		s := strconv.Itoa(int(time.Until(notification.end)/time.Second) + 1)
		countdown, err := sr.font.RenderUTF8Blended(s, sdl.Color{R: 255, G: 255, B: 255, A: 200})
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		defer countdown.Free()

		text, err := sr.font.RenderUTF8BlendedWrapped(notification.Message, sdl.Color{R: 200, G: 200, B: 200, A: 255}, int(sr.surface.W-sr.notifyIconSize.X-countdown.W))
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		defer text.Free()

		textShadow, err := sr.font.RenderUTF8BlendedWrapped(notification.Message, sdl.Color{R: 0, G: 0, B: 0, A: 150}, int(sr.surface.W-sr.notifyIconSize.X-countdown.W))
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		defer textShadow.Free()

		// draw border
		bc := notifyBorderColour[int(notification.Type)]
		sr.renderer.SetDrawColor(bc.Red, bc.Green, bc.Blue, 190)
		rect := sdl.Rect{
			X: sr.border - 1,
			Y: sr.border + offset - 1,
			W: windowRect.W - padding + 2,
			H: text.H + padding + 2,
		}
		sr.renderer.DrawRect(&rect)
		rect = sdl.Rect{
			X: sr.border,
			Y: sr.border + offset,
			W: windowRect.W - padding,
			H: text.H + padding,
		}
		sr.renderer.DrawRect(&rect)

		// fill background
		c := notifyColour[int(notification.Type)]
		sr.renderer.SetDrawColor(c.Red, c.Green, c.Blue, 190)
		rect = sdl.Rect{
			X: sr.border + 1,
			Y: sr.border + 1 + offset,
			W: sr.surface.W - padding - 2,
			H: text.H + padding - 2,
		}
		sr.renderer.FillRect(&rect)

		// render countdown
		rect = sdl.Rect{
			X: windowRect.W - padding - countdown.W,
			Y: padding + offset,
			W: countdown.W,
			H: countdown.H,
		}
		err = countdown.Blit(nil, surface, &rect)
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		sr._renderNotificationSurface(surface, &rect)

		// render shadow
		rect = sdl.Rect{
			X: padding + sr.notifyIconSize.X + 2,
			Y: padding + offset + 2,
			W: sr.surface.W - sr.notifyIconSize.X - countdown.W,
			H: text.H + padding - 2,
		}
		_ = textShadow.Blit(nil, surface, &rect)
		sr._renderNotificationSurface(surface, &rect)

		// render text
		rect = sdl.Rect{
			X: padding + sr.notifyIconSize.X,
			Y: padding + offset,
			W: sr.surface.W - sr.notifyIconSize.X - countdown.W,
			H: text.H + padding - 2,
		}
		err = text.Blit(nil, surface, &rect)
		if err != nil {
			panic(err) // TODO: don't panic!
		}
		sr._renderNotificationSurface(surface, &rect)

		if surface, ok := sr.notifyIcon[int(notification.Type)].Asset().(*sdl.Surface); ok {
			srcRect := &sdl.Rect{
				X: 0,
				Y: 0,
				W: surface.W,
				H: surface.H,
			}

			dstRect := &sdl.Rect{
				X: sr.border * 2,
				Y: offset + ((text.H + padding + padding + 2) / 2) - (sr.notifyIconSize.Y / 2),
				W: sr.notifyIconSize.X,
				H: sr.notifyIconSize.X,
			}

			texture, err := sr.renderer.CreateTextureFromSurface(surface)
			if err != nil {
				panic(err) //TODO: don't panic!
			}

			err = sr.renderer.Copy(texture, srcRect, dstRect)
			if err != nil {
				panic(err) //TODO: don't panic!
			}
		}

		offset += text.H + (sr.border * 3)
	}
}

func (sr *sdlRender) _renderNotificationSurface(surface *sdl.Surface, rect *sdl.Rect) {
	texture, err := sr.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err) //TODO: don't panic!
	}

	err = sr.renderer.Copy(texture, rect, rect)
	if err != nil {
		panic(err) //TODO: don't panic!
	}
}
