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
	"vitess.io/vitess-releaser/go/releaser/logging"
)

type doneDialog struct {
	height, width int
	title         string
	message       []string
	isDone        *bool
	onDoneAsync   func() (*logging.ProgressLogging, func())
}

var _ tea.Model = doneDialog{}

func (c doneDialog) Init() tea.Cmd {
	return nil
}

func (c doneDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

				if c.isDone == nil {
					return c, nil
				}
				if *(c.isDone) {
					*(c.isDone) = state.ToDo
				} else {
					*(c.isDone) = state.Done
				}

				// call the callback
				if c.onDoneAsync != nil {
					pl, fn := c.onDoneAsync()
					return c, tea.Batch(func() tea.Msg {
						fn()
						return ""
					}, pushDialog(newProgressDialog("Loading ...", pl)))
				}
			}
		}
	}

	return c, nil
}

func (c doneDialog) View() string {
	var rows [][]string
	for _, s := range c.message {
		rows = append(rows, []string{s})
	}

	s := false
	if c.isDone != nil {
		s = *(c.isDone)
	}

	lines := []string{
		c.title,
		"",
		fmt.Sprintf("Task isDone is: %s", state.Fmt(s)),
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
