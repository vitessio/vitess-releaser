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

package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tbl "github.com/charmbracelet/lipgloss/table"
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type (
	Menu struct {
		state   *releaser.State
		Items   []*MenuItem
		title   string
		idx     int
		columns []string
		width   int
	}

	MenuItem struct {
		State  *releaser.State
		Name   string
		IsDone bool
		Info   string
		Act    func(*MenuItem) (*MenuItem, tea.Cmd)
		Init   func(*MenuItem) tea.Cmd
		Update func(*MenuItem, tea.Msg) (*MenuItem, tea.Cmd)

		// SubItems is a slice of *MenuItem referring to the MenuItem embedded by this item
		SubItems []*MenuItem

		previous            *MenuItem
		DontCountInProgress bool
	}
)

var columns = []string{"TASK", "STATUS", "INFO"}

func NewMenu(ctx context.Context, title string, items ...*MenuItem) *Menu {
	for i, item := range items {
		if i == 0 {
			continue
		}
		item.previous = items[i-1]
	}
	return &Menu{
		state:   releaser.UnwrapState(ctx),
		columns: columns,
		title:   title,
		Items:   items,
	}
}

func (m *Menu) MoveCursorToNextElem() {
	for i, item := range m.Items {
		// We do a little special handling if the Check PRs/Issues task is listed.
		// We skip it only if the next task is marked as done too.
		// That way, it is not the first task the user has to do.
		if item.IsDone || item.Name == "" || item.Name == steps.CheckAndAdd && m.isNextTaskDone(i+1) {
			m.idx++
		} else {
			break
		}
	}
}

func (m *Menu) isNextTaskDone(i int) bool {
	// we skip all placeholder Menu items
	for ; i < len(m.Items); i++ {
		if m.Items[i].Name != "" {
			break
		}
	}
	if i == len(m.Items) {
		return true
	}
	return m.Items[i].IsDone
}

func (m *Menu) At(row, cell int) string {
	item := m.Items[row]
	if item.Name == "" {
		return ""
	}
	if cell == 1 {
		if len(item.SubItems) > 0 {
			done := 0
			for _, subItem := range item.SubItems {
				if subItem.IsDone {
					done++
				}
			}
			nb := len(item.SubItems)
			if done == nb {
				item.IsDone = state.Done
			}
			if !item.IsDone {
				return fmt.Sprintf("%s %d/%d", state.Fmt(item.IsDone), done, nb)
			}
			msg := fmt.Sprintf("%s %d/%d", state.Fmt(item.IsDone), done, nb)
			if item.IsDone {
				msg += " \U0001f44d"
			}
			return msg
		}

		// if there are no sub items, let's just return the current status
		msg := state.Fmt(item.IsDone)
		if item.IsDone {
			msg += " \U0001f44d"
		}
		return msg
	}
	if cell == 2 {
		return item.Info
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

	return prefix + item.Name
}

func (m *Menu) Rows() int {
	return len(m.Items)
}

func (m *Menu) Columns() int {
	return 3
}

func (m *Menu) Init() tea.Cmd {
	var cmds []tea.Cmd
	for idx, mi := range m.Items {
		if mi.Init != nil {
			cmds = append(cmds, mi.Init(mi))
			m.Items[idx].Init = nil
		}
	}
	return tea.Batch(cmds...)
}

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	size := len(m.Items)
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
				if m.Items[m.idx].Name != "" {
					break
				}
			}
		case "down":
			for {
				m.idx = (m.idx + 1) % size
				if m.Items[m.idx].Name != "" {
					break
				}
			}
		case "enter":
			selected := m.Items[m.idx]
			if selected.isActBlocked() {
				return m, nil
			}
			var cmd tea.Cmd
			m.Items[m.idx], cmd = selected.Act(selected)
			return m, cmd
		}
	default:
		var cmds []tea.Cmd
		for idx, mi := range m.Items {
			if mi.Update != nil {
				newMi, cmd := mi.Update(mi, msg)
				m.Items[idx] = newMi
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (mi *MenuItem) isActBlocked() bool {
	return mi.Act == nil
}

func (m *Menu) View() string {
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

func (m *Menu) Done() bool {
	for _, item := range m.Items {
		if !item.IsDone {
			return state.ToDo
		}
	}
	return state.Done
}
