package tasklist

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	//	PageUp     key.Binding
	//	PageDown   key.Binding
	GoToTop    key.Binding
	GoToBottom key.Binding

	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	SetInProgress key.Binding
	SetComplete   key.Binding
	ClearStatus   key.Binding

	Quit      key.Binding
	ForceQuit key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		//PageUp: key.NewBinding(
		//	key.WithKeys("left", "h", "pgup", "b", "u"),
		//	key.WithHelp("←/h/pgup", "page up"),
		//),
		//PageDown: key.NewBinding(
		//	key.WithKeys("right", "l", "pgdown", "f", "d"),
		//	key.WithHelp("→/l/pgdn", "next page"),
		//),
		GoToTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		//SetInProgress key.Binding
		//SetDone       key.Binding
		//ClearStatus   key.Binding

		SetInProgress: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "set in progress"),
		),
		SetComplete: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "set completed"),
		),
		ClearStatus: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear status"),
		),
		//SetInComplete: key.NewBinding(key.WithHelp("c"), key.WithHelp("c", "set complete")),
		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
