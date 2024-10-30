//go:build !no_cgo && !linux && !(windows && amd64) && !darwin
// +build !no_cgo
// +build !linux
// +build !windows !amd64
// +build !darwin

/*
	This file uses the C SQLite3 library. To compile it you will need gcc
	installed as well as Go. This is why it is disabled by default, with the
	pure Go driver favoured instead.

	However any pre-built binaries available on Murex's website will be
	compiled against this C library instead.
*/

package elementCsv

import (
	_ "github.com/mattn/go-sqlite3"
)

const driverName = "sqlite3"
