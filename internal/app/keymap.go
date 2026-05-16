package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Tab1     key.Binding
	Tab2     key.Binding
	Tab3     key.Binding
	Tab4     key.Binding
	Tab5     key.Binding
	Tab6     key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Add      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Toggle   key.Binding
	Ping     key.Binding
	Test     key.Binding
	Promote  key.Binding
	MoveUp   key.Binding
	MoveDown key.Binding
	Filter   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Tab1:     key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "overview")),
		Tab2:     key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "messages")),
		Tab3:     key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "chat")),
		Tab4:     key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "codex")),
		Tab5:     key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "gemini")),
		Tab6:     key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "more")),
		NextTab:  key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "next tab")),
		PrevTab:  key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "prev tab")),
		Up:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
		Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("↵", "select")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Toggle:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "toggle")),
		Ping:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "ping")),
		Test:     key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "test")),
		Promote:  key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "promote")),
		MoveUp:   key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "move up")),
		MoveDown: key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "move down")),
		Filter:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Help, k.Tab1, k.Tab2, k.Tab3, k.Tab4, k.Tab5, k.Tab6}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab1, k.Tab2, k.Tab3, k.Tab4, k.Tab5, k.Tab6},
		{k.Up, k.Down, k.Enter, k.Back},
		{k.Add, k.Edit, k.Delete, k.Toggle},
		{k.Ping, k.Test, k.Filter, k.Help},
	}
}
