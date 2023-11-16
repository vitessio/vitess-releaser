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

type checkListItem struct {
	name string
	done bool
}

func newCheckListItem(name string) *checkListItem {
	return &checkListItem{
		name: name,
	}
}

func (i *checkListItem) FilterValue() string { return "" }

type checkListItemDelegate struct{}

func (d checkListItemDelegate) Height() int                             { return 1 }
func (d checkListItemDelegate) Spacing() int                            { return 0 }
func (d checkListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d checkListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*checkListItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.name)
	if i.done {
		str = fmt.Sprintf("%s: done ✅", str)
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
