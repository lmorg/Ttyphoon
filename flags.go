package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	fAnsiImage string
)

const (
	ANSI_BEGIN = "\x1bPtmux;\x1b\x1b\x1b_begin;%s;%s\x07\x1b\\"
	ANSI_END   = "\x1bPtmux;\x1b\x1b\x1b_end;%s\x07\x1b\\"
	//ANSI_INSERT = "\x1bPtmux;\x1b\x1b\x1b_insert;%s;%s\x07\x1b\\"
)

func getFlags() {
	flag.StringVar(&fAnsiImage, "image", "", "")
	flag.Parse()

	switch {
	case fAnsiImage != "":
		ansiImage()
	}
}

func ansiImage() {
	f, err := os.Open(fAnsiImage)
	die(err)

	b, err := io.ReadAll(f)
	die(err)

	b64 := base64.StdEncoding.EncodeToString(b)

	die(err)
	fmt.Printf(ANSI_BEGIN+strings.Repeat(".\n", 15)+ANSI_END, "image", params(map[string]any{
		"base64": b64,
	}), "image")
	os.Exit(0)
}

func die(err error) {
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func params(params map[string]any) string {
	b, err := json.Marshal(params)
	die(err)
	return string(b)
}
