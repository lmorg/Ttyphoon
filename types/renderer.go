package types

type NotificationType int

const (
	NOTIFY_DEBUG = iota
	NOTIFY_INFO
	NOTIFY_WARNING
	NOTIFY_ERROR
)

type Renderer interface {
	Start(Term)
	FocusWindow()
	TermSize() *XY
	PrintCell(*Cell, *XY) error
	GetWindowTitle() string
	SetWindowTitle(string)
	Bell()
	TriggerRedraw()
	NewElement(elementType ElementID, size *XY, data []byte) Element
	DisplayNotification(NotificationType, string)
	AddRenderFnToStack(func())
	GetWindowMeta() any
	ResizeWindow(*XY)
	Close()
}

type Image interface {
	Size() *XY
	Asset() any
	Draw(size *XY, rect *Rect)
	Close()
}
