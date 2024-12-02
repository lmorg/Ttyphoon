package types

type MenuCallbackT func(int)

type Renderer interface {
	Start(Term, any)
	ShowAndFocusWindow()
	GetTermSize() *XY
	GetWindowSizeCells() *XY
	GetGlyphSize() *XY
	PrintCell(*Cell, *XY)
	PrintCellBlock([]Cell, *XY)
	DrawTable(*XY, int32, []int32)
	DrawHighlightRect(*XY, *XY)
	GetWindowTitle() string
	SetWindowTitle(string)
	StatusBarText(string)
	RefreshWindowList()
	Bell()
	TriggerRedraw()
	TriggerQuit()
	NewElement(elementType ElementID) Element
	DisplayNotification(NotificationType, string)
	DisplaySticky(NotificationType, string) Notification
	DisplayInputBox(string, string, func(string))
	DisplayMenu(string, []string, MenuCallbackT, MenuCallbackT, MenuCallbackT)
	GetWindowMeta() any
	ResizeWindow(*XY)
	SetKeyboardFnMode(KeyboardMode)
	ClearFontCache()
	Close()
}

type Image interface {
	Size() *XY
	Asset() any
	Draw(size *XY, pos *XY)
	Close()
}
