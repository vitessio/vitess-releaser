/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package interactive

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tbl "github.com/charmbracelet/lipgloss/table"
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
)

type (
	menu struct {
		state   *releaser.State
		items   []*menuItem
		title   string
		idx     int
		columns []string
		width   int
	}

	menuItem struct {
		state  *releaser.State
		name   string
		isDone bool
		info   string
		act    func(*menuItem) (*menuItem, tea.Cmd)
		init   func(*menuItem) tea.Cmd
		update func(*menuItem, tea.Msg) (*menuItem, tea.Cmd)

		// subItems is a slice of *menuItem referring to the menuItem embedded by this item
		subItems []*menuItem

		previous            *menuItem
		dontCountInProgress bool
	}
)

var columns = []string{"TASK", "STATUS", "INFO"}

func newMenu(ctx context.Context, title string, items ...*menuItem) *menu {
	for i, item := range items {
		if i == 0 {
			continue
		}
		item.previous = items[i-1]
	}
	return &menu{
		state:   releaser.UnwrapState(ctx),
		columns: columns,
		title:   title,
		items:   items,
	}
}

func (m *menu) moveCursorToNextElem() {
	for _, item := range m.items {
		if item.isDone || item.name == "" {
			m.idx++
		} else {
			break
		}
	}
}

func (m *menu) At(row, cell int) string {
	item := m.items[row]
	if item.name == "" {
		return ""
	}
	if cell == 1 {
		if len(item.subItems) > 0 {
			done := 0
			for _, subItem := range item.subItems {
				if subItem.isDone {
					done++
				}
			}
			nb := len(item.subItems)
			if done == nb {
				item.isDone = state.Done
			}
			if !item.isDone {
				return fmt.Sprintf("%s %d/%d", state.Fmt(item.isDone), done, nb)
			}
			msg := fmt.Sprintf("%s %d/%d", state.Fmt(item.isDone), done, nb)
			if item.isDone {
				msg += " \U0001f44d"
			}
			return msg
		}

		// if there are no sub items, let's just return the current status
		msg := state.Fmt(item.isDone)
		if item.isDone {
			msg += " \U0001f44d"
		}
		return msg
	}
	if cell == 2 {
		return item.info
	}

	var prefix string
	switch {
	case m.idx != row:
		prefix = "   " // this is not the line we are standing on
	case item.isActBlocked():
		prefix = "\U0001f512 " // we are standing on this line, but it has no action
	default:
		prefix = "\U0001f449 "
	}

	return prefix + item.name
}

func (m *menu) Rows() int {
	return len(m.items)
}

func (m *menu) Columns() int {
	return 3
}

func (m *menu) Init() tea.Cmd {
	var cmds []tea.Cmd
	for idx, mi := range m.items {
		if mi.init != nil {
			cmds = append(cmds, mi.init(mi))
			m.items[idx].init = nil
		}
	}
	return tea.Batch(cmds...)
}

func (m *menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	size := len(m.items)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, popDialog
		case "up":
			for {
				m.idx = (m.idx - 1 + size) % size
				if m.items[m.idx].name != "" {
					break
				}
			}
		case "down":
			for {
				m.idx = (m.idx + 1) % size
				if m.items[m.idx].name != "" {
					break
				}
			}
		case "enter":
			selected := m.items[m.idx]
			if selected.isActBlocked() {
				return m, nil
			}
			var cmd tea.Cmd
			m.items[m.idx], cmd = selected.act(selected)
			return m, cmd
		}
	default:
		var cmds []tea.Cmd
		for idx, mi := range m.items {
			if mi.update != nil {
				newMi, cmd := mi.update(mi, msg)
				m.items[idx] = newMi
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (mi *menuItem) isActBlocked() bool {
	if mi.act == nil || mi.isDone {
		return true
	}

	currMenuItem := mi.previous
	for currMenuItem != nil && currMenuItem.name == "" || !currMenuItem.dontCountInProgress {
		if currMenuItem.previous != nil {
			currMenuItem = currMenuItem.previous
		} else {
			break
		}
	}
	return currMenuItem != nil && !currMenuItem.isDone
}

func (m *menu) View() string {
	list := tbl.
		New().
		Width(m.width).
		Headers(m.columns...).
		Data(m).
		Border(lipgloss.ThickBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) (s lipgloss.Style) {
			switch row {
			case 0:
				s = headerStyle
			case m.idx + 1:
				s = selectedStyle
			default:
				s = cellStyle
			}
			switch col {
			case 0:
				s = s.Copy().MaxWidth(3)
			case 1:
				s = s.Copy().MaxWidth(1)
			}
			return
		}).
		Render()

	return lipgloss.JoinVertical(lipgloss.Left,
		m.title,
		list,
	)
}
