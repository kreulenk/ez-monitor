package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Quit           key.Binding
	Previous       key.Binding
	Next           key.Binding
	HistoricalView key.Binding
	LiveView       key.Binding
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m Model) HelpView() string {
	if m.activeView == LiveData {
		return m.Help.ShortHelpView([]key.Binding{keys.Quit, keys.Previous, keys.Next, keys.HistoricalView})
	} else {
		return m.Help.ShortHelpView([]key.Binding{keys.Quit, keys.Previous, keys.Next, keys.LiveView})
	}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Previous: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "previous host"),
	),
	Next: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "next host"),
	),
	LiveView: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "live data"),
	),
	HistoricalView: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "historical data"),
	),
}
