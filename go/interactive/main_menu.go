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
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/state"
)

type (
	// model is a screen with a current active window,
	// with the idea that new windows can come to the front,
	// but the old ones are still there behind
	model struct {
		active tea.Model
		stack  []tea.Model
	}
	_pop  struct{}
	_push struct {
		m tea.Model
	}
)

var pop tea.Cmd = func() tea.Msg { return _pop{} }

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case _pop:
		if len(m.stack) == 0 {
			return m, tea.Quit
		}
		lastIndex := len(m.stack) - 1
		m.active = m.stack[lastIndex]
		m.stack = m.stack[:lastIndex]
		return m, nil
	case _push:
		m.stack = append(m.stack, m.active)
		m.active = msg.m
		return m, nil
	}

	newActive, cmd := m.active.Update(msg)
	m.active = newActive
	return m, cmd
}

func (m model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.active.View(),
		"Vitess Releaser",
		fmt.Sprintf("Repo: %s Major Version: %s", state.VitessRepo, state.MajorRelease),
	)
}

func push(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return _push{m: m}
	}
}
func MainScreen() {
	prereq := newMenu(
		"Prerequisites",
		[]string{"Task", "Info"},
		[]menuItem{
			menuItem{name: "Create Release Issue", act: createIssue},
			menuItem{name: "Announce the release on Slack", act: nil},
			menuItem{name: "Ensure all Pull Requests have been merged", act: checkPRs},
		},
	)

	m := newMenu("Main", []string{"Task", "Info"}, []menuItem{
		{
			name:  "Prerequisites",
			state: "",
			act: func(mi menuItem) (menuItem, tea.Cmd) {
				return mi, push(prereq)
			},
		}, {
			name: "Pre Release",
			act:  nil,
		}, {
			name: "Release",
			act:  nil,
		},
	})

	if _, err := tea.NewProgram(model{active: m}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func createIssue(item menuItem) (menuItem, tea.Cmd) {
	prs := prerequisite.CheckPRs(state.MajorRelease)
	var cmd tea.Cmd
	if len(prs) == 0 {
		item.state = "[x]"
	} else {
		cmd = push(&closePRs{
			prs: prs,
		})
	}
	return item, cmd
}
