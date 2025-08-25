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

func BackToDevModeBaseBranchItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := backToDevModeBaseBranchAct

	if state.Issue.BackToDevModeBaseBranch.Done {
		act = nil
	}

	return &ui.MenuItem{
		State:  state,
		Name:   steps.BackToDevOnBaseBranch,
		Act:    act,
		Update: backToDevModeBaseBranchUpdate,
		Info:   state.Issue.BackToDevModeBaseBranch.URL,
		IsDone: state.Issue.BackToDevModeBaseBranch.Done,

		// We only want to do this during the GA release
		Ignore: !state.Issue.GA,
	}
}

type backToDevModeBaseBranchUrl string

func backToDevModeBaseBranchUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(backToDevModeBaseBranchUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.BackToDevModeBaseBranch.URL
	mi.IsDone = mi.State.Issue.BackToDevModeBaseBranch.Done

	return mi, nil
}

func backToDevModeBaseBranchAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, back := release.BackToDevModeOnBranch(mi.State, &mi.State.Issue.BackToDevModeBaseBranch, mi.State.VitessRelease.BaseReleaseBranch)

	return mi, tea.Batch(func() tea.Msg {
		return backToDevModeBaseBranchUrl(back())
	}, ui.PushDialog(ui.NewProgressDialog("Back To Dev Mode on Base Branch", pl)))
}
