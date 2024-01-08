package typeface

import (
	"github.com/flopp/go-findfont"
	"github.com/lmorg/mxtty/types"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var fontSize *types.XY

func init() {
	err := ttf.Init()
	if err != nil {
		panic(err.Error())
	}
}

func Close() {
	ttf.Quit()
}

func Open(name string, size int) (*ttf.Font, error) {
	path, err := findfont.Find(name)
	if err != nil {
		panic(err)
	}

	font, err := ttf.OpenFont(path, size)
	if err != nil {
		return nil, err
	}

	font.SetHinting(ttf.HINTING_MONO)

	fontSize, err = getSize(font)
	return font, err
}

func GetSize() *types.XY {
	return fontSize
}

func getSize(font *ttf.Font) (*types.XY, error) {
	surface, err := font.RenderGlyphSolid('W', sdl.Color{R: 0, G: 0, B: 0, A: 255})
	if err != nil {
		return nil, err
	}
	return &types.XY{
		X: surface.W,
		Y: surface.H,
	}, nil
}
