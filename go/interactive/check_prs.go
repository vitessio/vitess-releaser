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
	"github.com/charmbracelet/lipgloss/table"
)

type closePRs struct {
	parent        tea.Model
	height, width int
	prs           []string
}

var _ tea.Model = (*closePRs)(nil)

func (c *closePRs) Init() tea.Cmd {
	return nil
}

func (c *closePRs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.height = msg.Height
		c.width = msg.Width
		return c, nil

	case tea.KeyMsg:
		return c.parent, nil
	}

	return nil, nil
}

func (c *closePRs) View() string {
	var rows [][]string
	for _, s := range c.prs {
		rows = append(rows, []string{s})
	}

	return "\n\n" +
		"These PRs still need to be closed before we can continue" +
		"\n\n" +
		table.New().Rows(rows...).Render() +
		"\n\n" +
		"Press any key to continue"
}
