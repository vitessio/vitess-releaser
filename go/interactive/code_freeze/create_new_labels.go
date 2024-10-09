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
	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/code_freeze"
	"github.com/vitessio/vitess-releaser/go/releaser/steps"
)

func CreateNewLabelsMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := createNewLabelsAct
	if state.Issue.CreateNewLabels.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.CreateNewLabels,
		Act:    act,
		Update: createNewLabelsUpdate,
		IsDone: state.Issue.CreateNewLabels.Done,
		Info:   state.Issue.CreateNewLabels.URL,

		// We only need to run this step when we are creating a new branch, aka doing RC-1
		Ignore: state.Issue.RC != 1,
	}
}

type createNewLabelsUrl string

func createNewLabelsUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(createNewLabelsUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.CreateNewLabels.Done
	mi.Info = mi.State.Issue.CreateNewLabels.URL
	return mi, nil
}

func createNewLabelsAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, create := code_freeze.CreateNewLabels(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return createNewLabelsUrl(create())
	}, ui.PushDialog(ui.NewProgressDialog("Create New Labels", pl)))
}
