/*
Copyright 2024 The Vitess Authors.

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
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/vitessio/vitess-releaser/go/releaser"
)

type (
	// but the old ones are still there behind.
	UI struct {
		State  *releaser.State
		Active tea.Model
		Stack  []tea.Model
		Size   tea.WindowSizeMsg
	}
	_pop  struct{}
	_push struct {
		m tea.Model
	}
)

var popDialog tea.Cmd = func() tea.Msg { return _pop{} }

func PushDialog(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return _push{m: m}
	}
}

func (m UI) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, m := range m.Stack {
		cmds = append(cmds, m.Init())
	}

	cmds = append(cmds, m.Active.Init())
	cmds = append(cmds, tea.EnterAltScreen)

	return tea.Batch(cmds...)
}

func (m UI) newActive(d tea.Model) (UI, tea.Cmd) {
	m.Active = d
	initCmd := d.Init() // we call Init() every time a UI becomes Active

	var sizeCmd tea.Cmd
	m.Active, sizeCmd = m.Active.Update(m.Size)

	return m, tea.Batch(initCmd, sizeCmd)
}

func (m UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// these messages we take care of here
	switch msg := msg.(type) {
	case _pop:
		if len(m.Stack) == 0 {
			return m, tea.Quit
		}

		lastIndex := len(m.Stack) - 1
		popped := m.Stack[lastIndex]
		m.Stack = m.Stack[:lastIndex]

		return m.newActive(popped)
	case _push:
		m.Stack = append(m.Stack, m.Active)
		return m.newActive(msg.m)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		newActive, cmd := m.Active.Update(msg)
		m.Active = newActive

		return m, cmd

	case tea.WindowSizeMsg:
		m.Size = msg
	}

	// Other messages are passed on to all dialogs of the Stack
	var cmds []tea.Cmd

	newStack := make([]tea.Model, len(m.Stack))

	for i, m := range m.Stack {
		var cmd tea.Cmd
		newStack[i], cmd = m.Update(msg)
		cmds = append(cmds, cmd)
	}

	newActive, cmd := m.Active.Update(msg)
	cmds = append(cmds, cmd)
	m.Active = newActive

	return m, tea.Batch(cmds...)
}

func (m UI) View() string {
	if _, ok := m.Active.(ProgressDialog); ok {
		return m.Active.View()
	}

	elems := []string{
		m.Active.View(),
		"",
	}

	elems = append(elems, bgStyle.Render("Vitess Releaser: 'q' = back, 'enter' = action"))
	elems = append(elems, bgStyle.Render(fmt.Sprintf("Vitess repo: %s | Vitess release: v%s", m.State.VitessRelease.Repo, m.State.VitessRelease.Release)))

	if m.State.VtOpRelease.Release != "" {
		elems = append(elems, bgStyle.Render(fmt.Sprintf("Vtop repo: %s | Vtop release: v%s", m.State.VtOpRelease.Repo, releaser.AddRCToReleaseTitle(m.State.VtOpRelease.Release, m.State.Issue.RC))))
	}

	elems = append(elems, bgStyle.Render(fmt.Sprintf("Release Date: %s", m.State.Issue.Date.Format(time.DateOnly))))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		elems...,
	)
}
