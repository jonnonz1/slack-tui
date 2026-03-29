package sidebar

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// testKeyMsg implements tea.KeyMsg for tests.
type testKeyMsgType struct {
	str string
}

func (t testKeyMsgType) String() string { return t.str }
func (t testKeyMsgType) Key() tea.Key   { return tea.Key{Code: rune(t.str[0])} }
func (t testKeyMsgType) Format(f fmt.State, c rune) {
	fmt.Fprint(f, t.str)
}

func testKeyMsg(s string) tea.KeyMsg {
	return testKeyMsgType{str: s}
}
