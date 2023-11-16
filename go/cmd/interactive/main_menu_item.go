/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
