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

	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/release"
	"github.com/vitessio/vitess-releaser/go/releaser/steps"
)

func VtopManualUpdateItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)

	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopManualUpdate,
		Act:    vtopManualUpdateAct,
		Update: vtopManualUpdateUpdate,
		IsDone: state.Issue.VtopManualUpdate,
		Ignore: state.VtOpRelease.Release == "",
	}
}

type vtopManualUpdateMsg []string

func vtopManualUpdateUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	switch msg := msg.(type) {
	case vtopManualUpdateMsg:
		return mi, ui.PushDialog(&ui.DoneDialog{
			StepName: mi.Name,
			Title:    steps.VtopManualUpdate,
			Message:  msg,
			IsDone:   mi.IsDone,
		})
	case ui.DoneDialogAction:
		if string(msg) != mi.Name {
			return mi, nil
		}

		mi.State.Issue.VtopManualUpdate = !mi.State.Issue.VtopManualUpdate
		mi.IsDone = !mi.IsDone
		pl, fn := mi.State.UploadIssue()

		return mi, tea.Batch(func() tea.Msg {
			fn()
			return tea.Msg("")
		}, ui.PushDialog(ui.NewProgressDialog("Updating the Release Issue", pl)))
	}

	return mi, nil
}

func vtopManualUpdateAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return vtopManualUpdateMsg(release.VtopManualUpdateMessage(mi.State))
	}
}
