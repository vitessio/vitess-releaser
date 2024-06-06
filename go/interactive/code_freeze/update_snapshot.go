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

package code_freeze

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/code_freeze"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func UpdateSnapshotOnMainMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := updateSnapshotOnMainAct
	if state.Issue.UpdateSnapshotOnMain.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.UpdateSnapshotOnMain,
		Act:    act,
		Update: updateSnapshotOnMainUpdate,
		Info:   state.Issue.UpdateSnapshotOnMain.URL,
		IsDone: state.Issue.UpdateSnapshotOnMain.Done,

		// We only want to update the SNAPSHOT version on main if we are doing a first RC release.
		// For higher RC releases we can assume it was already done during the first release.
		Ignore: state.Issue.RC != 1,
	}
}

type updateSnapshotOnMainUrl string

func updateSnapshotOnMainUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(updateSnapshotOnMainUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.UpdateSnapshotOnMain.URL
	mi.IsDone = mi.State.Issue.UpdateSnapshotOnMain.Done
	return mi, nil
}

func updateSnapshotOnMainAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, update := code_freeze.UpdateSnapshotOnMain(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return updateSnapshotOnMainUrl(update())
	}, ui.PushDialog(ui.NewProgressDialog("Update SNAPSHOT on main", pl)))
}
