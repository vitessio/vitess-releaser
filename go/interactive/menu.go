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

type menu struct {
	items   []menuItem
	title   string
	idx     int
	columns []string
}

func (m menu) At(row, cell int) string {
	item := m.items[row]
	if cell == 1 {
		return item.state
	}
	prefix := "   "
	if m.idx == row {
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

type menuItem struct {
	name  string
	state string
	act   func(menuItem) (menuItem, tea.Cmd)
}

func (m menu) Init() tea.Cmd { return nil }

func (m menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	size := len(m.items)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, pop
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.idx = (m.idx - 1 + size) % size
		case "down":
			m.idx = (m.idx + 1) % size
		case "enter":
			selected := m.items[m.idx]
			newItem, cmd := selected.act(selected)
			m.items[m.idx] = newItem
			return m, cmd
		}
	}

	return m, nil
}

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
	black    = lipgloss.Color("#00000")
)

var (
	cellStyle     = lipgloss.NewStyle().Foreground(darkGray)
	selectedStyle = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle   = lipgloss.NewStyle().Foreground(black)
)

func (m menu) View() string {
	list := tbl.
		New().
		Width(100).
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

func newMenu(title string, columns []string, items []menuItem) menu {
	return menu{
		columns: columns,
		title:   title,
		items:   items,
	}
}
