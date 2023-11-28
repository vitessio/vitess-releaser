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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tbl "github.com/charmbracelet/lipgloss/table"
)

type (
	menu struct {
		items   []menuItem
		title   string
		idx     int
		columns []string
		width   int
	}
	menuItem struct {
		name   string
		state  string
		act    func(menuItem) (menuItem, tea.Cmd)
		init   func() tea.Cmd
		update func(menuItem, tea.Msg) (menuItem, tea.Cmd)
	}
)

var columns = []string{"Task", "Info"}

func newMenu(title string, items ...menuItem) menu {
	return menu{
		columns: columns,
		title:   title,
		items:   items,
	}
}

func (m menu) At(row, cell int) string {
	item := m.items[row]
	if cell == 1 {
		return item.state
	}

	var prefix string
	switch {
	case m.idx != row:
		prefix = "   " // this is not the line we are standing on
	case item.act == nil:
		prefix = "  :" // we are standing on this line, but it has no action
	default:
		prefix = "-> "
	}

	return prefix + item.name
}

func (m menu) Rows() int {
	return len(m.items)
}

func (m menu) Columns() int {
	return 2
}

func (m menu) Init() tea.Cmd {
	var cmds []tea.Cmd
	for idx, mi := range m.items {
		if mi.init != nil {
			cmds = append(cmds, mi.init())
			m.items[idx].init = nil
		}
	}
	return tea.Batch(cmds...)
}

func (m menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	size := len(m.items)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, popDialog
		case "up":
			m.idx = (m.idx - 1 + size) % size
		case "down":
			m.idx = (m.idx + 1) % size
		case "enter":
			selected := m.items[m.idx]
			if selected.act == nil {
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

func (m menu) View() string {
	list := tbl.
		New().
		Width(m.width).
		Headers(m.columns...).
		Data(m).
		StyleFunc(func(row, _ int) lipgloss.Style {
			switch row {
			case 0:
				return headerStyle
			case m.idx + 1:
				return selectedStyle
			default:
				return cellStyle
			}
		}).
		Render()

	return lipgloss.JoinVertical(lipgloss.Left,
		m.title,
		list,
	)
}
