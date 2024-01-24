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

package pre_release

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func VtopUpdateCompatibilityTableMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopUpdateCompatibilityTableAct
	if state.Issue.VtopUpdateCompatibilityTable {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopUpdateCompatibilityTable,
		Act:    act,
		Update: vtopUpdateCompatibilityTableUpdate,
		IsDone: state.Issue.VtopUpdateCompatibilityTable,

		Ignore: state.VtOpRelease.Release == "" || state.Issue.RC != 1,
	}
}

type vtopUpdateCompatibilityTableMsg []string

func vtopUpdateCompatibilityTableUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	switch msg := msg.(type) {
	case vtopUpdateCompatibilityTableMsg:
		return mi, ui.PushDialog(&ui.DoneDialog{
			StepName: mi.Name,
			Title:    steps.VtopUpdateCompatibilityTable,
			Message:  msg,
			IsDone:   mi.IsDone,
		})
	case ui.DoneDialogAction:
		if string(msg) != mi.Name {
			return mi, nil
		}
		mi.State.Issue.VtopUpdateCompatibilityTable = !mi.State.Issue.VtopUpdateCompatibilityTable
		mi.IsDone = !mi.IsDone
		pl, fn := mi.State.UploadIssue()
		return mi, tea.Batch(func() tea.Msg {
			fn()
			return tea.Msg("")
		}, ui.PushDialog(ui.NewProgressDialog(steps.VtopUpdateCompatibilityTable, pl)))
	}
	return mi, nil
}

func vtopUpdateCompatibilityTableAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return vtopUpdateCompatibilityTableMsg(pre_release.VtopUpdateCompatibilityTable(mi.State))
	}
}
