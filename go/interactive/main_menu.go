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
	"github.com/vitessio/vitess-releaser/go/interactive/code_freeze"
	"github.com/vitessio/vitess-releaser/go/interactive/post_release"
	"github.com/vitessio/vitess-releaser/go/interactive/pre_release"
	"github.com/vitessio/vitess-releaser/go/interactive/release"
	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
)

func blankLineMenu() *ui.MenuItem {
	return &ui.MenuItem{}
}

func MainScreen(ctx context.Context, state *releaser.State) {

	prereqMenu := ui.NewMenu(
		ctx,
		"Prerequisites",
		generalPrerequisiteMenuItem(ctx),
		slackAnnouncementMenuItem(ctx, slackAnnouncementPreRequisite),
		checkSummaryMenuItem(ctx),
		draftBlogPostMenuItem(ctx),
		requestCrossPostBlogPostMenuItem(ctx),
	)

	codeFreezeMenu := ui.NewMenu(
		ctx,
		"Code Freeze",
		code_freeze.CodeFreezeMenuItem(ctx),
		code_freeze.CopyBranchProtectionMenuItem(ctx),
		code_freeze.CreateNewLabelsMenuItem(ctx),
		code_freeze.UpdateSnapshotOnMainMenuItem(ctx),
		code_freeze.CreateMilestoneMenuItem(ctx),
		code_freeze.VtopCreateBranchMenuItem(ctx),
		code_freeze.VtopBumpMainVersionMenuItem(ctx),
		vtopUpdateCompatibilityTableMenuItem(ctx),
	)

	preReleaseMenu := ui.NewMenu(
		ctx,
		"Pre Release",
		pre_release.CreateReleasePRMenuItem(ctx),
		pre_release.VtopUpdateGolangMenuItem(ctx),
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

	menuTitle := fmt.Sprintf("Main Menu (%s)", github.CurrentUser())
	m := ui.NewMenu(ctx, menuTitle,
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
			IsDone:   codeFreezeMenu.Done(),
			SubItems: codeFreezeMenu.Items,
			Name:     "Code Freeze",
			Act:      subMenu(codeFreezeMenu),
			Ignore:   state.Issue.RC > 1 || state.Issue.GA,
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
