package app

import "charm.land/bubbletea/v2"

type KeyMap struct {
	Quit         tea.Key
	FocusNext    tea.Key
	FocusPrev    tea.Key
	ChannelUp    tea.Key
	ChannelDown  tea.Key
	SelectChan   tea.Key
	ScrollUp     tea.Key
	ScrollDown   tea.Key
	PageUp       tea.Key
	PageDown     tea.Key
	OpenThread   tea.Key
	ClosePanel   tea.Key
	FocusInput   tea.Key
	SendMessage  tea.Key
	NewLine      tea.Key
	QuickSwitch  tea.Key
	Search       tea.Key
	AddReaction  tea.Key
	Help         tea.Key
	AIToggle     tea.Key
	AIDraft      tea.Key
	AISummarize  tea.Key
	NextUnread   tea.Key
	PrevUnread   tea.Key
}

var DefaultKeyMap = KeyMap{
	Quit:         tea.Key{Code: 'c', Mod: tea.ModCtrl},
	FocusNext:    tea.Key{Code: tea.KeyTab},
	FocusPrev:    tea.Key{Code: tea.KeyTab, Mod: tea.ModShift},
	ChannelUp:    tea.Key{Code: 'k'},
	ChannelDown:  tea.Key{Code: 'j'},
	SelectChan:   tea.Key{Code: tea.KeyEnter},
	ScrollUp:     tea.Key{Code: 'k'},
	ScrollDown:   tea.Key{Code: 'j'},
	PageUp:       tea.Key{Code: tea.KeyPgUp},
	PageDown:     tea.Key{Code: tea.KeyPgDown},
	OpenThread:   tea.Key{Code: 't'},
	ClosePanel:   tea.Key{Code: tea.KeyEscape},
	FocusInput:   tea.Key{Code: 'i'},
	SendMessage:  tea.Key{Code: tea.KeyEnter},
	NewLine:      tea.Key{Code: tea.KeyEnter, Mod: tea.ModShift},
	QuickSwitch:  tea.Key{Code: 'k', Mod: tea.ModCtrl},
	Search:       tea.Key{Code: 'f', Mod: tea.ModCtrl},
	AddReaction:  tea.Key{Code: 'r'},
	Help:         tea.Key{Code: '?'},
	AIToggle:     tea.Key{Code: 'a', Mod: tea.ModCtrl},
	AIDraft:      tea.Key{Code: 'd', Mod: tea.ModCtrl},
	AISummarize:  tea.Key{Code: 's', Mod: tea.ModCtrl},
	NextUnread:   tea.Key{Code: 'n', Mod: tea.ModCtrl},
	PrevUnread:   tea.Key{Code: 'p', Mod: tea.ModCtrl},
}
