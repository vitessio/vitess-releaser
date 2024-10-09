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

func BackToDevModeItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := backToDevModeAct
	if state.Issue.BackToDevMode.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.BackToDev,
		Act:    act,
		Update: backToDevModeUpdate,
		Info:   state.Issue.BackToDevMode.URL,
		IsDone: state.Issue.BackToDevMode.Done,
	}
}

type backToDevModeUrl string

func backToDevModeUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(backToDevModeUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.BackToDevMode.URL
	mi.IsDone = mi.State.Issue.BackToDevMode.Done
	return mi, nil
}

func backToDevModeAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, back := release.BackToDevModeOnBranch(mi.State, &mi.State.Issue.BackToDevMode, mi.State.VitessRelease.ReleaseBranch)
	return mi, tea.Batch(func() tea.Msg {
		return backToDevModeUrl(back())
	}, ui.PushDialog(ui.NewProgressDialog("Back To Dev Mode", pl)))
}
