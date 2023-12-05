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
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"vitess.io/vitess-releaser/go/interactive/state"
)

type doneDialog2DoneAct string

type doneDialog2 struct {
	height, width int
	title         string
	message       []string
	isDone        bool
}

var _ tea.Model = &doneDialog2{}

func (c *doneDialog2) Init() tea.Cmd {
	return nil
}

func (c *doneDialog2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case "x":
				return c, func() tea.Msg {
					c.isDone = !c.isDone
					return doneDialog2DoneAct("")
				}
			}
		}
	}

	return c, nil
}

func (c *doneDialog2) View() string {
	var rows [][]string
	for _, s := range c.message {
		rows = append(rows, []string{s})
	}

	lines := []string{
		c.title,
		"",
		fmt.Sprintf("Task status is: %s", state.Fmt(c.isDone)),
	}
	lines = append(lines, table.New().Data(table.NewStringData(rows...)).Width(c.width).Render())
	lines = append(
		lines,
		"",
		"Press 'x' to mark the item as Done/To do.",
		"Press 'enter' to quit.",
	)

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}
