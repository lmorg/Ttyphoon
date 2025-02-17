package typeface

import (
	"github.com/lmorg/mxtty/config"
	"github.com/lmorg/mxtty/types"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type typefaceRenderer interface {
	Init() error
	Open(string, int) error
	GetSize() *types.XY
	SetStyle(types.SgrFlag)
	RenderGlyph(rune, *types.Colour, *sdl.Rect) (*sdl.Surface, error)
	glyphIsProvided(int, rune) bool
	Deprecated_GetFont() *ttf.Font
	Close()
}

var renderer typefaceRenderer

func Init() error {
	if config.Config.TypeFace.HarfbuzzRenderer {
		renderer = new(fontHarfbuzz)
	} else {
		renderer = new(fontSdl)
	}

	return renderer.Init()
}

func Open(name string, size int) (err error) {
	return renderer.Open(name, size)
}

func GetSize() *types.XY {
	return renderer.GetSize()
}

func SetStyle(style types.SgrFlag) {
	renderer.SetStyle(style)
}

func RenderGlyph(char rune, fg *types.Colour, cellRect *sdl.Rect) (*sdl.Surface, error) {
	return renderer.RenderGlyph(char, fg, cellRect)
}

func Close() {
	ttf.Quit()
}

func Deprecated_GetFont() *ttf.Font {
	return renderer.Deprecated_GetFont()
}

/*
func ligSplitSequence(runes []rune) [][]rune {
	var (
		seq [][]rune
		i   int
	)

	for _, r := range runes {
		if renderer.glyphIsProvided(_FONT_DEFAULT, r) {
			seq[i] = append(seq[i], r)
			continue
		}

		if len(seq[i]) == 0 {
			seq[i] = append(seq[i], r)
			seq = append(seq, []rune{})
			i++
		} else {
			seq = append(seq, []rune{r}, []rune{})
			i += 2
		}
	}

	return seq
}
*/
