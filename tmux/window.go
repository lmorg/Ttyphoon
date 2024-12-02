package tmux

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/lmorg/mxtty/debug"
)

/*
	window_active                1 if window active
	window_active_clients        Number of clients viewing this window
	window_active_clients_list   List of clients viewing this window
	window_active_sessions       Number of sessions on which this window is active
	window_active_sessions_list  List of sessions on which this window is active
	window_activity              Time of window last activity
	window_activity_flag         1 if window has activity
	window_bell_flag             1 if window has bell
	window_bigger                1 if window is larger than client
	window_cell_height           Height of each cell in pixels
	window_cell_width            Width of each cell in pixels
	window_end_flag              1 if window has the highest index
	window_flags             #F  Window flags with # escaped as ##
	window_format                1 if format is for a window
	window_height                Height of window
	window_id                    Unique window ID
	window_index             #I  Index of window
	window_last_flag             1 if window is the last used
	window_layout                Window layout description, ignoring zoomed window panes
	window_linked                1 if window is linked across sessions
	window_linked_sessions       Number of sessions this window is linked to
	window_linked_sessions_list  List of sessions this window is linked to
	window_marked_flag           1 if window contains the marked pane
	window_name              #W  Name of window
	window_offset_x              X offset into window if larger than client
	window_offset_y              Y offset into window if larger than client
	window_panes                 Number of panes in window
	window_raw_flags             Window flags with nothing escaped
	window_silence_flag          1 if window has silence alert
	window_stack_index           Index in session most recent stack
	window_start_flag            1 if window has the lowest index
	window_visible_layout        Window layout description, respecting zoomed window panes
	window_width                 Width of window
	window_zoomed_flag           1 if window is zoomed
*/

var CMD_LIST_WINDOWS = "list-windows"

type WINDOW_T struct {
	Name       string `tmux:"window_name"`
	Id         string `tmux:"window_id"`
	Index      int    `tmux:"window_index"`
	Width      int    `tmux:"window_width"`
	Height     int    `tmux:"window_height"`
	Active     bool   `tmux:"?window_active,true,false"`
	panes      map[string]*PANE_T
	activePane *PANE_T
	closed     bool
}

func (tmux *Tmux) initSessionWindows() error {
	windows, err := tmux.sendCommand(CMD_LIST_WINDOWS, reflect.TypeOf(WINDOW_T{}))
	if err != nil {
		return err
	}

	tmux.win = make(map[string]*WINDOW_T)

	for i := range windows.([]any) {
		win := windows.([]any)[i].(*WINDOW_T)
		win.panes = make(map[string]*PANE_T)
		tmux.win[win.Id] = win
		if win.Active {
			tmux.activeWindow = win
		}

		command := fmt.Sprintf("set-option -w -t %s window-size latest", win.Id)
		_, _ = tmux.SendCommand([]byte(command))
		//if err != nil {
		//	return err
		//}
	}

	debug.Log(windows.([]any))
	return nil
}

func (tmux *Tmux) newWindow(winId string) *WINDOW_T {
	win := &WINDOW_T{
		Id:    winId,
		panes: make(map[string]*PANE_T),
	}

	tmux.win[winId] = win
	//tmux.activeWindow = win
	return win
}

type winInfo struct {
	Id     string `tmux:"window_id"`
	Index  int    `tmux:"window_index"`
	Name   string `tmux:"window_name"`
	Width  int    `tmux:"window_width"`
	Height int    `tmux:"window_height"`
	Active bool   `tmux:"?window_active,true,false"`
}

// updateWinInfo, winId is optional. Leave blank to update all windows
func (tmux *Tmux) updateWinInfo(winId string) error {
	var filter string
	if winId != "" {
		filter = fmt.Sprintf("-f '#{m:#{window_id},%s}'", winId)
	}

	v, err := tmux.sendCommand(CMD_LIST_WINDOWS, reflect.TypeOf(winInfo{}), filter)
	if err != nil {
		return err
	}

	wins, ok := v.([]any)
	if !ok {
		return fmt.Errorf("expecting an array of windows, instead got %T", v)
	}

	for i := range wins {

		info, ok := wins[i].(*winInfo)
		if !ok {
			return fmt.Errorf("expecting info on a window, instead got %T", info)
		}

		win, ok := tmux.win[info.Id]
		if !ok {
			win = tmux.newWindow(info.Id)
		}
		win.Index = info.Index
		win.Name = info.Name
		win.Width = info.Width
		win.Height = info.Height
		win.Active = info.Active

		if win.Active {
			tmux.activeWindow = win
		}
	}

	return nil
}

func (tmux *Tmux) RenderWindows() []*WINDOW_T {
	var wins []*WINDOW_T

	for _, win := range tmux.win {
		if win.closed {
			continue
		}
		wins = append(wins, win)
	}

	sort.Slice(wins, func(i, j int) bool {
		return wins[i].Index < wins[j].Index
	})

	return wins
}

func (win *WINDOW_T) ActivePane() *PANE_T {
	return win.activePane
}

func (win *WINDOW_T) Rename(name string) error {
	command := fmt.Sprintf("rename-window -t %s '%s'", win.Id, name)
	_, err := win.activePane.tmux.SendCommand([]byte(command))
	return err
}

func (tmux *Tmux) SelectWindow(winId string) error {
	size := tmux.renderer.GetWindowSizeCells()
	command := fmt.Sprintf("resize-window -t %s -x %d -y %d", winId, size.X, size.Y)
	_, _ = tmux.SendCommand([]byte(command))
	/*if err != nil {
		p.Width = int(size.X)
		p.Height = int(size.Y)
		return err
	}*/

	command = fmt.Sprintf("select-window -t %s", winId)
	_, err := tmux.SendCommand([]byte(command))
	if err != nil {
		return err
	}

	go tmux.UpdateSession()

	return err
}
