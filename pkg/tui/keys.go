package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Quit     key.Binding
	Previous key.Binding
	Next     key.Binding
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m Model) HelpView() string {
	return m.Help.View(keys)
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.Quit, km.Previous, km.Next}
}

// FullHelp is needed to satisfy the keyMap interface
func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		km.ShortHelp(),
	}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Previous: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Next: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
}
