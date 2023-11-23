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
)

func MainScreen() {
	prereq := newMenu(
		"Prerequisites",
		slackAnnouncementMenuItem(slackAnnouncementPreRequisite),
		checkPRsMenuItem(),
	)

	prerelease := newMenu(
		"Pre Release",
		codeFreezeMenuItem(),
	)

	postRelease := newMenu(
		"Post Release",
		slackAnnouncementMenuItem(slackAnnouncementPostRelease),
	)

	m := newMenu("Main",
		createIssueMenuItem(),
		menuItem{
			name:  "Prerequisites",
			state: "",
			act:   subMenu(prereq)},
		menuItem{
			name: "Pre Release",
			act:  subMenu(prerelease)},
		menuItem{
			name: "Release",
			act:  nil,
		},
		menuItem{
			name: "Post Release",
			act:  subMenu(postRelease),
		},
	)

	if _, err := tea.NewProgram(ui{active: m}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func subMenu(sub menu) func(menuItem) (menuItem, tea.Cmd) {
	return func(mi menuItem) (menuItem, tea.Cmd) { return mi, push(sub) }
}
