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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"vitess.io/vitess-releaser/go/interactive/state"
)

type DoneDialogAction string

type DoneDialog struct {
	height, width int
	Title         string
	Message       []string
	IsDone        bool
	StepName      string
}

var _ tea.Model = &DoneDialog{}

func (c *DoneDialog) Init() tea.Cmd {
	return nil
}

func (c *DoneDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.height = msg.Height
		c.width = msg.Width
		return c, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return c, popDialog
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				return c, popDialog
			case "x":
				return c, func() tea.Msg {
					c.IsDone = !c.IsDone
					return DoneDialogAction(c.StepName)
				}
			}
		}
	}

	return c, nil
}

func (c *DoneDialog) View() string {
	var rows [][]string
	for _, s := range c.Message {
		rows = append(rows, []string{s})
	}

	style := lipgloss.NewStyle().Bold(true)
	style = statusColor(style, c.IsDone)

	lines := []string{
		c.Title,
		"",
		fmt.Sprintf("Task status is: %s", style.Render(state.Fmt(c.IsDone))),
	}
	lines = append(lines, table.New().Data(table.NewStringData(rows...)).Width(c.width).Render())
	lines = append(
		lines,
		"",
		"Press 'x' to toggle the item between Done and To do.",
		"Press 'q' or 'enter' to quit.",
	)

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}
