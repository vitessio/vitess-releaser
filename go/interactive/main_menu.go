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
	"vitess.io/vitess-releaser/go/interactive/pre_release"
	"vitess.io/vitess-releaser/go/interactive/prerequisites"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
)

func blankLineMenu() *ui.MenuItem {
	return &ui.MenuItem{}
}

func MainScreen(ctx context.Context) {
	prereq := ui.NewMenu(
		ctx,
		"Prerequisites",
		slackAnnouncementMenuItem(ctx, slackAnnouncementPreRequisite),
		prerequisites.CheckSummaryMenuItem(ctx),
	)

	prerelease := ui.NewMenu(
		ctx,
		"Pre Release",
		pre_release.CodeFreezeMenuItem(ctx),
		pre_release.CreateReleasePRMenuItem(ctx),
		pre_release.CreateMilestoneMenuItem(ctx),
	)

	release := ui.NewMenu(
		ctx,
		"Release",
		nil,
	)

	postRelease := ui.NewMenu(
		ctx,
		"Post Release",
		slackAnnouncementMenuItem(ctx, slackAnnouncementPostRelease),
	)

	m := ui.NewMenu(ctx, "Main Menu",
		createIssueMenuItem(ctx),
		checkAndAddMenuItem(ctx),
		blankLineMenu(),
		&ui.MenuItem{
			IsDone:   prereq.Done(),
			SubItems: prereq.Items,
			Name:     "Prerequisites",
			Act:      subMenu(prereq),
		},
		&ui.MenuItem{
			IsDone:   prerelease.Done(),
			SubItems: prerelease.Items,
			Name:     "Pre Release",
			Act:      subMenu(prerelease),
		},
		&ui.MenuItem{
			IsDone:   release.Done(),
			SubItems: release.Items,
			Name:     "Release",
			Act:      subMenu(release),
		},
		&ui.MenuItem{
			IsDone:   postRelease.Done(),
			SubItems: postRelease.Items,
			Name:     "Post Release",
			Act:      subMenu(postRelease),
		},
	)

	m.MoveCursorToNextElem()

	if _, err := tea.NewProgram(ui.UI{State: releaser.UnwrapState(ctx), Active: m}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func subMenu(sub *ui.Menu) func(*ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return func(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) { return mi, ui.PushDialog(sub) }
}
