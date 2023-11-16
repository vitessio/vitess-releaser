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

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type mainMenu struct {
	list     list.Model
	quitting bool
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(mainMenuItem)
			if !ok {
				return m, nil
			}
			return i.act(m), nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m mainMenu) View() string {
	if m.quitting {
		return quitTextStyle.Render("Goodbye.")
	}
	return "\n" + m.list.View()
}

func MainScreen() {
	items := []list.Item{
		newItem("Prerequisite", []list.Item{
			newCheckListItem("Create Release Issue"),
			newCheckListItem("Announce the release on Slack"),
			newCheckListItem("Ensure all Pull Requests have been merged"),
		}),
		newItem("Pre-Release", nil),
		newItem("Release", nil),
		newItem("Post-Release", nil),
	}

	const defaultWidth = 40

	l := list.New(items, mainMenuItemDelegate{}, defaultWidth, listHeight)
	l.Title = "Vitess releaser, press enter choose the release step."
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := mainMenu{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
