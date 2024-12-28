package tmux

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lmorg/mxtty/debug"
	"github.com/lmorg/mxtty/types"
	"github.com/lmorg/mxtty/utils/octal"
)

func (p *PaneT) File() *os.File      { return nil }
func (p *PaneT) Read() (rune, error) { return p.buf.Read() }
func (p *PaneT) Close()              { p.buf.Close() }

func (p *PaneT) Write(b []byte) error {
	if len(b) == 0 {
		return errors.New("nothing to write")
	}

	ok, err := p._hotkey(b)
	if ok {
		if err != nil {
			p.tmux.renderer.DisplayNotification(types.NOTIFY_ERROR, err.Error())
		}
		return nil
	}

	var flags string
	if b[0] == 0 {
		b = b[1:]
	} else {
		flags = "-l"
		b = octal.Escape(b)
	}

	command := []byte(fmt.Sprintf(`send-keys %s -t %s `, flags, p.Id))
	command = append(command, b...)
	_, err = p.tmux.SendCommand(command)
	return err
}

func (p *PaneT) _hotkey(b []byte) (bool, error) {
	var key string
	if b[0] == 0 {
		key = string(b[1 : len(b)-1])
	} else {
		key = string(b)
	}
	debug.Log(key)

	if p.prefixTtl.Before(time.Now()) {
		if key != p.tmux.keys.prefix {
			// standard key, do nothing
			return false, nil
		}

		// prefix key pressed
		p.prefixTtl = time.Now().Add(2 * time.Second)
		return true, nil
	}

	// run tmux function
	fn, ok := p.tmux.keys.fnTable[key]
	debug.Log(ok)
	if !ok {
		// no function to run, lets treat as standard key
		p.prefixTtl = time.Now()
		return false, nil
	}

	return true, fn.fn(p.tmux)
}
