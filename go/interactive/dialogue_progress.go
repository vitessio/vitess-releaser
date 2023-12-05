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
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

type progressDialog struct {
	height, width int
	title         string
	progress      []string
	pl            *logging.ProgressLogging
	progressBar   progress.Model
}

func newProgressDialog(title string, pl *logging.ProgressLogging) *progressDialog {
	return &progressDialog{
		title:       title,
		pl:          pl,
		progressBar: progress.New(progress.WithDefaultGradient()),
	}
}

type tickMsg time.Time

var _ tea.Model = progressDialog{}

func (c progressDialog) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (c progressDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.height = msg.Height
		c.width = msg.Width
		c.progressBar.Width = msg.Width
		return c, nil

	case tickMsg:
		cmd := c.progressBar.SetPercent(float64(c.pl.GetDone()) / float64(c.pl.GetTotal()))
		c.progress = c.pl.GetStepInProgress()
		return c, tea.Batch(tickCmd(), cmd)

	case tea.KeyMsg:
		if c.pl.GetDone() != c.pl.GetTotal() {
			return c, nil
		}
		return c, popDialog

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := c.progressBar.Update(msg)
		c.progressBar = progressModel.(progress.Model)
		return c, cmd
	}

	return c, nil
}

func (c progressDialog) View() string {
	var rows [][]string
	for _, s := range c.progress {
		rows = append(rows, []string{s})
	}

	lines := []string{c.title, c.progressBar.View(), ""}
	lines = append(lines, table.New().Data(table.NewStringData(rows...)).Width(c.width).Render())

	if c.pl.GetDone() == c.pl.GetTotal() {
		lines = append(lines, "", "Press any key to continue")
	}

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}
