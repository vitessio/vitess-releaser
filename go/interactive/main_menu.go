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
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
)

func MainScreen(ctx context.Context) {
	prereq := newMenu(
		ctx,
		"Prerequisites",
		slackAnnouncementMenuItem(ctx, slackAnnouncementPreRequisite),
		checkAndAddMenuItem(ctx),
		checkSummaryMenuItem(ctx),
	)

	prerelease := newMenu(
		ctx,
		"Pre Release",
		codeFreezeMenuItem(ctx),
		createReleasePRMenuItem(ctx),
		createMilestoneMenuItem(ctx),
	)

	postRelease := newMenu(
		ctx,
		"Post Release",
		slackAnnouncementMenuItem(ctx, slackAnnouncementPostRelease),
	)

	m := newMenu(ctx, "Main",
		createIssueMenuItem(ctx),
		&menuItem{
			isDone:                   state.ToDo,
			subItems:                 prereq.items,
			name:                     "Prerequisites",
			act:                      subMenu(prereq),
			blockActIfNoReleaseIssue: true,
		},
		&menuItem{
			isDone:                   state.ToDo,
			subItems:                 prerelease.items,
			name:                     "Pre Release",
			act:                      subMenu(prerelease),
			blockActIfNoReleaseIssue: true,
		},
		&menuItem{
			isDone:                   state.ToDo,
			subItems:                 nil,
			name:                     "Release",
			act:                      nil,
			blockActIfNoReleaseIssue: true,
		},
		&menuItem{
			isDone:                   state.ToDo,
			subItems:                 postRelease.items,
			name:                     "Post Release",
			act:                      subMenu(postRelease),
			blockActIfNoReleaseIssue: true,
		},
	)

	if _, err := tea.NewProgram(ui{state: releaser.UnwrapState(ctx), active: m}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func subMenu(sub *menu) func(*menuItem) (*menuItem, tea.Cmd) {
	return func(mi *menuItem) (*menuItem, tea.Cmd) { return mi, pushDialog(sub) }
}
