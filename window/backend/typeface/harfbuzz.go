package typeface

import (
	"image"
	"log"
	"regexp"
	"unsafe"

	"github.com/go-text/render"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/fontscan"
	"github.com/lmorg/mxtty/assets"
	"github.com/lmorg/mxtty/config"
	"github.com/lmorg/mxtty/types"
	"github.com/veandco/go-sdl2/sdl"
)

var _FONT_FAMILIES = []string{"monospace", "emoji", "math", "fantasy"}

type styleT int

const (
	_STYLE_NORMAL styleT = 0
	_STYLE_BOLD   styleT = 1 << iota
	_STYLE_ITALIC
	_STYLE_FAINT
	_STYLE_UNDERLINE
	_STYLE_STRIKETHROUGH
	_STYLE_FONT_AWESOME
)

func (f styleT) Is(flag styleT) bool {
	return f&flag != 0
}

func (f styleT) TypeFace() styleT {
	return f &^ (_STYLE_UNDERLINE | _STYLE_STRIKETHROUGH)
}

type fontHarfbuzz struct {
	size  *types.XY
	face  map[styleT]*font.Face
	style styleT
	fsize float32
	fmap  *fontscan.FontMap
	sdl   *fontSdl // just needed while I figure out how to do everything in harfbuzz
}

func (f *fontHarfbuzz) Init() error {
	f.face = make(map[styleT]*font.Face)
	f.fsize = float32(config.Config.TypeFace.FontSize)
	f.fmap = fontscan.NewFontMap(log.Default())

	err := f.fmap.UseSystemFonts("")
	if err != nil {
		return err
	}

	f.fmap.SetQuery(fontscan.Query{Families: _FONT_FAMILIES})

	f.sdl = new(fontSdl)
	return err
}

func (f *fontHarfbuzz) Open(name string, size int) (err error) {
	f.openAsset(assets.FONT_AWESOME, _STYLE_FONT_AWESOME)

	if name != "" {
		_FONT_FAMILIES = append([]string{name}, _FONT_FAMILIES...)
		name = f.fmap.FontLocation(f.getFace('W').Font).File
		return f.sdl.Open(name, size)
	}

	f.openAsset(assets.TYPEFACE, _STYLE_NORMAL)
	f.openAsset(assets.TYPEFACE_I, _STYLE_ITALIC)
	f.openAsset(assets.TYPEFACE_B, _STYLE_BOLD)
	f.openAsset(assets.TYPEFACE_BI, _STYLE_BOLD|_STYLE_ITALIC)
	f.openAsset(assets.TYPEFACE_L, _STYLE_FAINT)
	f.openAsset(assets.TYPEFACE_LI, _STYLE_FAINT|_STYLE_ITALIC)

	rx := regexp.MustCompile(`[-.]`)
	fontName := rx.Split(assets.TYPEFACE, 2)
	_FONT_FAMILIES = append(fontName[:1], _FONT_FAMILIES...)

	return f.sdl.Open("", size)
}

func (f *fontHarfbuzz) openAsset(name string, style styleT) {
	var (
		res font.Resource
		err error
	)

	res = assets.Reader(name)
	f.face[style], err = font.ParseTTF(res)
	if err != nil {
		panic(err)
	}

	f.fmap.AddFace(f.face[style], fontscan.Location{}, f.face[style].Describe())
}

func (f *fontHarfbuzz) getSize() *types.XY {
	return f.sdl.size
}

func (f *fontHarfbuzz) SetStyle(style types.SgrFlag) {
	query := fontscan.Query{Families: _FONT_FAMILIES}
	f.style = _STYLE_NORMAL

	if style.Is(types.SGR_ITALIC) {
		f.style |= _STYLE_ITALIC
		query.Aspect.Style = font.StyleItalic
	}

	if style.Is(types.SGR_BOLD) {
		f.style |= _STYLE_BOLD
		query.Aspect.Weight = font.WeightBold
	}

	if style.Is(types.SGR_FAINT) {
		f.style |= _STYLE_FAINT
		query.Aspect.Weight = font.WeightLight
	}

	if style.Is(types.SGR_SPECIAL_FONT_AWESOME) {
		f.style = _STYLE_FONT_AWESOME
	}

	if style.Is(types.SGR_UNDERLINE) {
		f.style |= _STYLE_UNDERLINE
	}

	if style.Is(types.SGR_STRIKETHROUGH) {
		f.style |= _STYLE_STRIKETHROUGH
	}

	f.fmap.SetQuery(query)
}

func (f *fontHarfbuzz) getFace(ch rune) *font.Face {
	style := f.style.TypeFace()

	/*if style >= _STYLE_FONT_AWESOME || f.style >= _STYLE_FONT_AWESOME || ch == '' {
		panic(style)
	}*/

	if f.face[style] != nil && f.glyphIsProvided(ch) {
		return f.face[style]
	}

	return f.fmap.ResolveFace(ch)
}

// RenderGlyph should be called from a font atlas
func (f *fontHarfbuzz) RenderGlyphs(fg *types.Colour, cellRect *sdl.Rect, ch ...rune) (*sdl.Surface, error) {
	textWidth := cellRect.W * int32(len(ch))

	img := image.NewNRGBA(image.Rect(0, 0, int(textWidth), int(cellRect.H)))

	textRenderer := &render.Renderer{
		FontSize: f.fsize,
		Color:    fg,
	}

	_ = textRenderer.DrawString(string(ch), img, f.getFace(ch[0]))

	surface, err := sdl.CreateRGBSurfaceWithFormatFrom(
		unsafe.Pointer(&img.Pix[0]),
		textWidth, cellRect.H,
		32, textWidth*4, uint32(sdl.PIXELFORMAT_RGBA32),
	)

	if err != nil || !f.style.Is(_STYLE_UNDERLINE) {
		return surface, err
	}

	renderer, err := sdl.CreateSoftwareRenderer(surface)
	if err != nil {
		return nil, err
	}

	size := f.getSize()
	y := size.Y - 1

	if f.style.Is(_STYLE_UNDERLINE) {
		_ = renderer.SetDrawColor(fg.Red, fg.Green, fg.Blue, fg.Alpha)
		_ = renderer.DrawLine(0, y, size.X, y)
	}

	if f.style.Is(_STYLE_STRIKETHROUGH) {
		_ = renderer.SetDrawColor(fg.Red, fg.Green, fg.Blue, fg.Alpha)
		_ = renderer.DrawLine(0, y, size.X, y)
	}

	renderer.Present()
	return surface, nil
}

func (f *fontHarfbuzz) glyphIsProvided(r rune) bool {
	_, found := f.face[f.style.TypeFace()].NominalGlyph(r)
	return found
}
