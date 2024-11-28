package octal

import (
	"fmt"
	"strconv"
)

func Escape(b []byte) []byte {
	var escaped []byte

	for _, c := range b {
		escaped = append(escaped, []byte(fmt.Sprintf(`\%03o`, c))...)
	}

	return escaped
}

func Unescape(b []byte) []byte {
	var (
		c = make([]byte, len(b))
		j int
	)

	for i := 0; i < len(b); j++ {
		if b[i] != '\\' {
			c[j] = b[i]
			i++
			continue
		}

		if b[i+1] == '\\' {
			c[j] = '\\'
			i += 2
			continue
		}

		parseInt, err := strconv.ParseInt(string(b[i+1:i+4]), 8, 64)
		if err != nil {
			panic(err)
		}
		c[j] = byte(parseInt)
		i += 4
	}

	return c[:j]
}
