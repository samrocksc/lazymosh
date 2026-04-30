package pkg

// screen is the active screen identifier.
type Screen int

const (
	ScreenList Screen = iota
	ScreenAdd
	ScreenEdit
)

var ScreenNames = map[Screen]string{
	ScreenList: "Servers",
	ScreenAdd:  "Add Server",
	ScreenEdit: "Edit Server",
}

// NavigateMsg switches to a different screen.
type NavigateMsg struct {
	Screen Screen
	Server any // config.Server — received as any to avoid config import here
	Index  int
}

// ReloadMsg tells the root to reload servers and return to list.
type ReloadMsg struct{}

// ConnectErrMsg is returned when a connection attempt fails.
type ConnectErrMsg struct{ Err error }

// SaveErrMsg is returned when a screen's save/delete fails.
type SaveErrMsg struct{ Err error }

// SaveSuccessMsg is returned when a screen's save/delete succeeds.
// The root handles this and navigates back to the list, which re-reads
// the config file from disk so added/edited/deleted servers appear.
type SaveSuccessMsg struct{}
