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

package release

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func ReleaseNotesOnReleaseBranchItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := releaseNotesOnReleaseBranchAct
	if state.Issue.ReleaseNotesOnReleaseBranch.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.ReleaseNotesOnReleaseBranch,
		Act:    act,
		Update: releaseNotesOnReleaseBranchUpdate,
		Info:   state.Issue.ReleaseNotesOnReleaseBranch.URL,
		IsDone: state.Issue.ReleaseNotesOnReleaseBranch.Done,

		// We want to ignore this step if we are doing a patch release.
		Ignore: !state.Issue.GA && state.Issue.RC == 0,
	}
}

type releaseNotesOnReleaseBranchUrl string

func releaseNotesOnReleaseBranchUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(releaseNotesOnReleaseBranchUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.ReleaseNotesOnReleaseBranch.URL
	mi.IsDone = mi.State.Issue.ReleaseNotesOnReleaseBranch.Done
	return mi, nil
}

func releaseNotesOnReleaseBranchAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, fn := release.CopyReleaseNotesToBranch(mi.State, &mi.State.Issue.ReleaseNotesOnReleaseBranch, mi.State.VitessRelease.BaseReleaseBranch)
	return mi, tea.Batch(func() tea.Msg {
		return releaseNotesOnMainUrl(fn())
	}, ui.PushDialog(ui.NewProgressDialog("Release Notes on Main", pl)))
}
