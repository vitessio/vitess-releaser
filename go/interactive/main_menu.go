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
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/state"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	mainMenu struct {
		list    list.Model
		items   []*menuItem
		loading bool

		am            *actionManager
		height, width int
	}

	menuItem struct {
		name, state string
		act         func(msg tea.Msg) (tea.Model, tea.Cmd)
	}
)

func (m *menuItem) Init() tea.Cmd {
	return nil
}

func (m *menuItem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.act(msg)
}

func (m *menuItem) View() string {
	return m.name + " " + m.state
}

func (m mainMenu) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		func() tea.Msg {
			github.GetReleaseIssue(state.MajorRelease)
			return nil
		})
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
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
	if m.loading {
		return quitTextStyle.Render("Loading...")
	}

	var items []string
	for _, item := range m.items {
		items = append(items, item.View())
	}

	return "\n" +
		m.list.View() +
		"\n\n" +
		"Repo: " +
		state.VitessRepo
}

func MainScreen() {
	am := &actionManager{}
	m := mainMenu{
		am: am,
		items: []*menuItem{
			{
				name:  "apa",
				state: "oh noes",
				act:   nil,
			},
			{
				name:  "fes",
				state: "oh noes",
				act:   nil,
			},
		},
	}

	items := getMenuItems(m)

	const defaultWidth = 40

	l := list.New(items, mainMenuItemDelegate{}, defaultWidth, listHeight)
	l.Title = "Vitess releaser, press enter choose the release step."
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	m.list = l

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func getMenuItems(m mainMenu) []list.Item {
	i := []list.Item{
		newCheckListItem("Create Release Issue", func() (string, tea.Model) {
			link := prerequisite.CreateReleaseIssue(state.MajorRelease)
			return link, nil
		}),
		newCheckListItem("Announce the release on Slack", nil),
		newCheckListItem("Ensure all Pull Requests have been merged", func() (string, tea.Model) {
			prs := prerequisite.CheckPRs(state.MajorRelease)
			if len(prs) == 0 {
				return "[x]", nil
			}
			return "", &closePRs{
				parent: m,
				prs:    prs,
			}
		}),
	}
	items := []list.Item{
		newItem("Prerequisite", i),
		newItem("Pre-Release", nil),
		newItem("Release", nil),
		newItem("Post-Release", nil),
	}
	return items
}

func checkPRs(returnTo tea.Model) (string, tea.Model) {
	prs := prerequisite.CheckPRs(state.MajorRelease)
	if len(prs) == 0 {
		return "[x]", nil
	}
	return "", &closePRs{
		parent: returnTo,
		prs:    prs,
	}
}
