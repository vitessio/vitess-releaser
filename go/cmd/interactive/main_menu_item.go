package interactive

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type mainMenuItem struct {
	name  string
	items []list.Item
}

func newItem(name string, items []list.Item) mainMenuItem {
	return mainMenuItem{
		name:  name,
		items: items,
	}
}

func (i mainMenuItem) FilterValue() string { return "" }

func (i mainMenuItem) act() tea.Model {
	return getCheckList(i.items)
}

type mainMenuItemDelegate struct{}

func (d mainMenuItemDelegate) Height() int                             { return 1 }
func (d mainMenuItemDelegate) Spacing() int                            { return 0 }
func (d mainMenuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d mainMenuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(mainMenuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
