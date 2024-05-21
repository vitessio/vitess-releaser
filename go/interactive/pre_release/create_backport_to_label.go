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

package pre_release

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func CreateBackportToLabelMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := createBackportToLabelAct
	if state.Issue.CreateBackportToLabel {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.CreateBackportToLabel,
		Act:    act,
		Update: createBackportToLabelUpdate,
		IsDone: state.Issue.CreateBackportToLabel,

		// We only need to run this step when we are creating a new branch, aka doing RC-1
		Ignore: state.Issue.RC != 1,
	}
}

type createBackportToLabelUrl string

func createBackportToLabelUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(createBackportToLabelUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.CreateBackportToLabel
	return mi, nil
}

func createBackportToLabelAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, create := pre_release.CreateBackportToLabel(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return copyBranchProtectionUrl(create())
	}, ui.PushDialog(ui.NewProgressDialog("Create Backport To label", pl)))
}
