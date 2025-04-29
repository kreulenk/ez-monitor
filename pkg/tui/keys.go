package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the help menu.
type keyMap struct {
	Quit       key.Binding
	Previous   key.Binding
	Next       key.Binding
	ViewToggle key.Binding
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m Model) HelpView() string {
	return m.Help.ShortHelpView([]key.Binding{keys.Quit, keys.Previous, keys.Next, keys.ViewToggle})
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
	ViewToggle: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "view toggle"),
	),
}
