package iterm2

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/lmorg/mxtty/debug"
	"github.com/lmorg/mxtty/types"
)

// GetTheme loads an iTerm2 .plist theme and returns a map of colors
func GetTheme(filename string) error {
	// Open the plist file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	theme, err := unmarshalTheme(file)
	if err != nil {
		return err
	}

	return convertToMxttyTheme(theme)
}

func convertToMxttyTheme(theme map[string]Color) error {
	for name, color := range theme {
		var err error
		switch name {
		case "Ansi 0 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_BLACK)
		case "Ansi 1 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_RED)
		case "Ansi 2 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_GREEN)
		case "Ansi 3 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_YELLOW)
		case "Ansi 4 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_BLUE)
		case "Ansi 5 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_MAGENTA)
		case "Ansi 6 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_CYAN)
		case "Ansi 7 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_WHITE)
		case "Ansi 8 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_BLACK_BRIGHT)
		case "Ansi 9 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_RED_BRIGHT)
		case "Ansi 10 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_GREEN_BRIGHT)
		case "Ansi 11 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_YELLOW_BRIGHT)
		case "Ansi 12 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_BLUE_BRIGHT)
		case "Ansi 13 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_MAGENTA_BRIGHT)
		case "Ansi 14 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_CYAN_BRIGHT)
		case "Ansi 15 Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_WHITE_BRIGHT)
		case "Background Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_BACKGROUND)
		case "Foreground Color":
			err = rgbRealToByteColor(color, types.SGR_COLOR_FOREGROUND)
		case "Selection Color":
			err = rgbRealToByteColor(color, types.COLOR_SELECTION)

		default:
			debug.Log("skipping component: " + name)
		}

		if err != nil {
			return fmt.Errorf("invalid component '%s': %v", name, err)
		}
	}

	return nil
}

func rgbRealToByteColor(rCol Color, bCol *types.Colour) error {
	if rCol.Red > 1 || rCol.Green > 1 || rCol.Blue > 1 {
		return errors.New("rgb value > 1")
	}

	if rCol.Red < 0 || rCol.Green < 0 || rCol.Blue < 0 {
		return errors.New("rgb value < 0")
	}

	bCol.Red = byte(math.Round(rCol.Red * 255))
	bCol.Green = byte(math.Round(rCol.Green * 255))
	bCol.Blue = byte(math.Round(rCol.Blue * 255))
	bCol.Alpha = byte(math.Round(rCol.Alpha * 255))
	if bCol.Alpha == 0 {
		bCol.Alpha = 255
	}
	
	return nil
}
