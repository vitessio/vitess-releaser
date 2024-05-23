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

package interactive

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/post_release"
	"vitess.io/vitess-releaser/go/interactive/pre_release"
	"vitess.io/vitess-releaser/go/interactive/release"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
)

func blankLineMenu() *ui.MenuItem {
	return &ui.MenuItem{}
}

func MainScreen(ctx context.Context) {
	prereqMenu := ui.NewMenu(
		ctx,
		"Prerequisites",
		generalPrerequisiteMenuItem(ctx),
		slackAnnouncementMenuItem(ctx, slackAnnouncementPreRequisite),
		checkSummaryMenuItem(ctx),
		draftBlogPostMenuItem(ctx),
		requestCrossPostBlogPostMenuItem(ctx),
	)

	preReleaseMenu := ui.NewMenu(
		ctx,
		"Pre Release",
		pre_release.CodeFreezeMenuItem(ctx),
		pre_release.CopyBranchProtectionMenuItem(ctx),
		pre_release.CreateBackportToLabelMenuItem(ctx),
		pre_release.UpdateSnapshotOnMainMenuItem(ctx),
		pre_release.CreateReleasePRMenuItem(ctx),
		pre_release.CreateMilestoneMenuItem(ctx),
		pre_release.VtopCreateBranchMenuItem(ctx),
		pre_release.VtopBumpMainVersionMenuItem(ctx),
		pre_release.VtopUpdateGolangMenuItem(ctx),
		vtopUpdateCompatibilityTableMenuItem(ctx),
		createBlogPostPRMenuItem(ctx),
	)

	releaseMenu := ui.NewMenu(
		ctx,
		"Release",
		release.MergeReleasePRItem(ctx),
		release.TagReleaseItem(ctx),
		release.JavaReleaseItem(ctx),
		release.VtopCreateReleasePRMenuItem(ctx),
		release.VtopManualUpdateItem(ctx),
		release.ReleaseNotesOnMainItem(ctx),
		release.ReleaseNotesOnReleaseBranchItem(ctx),
		release.BackToDevModeItem(ctx),
		mergeBlogPostPRMenuItem(ctx),
		websiteDocumentationItem(ctx),
		benchmarkedItem(ctx),
		dockerImagesItem(ctx),
		release.CloseMilestoneItem(ctx),
	)
	releaseMenu.Sequential = true

	postReleaseMenu := ui.NewMenu(
		ctx,
		"Post Release",
		slackAnnouncementMenuItem(ctx, slackAnnouncementPostRelease),
		twitterMenuItem(ctx),
		post_release.CloseIssueItem(ctx),
	)

	m := ui.NewMenu(ctx, "Main Menu",
		createIssueMenuItem(ctx),
		checkAndAddMenuItem(ctx),
		blankLineMenu(),
		&ui.MenuItem{
			IsDone:   prereqMenu.Done(),
			SubItems: prereqMenu.Items,
			Name:     "Prerequisites",
			Act:      subMenu(prereqMenu),
		},
		&ui.MenuItem{
			IsDone:   preReleaseMenu.Done(),
			SubItems: preReleaseMenu.Items,
			Name:     "Pre Release",
			Act:      subMenu(preReleaseMenu),
		},
		&ui.MenuItem{
			IsDone:   releaseMenu.Done(),
			SubItems: releaseMenu.Items,
			Name:     "Release",
			Act:      subMenu(releaseMenu),
		},
		&ui.MenuItem{
			IsDone:   postReleaseMenu.Done(),
			SubItems: postReleaseMenu.Items,
			Name:     "Post Release",
			Act:      subMenu(postReleaseMenu),
		},
	)

	m.MoveCursorToNextElem()

	if _, err := tea.NewProgram(ui.UI{State: releaser.UnwrapState(ctx), Active: m}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func subMenu(sub *ui.Menu) func(*ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return func(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
		sub.MoveCursorToNextElem()
		return mi, ui.PushDialog(sub)
	}
}
