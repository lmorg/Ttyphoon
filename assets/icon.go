package assets

import "embed"

const ICON_APP = "icon.bmp"

//go:embed icon.bmp
var embedFsIcons embed.FS

func init() {

	embedFs := embedFsIcons

	dir, err := embedFs.ReadDir(".")
	if err != nil {
		panic(err)
	}

	for i := range dir {
		name := dir[i].Name()

		b, err := embedFs.ReadFile(name)
		if err != nil {
			// not a bug in murex
			panic(err)
		}

		assets[name] = b
	}
}
