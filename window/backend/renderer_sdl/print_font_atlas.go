package rendersdl

import (
	"strings"

	"github.com/lmorg/mxtty/config"
	"github.com/lmorg/mxtty/types"
	"github.com/lmorg/mxtty/window/backend/renderer_sdl/layer"
	"github.com/lmorg/mxtty/window/backend/typeface"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var (
	fontAtlasCharacterPreload = []string{
		" ",                          // whitespace
		"1234567890",                 // numeric
		"abcdefghijklmnopqrstuvwxyz", // alpha, lower
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ", // alpha, upper
		"`",                          // backtick
		`!"£$%^&*()-=_+`,             // special, top row
		`[]{};'#:@~\|,./<>?`,         // special, others ascii
		`↑↓←→`,                       // special, mxtty
		`»…`,                         // special, murex
		`┏┓┗┛━─╶┃┠┨╔╗╚╝═║╟╢█`, //     // box drawing
	}
)

type fontCacheDefaultLookupT map[rune]*sdl.Rect
type fontAtlasT struct {
	sgrHash uint64
	lookup  fontCacheDefaultLookupT
	texture []*sdl.Texture
}
type fontTextureLookupTableT map[rune][]*fontAtlasT
type fontCacheT struct {
	atlas    *fontAtlasT
	extended fontTextureLookupTableT
	ligs     fontTextureLookupTableT
}

func NewFontCache(sr *sdlRender) *fontCacheT {
	chars := []rune(strings.Join(fontAtlasCharacterPreload, ""))

	fc := &fontCacheT{
		atlas:    newFontAtlas(chars, types.SGR_DEFAULT, sr.glyphSize, sr.renderer, sr.font, _FONT_ATLAS_NOT_LIG),
		extended: make(fontTextureLookupTableT),
		ligs:     make(fontTextureLookupTableT),
	}

	return fc
}

const _FONT_ATLAS_NOT_LIG = -1

func newFontAtlas(chars []rune, sgr *types.Sgr, glyphSize *types.XY, renderer *sdl.Renderer, font *ttf.Font, ligId rune) *fontAtlasT {
	glyphSizePlusShadow := &types.XY{
		X: glyphSize.X + dropShadowOffset,
		Y: glyphSize.Y + dropShadowOffset,
	}

	fa := &fontAtlasT{sgrHash: sgr.HashValue()}

	if ligId == _FONT_ATLAS_NOT_LIG {
		fa.newFontCacheDefaultLookup(chars, glyphSizePlusShadow)
		fa.texture = []*sdl.Texture{
			_HLTEXTURE_NONE:          fa.newFontTexture(chars, sgr, glyphSizePlusShadow, renderer, font, _HLTEXTURE_NONE),
			_HLTEXTURE_SELECTION:     fa.newFontTexture(chars, sgr, glyphSizePlusShadow, renderer, font, _HLTEXTURE_SELECTION),
			_HLTEXTURE_SEARCH_RESULT: fa.newFontTexture(chars, sgr, glyphSizePlusShadow, renderer, font, _HLTEXTURE_SEARCH_RESULT),
		}

	} else {
		fa.newFontCacheDefaultLookup([]rune{ligId}, glyphSizePlusShadow)
		fa.texture = []*sdl.Texture{
			_HLTEXTURE_NONE:          fa.newLigTexture(chars, sgr, glyphSizePlusShadow, renderer, font, _HLTEXTURE_NONE),
			_HLTEXTURE_SELECTION:     fa.newLigTexture(chars, sgr, glyphSizePlusShadow, renderer, font, _HLTEXTURE_SELECTION),
			_HLTEXTURE_SEARCH_RESULT: fa.newLigTexture(chars, sgr, glyphSizePlusShadow, renderer, font, _HLTEXTURE_SEARCH_RESULT),
		}
	}

	return fa
}

func (fa *fontAtlasT) newFontCacheDefaultLookup(chars []rune, glyphSize *types.XY) {
	fa.lookup = make(fontCacheDefaultLookupT)

	for i, r := range chars {
		fa.lookup[r] = &sdl.Rect{
			X: glyphSize.X * int32(i),
			Y: 0,
			W: glyphSize.X,
			H: glyphSize.Y,
		}
	}
}

func (fa *fontAtlasT) newFontTexture(chars []rune, sgr *types.Sgr, glyphSize *types.XY, renderer *sdl.Renderer, font *ttf.Font, hlTexture int) *sdl.Texture {
	surface := newFontSurface(glyphSize, int32(len(chars)))
	defer surface.Free()

	cellRect := &sdl.Rect{W: glyphSize.X, H: glyphSize.Y}

	font.SetStyle(fontStyle(sgr.Bitwise))
	typeface.SetStyle(sgr.Bitwise)

	var i int
	var err error
	for i = range chars {
		cellRect.X = int32(i) * glyphSize.X
		err = fa.printCellsToSurface(sgr, cellRect, surface, hlTexture, chars[i])
		if err != nil {
			panic(err) // TODO: better error handling please!
		}
	}

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err) // TODO: better error handling please!
	}

	//defer texture.Destroy() // we don't want to destroy!

	return texture
}

func (fa *fontAtlasT) newLigTexture(chars []rune, sgr *types.Sgr, glyphSize *types.XY, renderer *sdl.Renderer, font *ttf.Font, hlTexture int) *sdl.Texture {
	surface := newFontSurface(glyphSize, int32(len(chars)))
	defer surface.Free()

	cellRect := &sdl.Rect{W: glyphSize.X * int32(len(chars)), H: glyphSize.Y}

	font.SetStyle(fontStyle(sgr.Bitwise))
	typeface.SetStyle(sgr.Bitwise)

	err := fa.printCellsToSurface(sgr, cellRect, surface, hlTexture, chars...)
	if err != nil {
		panic(err) // TODO: better error handling please!
	}

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err) // TODO: better error handling please!
	}

	//defer texture.Destroy() // we don't want to destroy!

	return texture
}

func (fa *fontAtlasT) printCellsToSurface(sgr *types.Sgr, cellRect *sdl.Rect, surface *sdl.Surface, hlTexture int, chars ...rune) error {
	fg, bg := sgrOpts(sgr, hlTexture == _HLTEXTURE_SELECTION)

	// render background colour

	if bg != nil {
		var pixel uint32
		if hlTexture != 0 {
			pixel = sdl.MapRGBA(surface.Format, textShadow[hlTexture].Red, textShadow[hlTexture].Green, textShadow[hlTexture].Blue, 255)
		} else {
			pixel = sdl.MapRGBA(surface.Format, bg.Red, bg.Green, bg.Blue, 255)
		}
		fillRect := &sdl.Rect{
			X: cellRect.X,
			Y: cellRect.Y,
			W: cellRect.W - dropShadowOffset + 10, //,
			H: cellRect.H - dropShadowOffset + 10, //-1,
		}
		err := surface.FillRect(fillRect, pixel)
		if err != nil {
			return err
		}
	}

	// render drop shadow

	if (config.Config.TypeFace.DropShadow && bg == nil && !typeface.GlyphIsEmoji(chars[0])) ||
		hlTexture > _HLTEXTURE_SELECTION {

		shadowText, err := typeface.RenderGlyphs(textShadow[hlTexture], cellRect, chars...)
		if err != nil {
			return err
		}
		defer shadowText.Free()

		shadowRect := &sdl.Rect{
			X: cellRect.X + dropShadowOffset,
			Y: cellRect.Y + dropShadowOffset,
			W: cellRect.W,
			H: cellRect.H,
		}
		_ = shadowText.Blit(nil, surface, shadowRect)

		if hlTexture > _HLTEXTURE_SELECTION {
			shadowRect = &sdl.Rect{
				X: cellRect.X - dropShadowOffset,
				Y: cellRect.Y + dropShadowOffset,
				W: cellRect.W,
				H: cellRect.H,
			}
			_ = shadowText.Blit(nil, surface, shadowRect)
			shadowRect = &sdl.Rect{
				X: cellRect.X - dropShadowOffset,
				Y: cellRect.Y - dropShadowOffset,
				W: cellRect.W,
				H: cellRect.H,
			}
			_ = shadowText.Blit(nil, surface, shadowRect)
			shadowRect = &sdl.Rect{
				X: cellRect.X + dropShadowOffset,
				Y: cellRect.Y - dropShadowOffset,
				W: cellRect.W,
				H: cellRect.H,
			}
			_ = shadowText.Blit(nil, surface, shadowRect)
		}
	}

	// render cell char
	text, err := typeface.RenderGlyphs(fg, cellRect, chars...)
	if err != nil {
		return err
	}
	defer text.Free()

	if hlTexture != 0 {
		_ = text.SetBlendMode(sdl.BLENDMODE_ADD)
	}

	err = text.Blit(nil, surface, cellRect)
	if err != nil {
		return err
	}
	/*if config.Config.TypeFace.Ligatures && cell.Sgr.Bitwise.Is(types.SGR_BOLD) {
		_ = text.SetBlendMode(sdl.BLENDMODE_ADD)
		_ = text.Blit(nil, surface, cellRect)
	}*/

	return nil
}

func (fa *fontAtlasT) Render(sr *sdlRender, dstRect *sdl.Rect, r rune, hash uint64, hlMode int) bool {
	if hash != fa.sgrHash {
		//debug.Log("fontAtlasT: hash not found")
		return false
	}

	srcRect, ok := fa.lookup[r]
	if !ok {
		//debug.Log("fontAtlasT: srcRect not found")
		return false
	}

	texture := fa.texture[hlMode]

	sr.AddToElementStack(&layer.RenderStackT{texture, srcRect, dstRect, false})

	return true
}
