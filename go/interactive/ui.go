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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"vitess.io/vitess-releaser/go/releaser/state"
)

type (
	// ui is an ui with a current active window,
	// with the idea that new windows can come to the front,
	// but the old ones are still there behind
	ui struct {
		active tea.Model
		stack  []tea.Model
		width  int
		height int
	}
	_pop  struct{}
	_push struct {
		m tea.Model
	}
)

var popDialog tea.Cmd = func() tea.Msg { return _pop{} }

func pushDialog(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return _push{m: m}
	}
}

func (m ui) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, m := range m.stack {
		cmds = append(cmds, m.Init())
	}
	cmds = append(cmds, m.active.Init())
	cmds = append(cmds, tea.EnterAltScreen)
	return tea.Batch(cmds...)
}

func (m ui) newActive(d tea.Model) (ui, tea.Cmd) {
	m.active = d
	initCmd := d.Init() // we call Init() every time a ui becomes active
	var sizeCmd tea.Cmd
	m.active, sizeCmd = m.active.Update(tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.height,
	})

	return m, tea.Batch(initCmd, sizeCmd)
}

func (m ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case _pop:
		if len(m.stack) == 0 {
			return m, tea.Quit
		}
		lastIndex := len(m.stack) - 1
		popped := m.stack[lastIndex]
		m.stack = m.stack[:lastIndex]
		return m.newActive(popped)
	case _push:
		m.stack = append(m.stack, m.active)
		return m.newActive(msg.m)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, popDialog
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmds []tea.Cmd
	newStack := make([]tea.Model, len(m.stack))
	for i, m := range m.stack {
		var cmd tea.Cmd
		newStack[i], cmd = m.Update(msg)
		cmds = append(cmds, cmd)
	}
	newActive, cmd := m.active.Update(msg)
	cmds = append(cmds, cmd)
	m.active = newActive
	return m, tea.Batch(cmds...)
}

func (m ui) View() string {
	_, isMenu := m.active.(menu)
	if !isMenu {
		return m.active.View()
	}
	title := "Vitess Releaser"
	if m.width == 0 {
		m.width = 100
	}
	lft := bgStyle.Render(title)
	width := m.width - len(title)
	s := bgStyle.Copy().Width(width).Align(lipgloss.Right)
	rgt := fmt.Sprintf("Repo: %s Releasing Branch: %s", state.VitessRepo, state.MajorRelease)
	statusBar := lft + s.Render(rgt)
	return lipgloss.JoinVertical(
		lipgloss.Right,
		m.active.View(),
		statusBar,
	)
}
