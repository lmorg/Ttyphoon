package tmux

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/lmorg/mxtty/codes"
	"github.com/lmorg/mxtty/debug"
	virtualterm "github.com/lmorg/mxtty/term"
	"github.com/lmorg/mxtty/types"
	runebuf "github.com/lmorg/mxtty/utils/rune_buf"
)

/*
	pane_active                1 if active pane
	pane_at_bottom             1 if pane is at the bottom of window
	pane_at_left               1 if pane is at the left of window
	pane_at_right              1 if pane is at the right of window
	pane_at_top                1 if pane is at the top of window
	pane_bg                    Pane background colour
	pane_bottom                Bottom of pane
	pane_current_command       Current command if available
	pane_current_path          Current path if available
	pane_dead                  1 if pane is dead
	pane_dead_signal           Exit signal of process in dead pane
	pane_dead_status           Exit status of process in dead pane
	pane_dead_time             Exit time of process in dead pane
	pane_fg                    Pane foreground colour
	pane_format                1 if format is for a pane
	pane_height                Height of pane
	pane_id                #D  Unique pane ID
	pane_in_mode               1 if pane is in a mode
	pane_index             #P  Index of pane
	pane_input_off             1 if input to pane is disabled
	pane_key_mode              Extended key reporting mode in this pane
	pane_last                  1 if last pane
	pane_left                  Left of pane
	pane_marked                1 if this is the marked pane
	pane_marked_set            1 if a marked pane is set
	pane_mode                  Name of pane mode, if any
	pane_path                  Path of pane (can be set by application)
	pane_pid                   PID of first process in pane
	pane_pipe                  1 if pane is being piped
	pane_right                 Right of pane
	pane_search_string         Last search string in copy mode
	pane_start_command         Command pane started with
	pane_start_path            Path pane started with
	pane_synchronized          1 if pane is synchronized
	pane_tabs                  Pane tab positions
	pane_title             #T  Title of pane (can be set by application)
	pane_top                   Top of pane
	pane_tty                   Pseudo terminal of pane
	pane_unseen_changes        1 if there were changes in pane while in mode
	pane_width                 Width of pane
*/

var CMD_LIST_PANES = "list-panes"

type PaneT struct {
	Title    string `tmux:"PaneTitle"`
	Id       string `tmux:"pane_id"`
	Width    int    `tmux:"pane_width"`
	Height   int    `tmux:"pane_height"`
	Active   bool   `tmux:"?pane_active,true,false"`
	WindowId string `tmux:"window_id"`
	tmux     *Tmux
	tile     *types.Tile
	buf      *runebuf.Buf
	closed   bool
}

func (tmux *Tmux) initSessionPanes(renderer types.Renderer) error {
	panes, err := tmux.sendCommand(CMD_LIST_PANES, reflect.TypeOf(PaneT{}), "-s")
	if err != nil {
		return err
	}

	for i := range panes.([]any) {
		pane := panes.([]any)[i].(*PaneT)
		pane.tmux = tmux
		pane.tile = new(types.Tile)

		pane.buf = runebuf.New()
		debug.Log(pane)
		tmux.win[pane.WindowId].panes[pane.Id] = pane
		if pane.Active {
			tmux.win[pane.WindowId].activePane = pane
		}
		tmux.pane[pane.Id] = pane

		virtualterm.NewTerminal(pane.tile, renderer, &types.XY{int32(pane.Width), int32(pane.Height)}, false)
		pane.tile.Term.Start(pane)

		command := fmt.Sprintf("capture-pane -J -e -p -t %s", pane.Id)
		resp, err := tmux.SendCommand([]byte(command))
		if err != nil {
			renderer.DisplayNotification(types.NOTIFY_ERROR, err.Error())
		} else {
			b := bytes.Join(resp.Message, []byte{'\r', '\n'}) // CRLF
			pane.buf.Write(b)
		}

		command = fmt.Sprintf(`display-message -p -t %s "#{e|+:#{cursor_y},1};#{e|+:#{cursor_x},1}H"`, pane.Id)
		resp, err = tmux.SendCommand([]byte(command))
		if err != nil {
			renderer.DisplayNotification(types.NOTIFY_ERROR, err.Error())
		} else {
			b := append([]byte{codes.AsciiEscape, '['}, resp.Message[0]...)
			pane.buf.Write(b)
		}
	}

	return nil
}

func (tmux *Tmux) newPane(info *paneInfo) *PaneT {
	pane := &PaneT{
		Id:   info.Id,
		tmux: tmux,
		tile: new(types.Tile),
		buf:  runebuf.New(),
	}

	virtualterm.NewTerminal(
		pane.tile, tmux.renderer,
		&types.XY{X: int32(info.Width), Y: int32(info.Height)},
		false)

	pane.tile.Term.Start(pane)

	tmux.pane[pane.Id] = pane

	return pane
}

type paneInfo struct {
	Id        string `tmux:"pane_id"`
	Title     string `tmux:"pane_title"`
	Width     int    `tmux:"pane_width"`
	Height    int    `tmux:"pane_height"`
	PosLeft   int    `tmux:"pane_left"`
	PosTop    int    `tmux:"pane_top"`
	PosRight  int    `tmux:"pane_right"`
	PosBottom int    `tmux:"pane_bottom"`
	Active    bool   `tmux:"?pane_active,true,false"`
	Dead      bool   `tmux:"?pane_dead,true,false"`
	WindowId  string `tmux:"window_id"`
	WinActive bool   `tmux:"?window_active,true,false"`
}

// updatePaneInfo, paneId is optional. Leave blank to update all panes
func (tmux *Tmux) updatePaneInfo(paneId string) error {
	var filter string
	if paneId != "" {
		filter = fmt.Sprintf("-f '#{m:#{pane_id},%s}'", paneId)
	}

	v, err := tmux.sendCommand(CMD_LIST_PANES, reflect.TypeOf(paneInfo{}), "-s", filter)
	if err != nil {
		return err
	}

	panes, ok := v.([]any)
	if !ok {
		return fmt.Errorf("expecting an array of panes, instead got %T", v)
	}

	for i := range panes {
		info, ok := panes[i].(*paneInfo)
		if !ok {
			return fmt.Errorf("expecting info on a pane, instead got %T", info)
		}

		pane, ok := tmux.pane[info.Id]
		if !ok {
			pane = tmux.newPane(info)
		}

		/*if info.Dead {
			pane.Close()
			continue
		}*/

		if pane.closed {
			continue
		}

		pane.Title = info.Title
		pane.Width = info.Width
		pane.Height = info.Height
		pane.Active = info.Active
		pane.WindowId = info.WindowId
		pane.tile.PaneId = info.Id
		pane.tile.Left = int32(info.PosLeft)
		pane.tile.Top = int32(info.PosTop)
		pane.tile.Right = int32(info.PosRight)
		pane.tile.Bottom = int32(info.PosBottom)
		if pane.tile.Term != nil {
			pane.tile.Term.MakeVisible(info.WinActive)
			pane.tile.Term.HasFocus(info.Active)
			pane.tile.Term.Resize(&types.XY{X: int32(info.Width), Y: int32(info.Height)})
		}

		tmux.win[pane.WindowId].panes[pane.Id] = pane
		if pane.Active {
			tmux.win[pane.WindowId].activePane = pane
		}
	}

	return nil
}

func (tmux *Tmux) ActivePane() *PaneT {
	return tmux.activeWindow.activePane
}

func (p *PaneT) Term() types.Term {
	return p.tile.Term
}

func (tmux *Tmux) SelectPane(paneId string) error {
	command := fmt.Sprintf("select-pane -t %s", paneId)
	_, err := tmux.SendCommand([]byte(command))

	go tmux.UpdateSession()

	return err
}

func (tmux *Tmux) paneExited() error {
	v, err := tmux.sendCommand(CMD_LIST_PANES, reflect.TypeOf(paneInfo{}), "-s")
	if err != nil {
		return err
	}

	panes, ok := v.([]any)
	if !ok {
		return fmt.Errorf("expecting an array of panes, instead got %T", v)
	}

	for _, pane := range tmux.pane {
		pane.closed = true
	}

	for i := range panes {
		info, ok := panes[i].(*paneInfo)
		if !ok {
			return fmt.Errorf("expecting info on a pane, instead got %T", info)

		}

		pane, ok := tmux.pane[info.Id]
		if !ok {
			continue // ignore new panes
		}

		pane.closed = false
	}

	go tmux.renderer.RefreshWindowList()

	return nil
}
