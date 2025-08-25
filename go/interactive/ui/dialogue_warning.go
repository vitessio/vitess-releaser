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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type warningDialog struct {
	height, width int
	title         string
	message       []string
}

var _ tea.Model = warningDialog{}

func (c warningDialog) Init() tea.Cmd {
	return nil
}

func (c warningDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.height = msg.Height
		c.width = msg.Width

		return c, nil

	case tea.KeyMsg:
		return c, popDialog
	}

	return c, nil
}

func (c warningDialog) View() string {
	var rows [][]string
	for _, s := range c.message {
		rows = append(rows, []string{s})
	}

	lines := []string{c.title, ""}
	lines = append(lines, table.New().Data(table.NewStringData(rows...)).Width(c.width).Render())
	lines = append(lines, "", "Press any key to continue")

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}
