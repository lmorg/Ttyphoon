package tmux

import (
	"reflect"

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

var CMD_LIST_WINDOWS = &cmdDefinitionT{
	cmd: "list-windows",
	fields: []cmdFieldT{
		{
			name:   "Name",
			format: "window_name",
		},
		{
			name:   "Id",
			format: "window_id",
		},
		{
			name:   "Width",
			format: "window_width",
		},
		{
			name:   "Height",
			format: "window_height",
		},
		{
			name:   "Active",
			format: "?window_active,true,false",
		},
	},
}

type windowT struct {
	Name       string
	Id         string
	Width      int
	Height     int
	Active     bool
	panes      map[string]*paneT
	activePane *paneT
}

func (tmux *Tmux) initSessionWindows() error {
	windows, err := tmux.sendCommand(CMD_LIST_WINDOWS, reflect.TypeOf(windowT{}))
	if err != nil {
		return err
	}

	tmux.wins = make(map[string]*windowT)

	for i := range windows.([]any) {
		win := windows.([]any)[i].(*windowT)
		tmux.wins[win.Id] = win
		if win.Active {
			tmux.activeWindow = win
		}
	}

	debug.Log(windows.([]any))
	return nil
}
