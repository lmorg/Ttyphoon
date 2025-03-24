package config

import (
	"bytes"
	_ "embed"
	"os"
	"strings"

	"github.com/lmorg/mxtty/utils/themes/iterm2"
	"gopkg.in/yaml.v3"
)

/*
	Eventually these will be user configurable rather than compiled time
	options.
*/

//go:embed defaults.yaml
var defaults []byte

func init() {
	err := Default()
	if err != nil {
		panic(err)
	}
}

func Default() error {
	yml := yaml.NewDecoder(bytes.NewReader(defaults))
	yml.KnownFields(true)

	err := yml.Decode(&Config)
	if err != nil {
		return err
	}

	if Config.Terminal.ColorTheme != "" {
		colorTheme := os.ExpandEnv(Config.Terminal.ColorTheme)
		home, _ := os.UserHomeDir()
		colorTheme = strings.ReplaceAll(colorTheme, "~", home)
		err = iterm2.GetTheme(colorTheme)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

var Config configT

type configT struct {
	Tmux struct {
		Enabled bool `yaml:"Enabled"`
	} `yaml:"Tmux"`

	Shell struct {
		Default  []string `yaml:"Default"`
		Fallback []string `yaml:"Fallback"`
	} `yaml:"Shell"`

	Terminal struct {
		ColorTheme              string `yaml:"ColorTheme"`
		ScrollbackHistory       int    `yaml:"ScrollbackHistory"`
		ScrollbackCloseKeyPress bool   `yaml:"ScrollbackCloseKeyPress"`
		JumpScrollLineCount     int    `yaml:"JumpScrollLineCount"`
		LightMode               bool   `yaml:"LightMode"`
		AutoHotlink             bool   `yaml:"AutoHotlink"`

		Widgets struct {
			Table struct {
				ScrollMultiplierX int32 `yaml:"ScrollMultiplierX"`
				ScrollMultiplierY int32 `yaml:"ScrollMultiplierY"`
			} `yaml:"Table"`
		} `yaml:"Widgets"`
	} `yaml:"Terminal"`

	Window struct {
		Opacity         int  `yaml:"Opacity"`
		InactiveOpacity int  `yaml:"InactiveOpacity"`
		StatusBar       bool `yaml:"StatusBar"`
		RefreshInterval int  `yaml:"RefreshInterval"`
		UseGPU          bool `yaml:"UseGPU"`
	} `yaml:"Window"`

	TypeFace struct {
		FontName         string   `yaml:"FontName"`
		FontSize         int      `yaml:"FontSize"`
		Ligatures        bool     `yaml:"Ligatures"`
		LigaturePairs    []string `yaml:"LigaturePairs"`
		DropShadow       bool     `yaml:"DropShadow"`
		AdjustCellWidth  int      `yaml:"AdjustCellWidth"`
		AdjustCellHeight int      `yaml:"AdjustCellHeight"`
	} `yaml:"TypeFace"`
}
