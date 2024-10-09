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

func VtopCreateBranchMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopCreateBranchAct
	if state.Issue.VtopCreateBranch {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopCreateBranch,
		Act:    act,
		Update: vtopCreateBranchUpdate,
		IsDone: state.Issue.VtopCreateBranch,

		Ignore: state.VtOpRelease.Release == "" || state.Issue.RC != 1,
	}
}

type vtopCreateBranchUrl string

func vtopCreateBranchUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(vtopCreateBranchUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.VtopCreateBranch
	return mi, nil
}

func vtopCreateBranchAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, freeze := code_freeze.VtopCreateBranch(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return vtopCreateBranchUrl(freeze())
	}, ui.PushDialog(ui.NewProgressDialog(steps.VtopCreateBranch, pl)))
}
