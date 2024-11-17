package ptty

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/creack/pty"
	"github.com/lmorg/mxtty/types"
	runebuf "github.com/lmorg/mxtty/utils/rune_buf"
)

type PTY struct {
	primary   *os.File
	secondary *os.File
	buf       *runebuf.Buf
}

func NewPTY(size *types.XY) (types.Pty, error) {
	secondary, primary, err := pty.Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open pty: %s", err.Error())
	}

	err = pty.Setsize(primary, &pty.Winsize{
		Cols: uint16(size.X),
		Rows: uint16(size.Y),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to set pty size: %s", err.Error())
	}

	p := &PTY{
		primary:   primary,
		secondary: secondary,
		buf:       runebuf.New(),
	}

	go p.read(secondary)

	return p, err
}

func (p *PTY) File() *os.File {
	return p.primary
}

func (p *PTY) Write(b []byte) error {
	_, err := p.secondary.Write(b)
	return err
}

func (p *PTY) read(f *os.File) {
	for {
		b := make([]byte, 10*1024)
		i, err := f.Read(b)
		if err != nil && err.Error() != io.EOF.Error() {
			log.Printf("ERROR: problem reading from PTY (%d bytes dropped): %v", i, err)
			continue
		}

		p.buf.Write(b[:i])
	}
}

func (p *PTY) Read() rune {
	return p.buf.Read()
}

func (p *PTY) Resize(size *types.XY) error {
	return pty.Setsize(p.File(), &pty.Winsize{
		Cols: uint16(size.X),
		Rows: uint16(size.Y),
	})
}
