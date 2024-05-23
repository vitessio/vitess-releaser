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

func CodeFreezeMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := codeFreezeAct
	if state.Issue.CodeFreeze.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.CodeFreeze,
		Act:    act,
		Update: codeFreezeUpdate,
		Info:   state.Issue.CodeFreeze.URL,
		IsDone: state.Issue.CodeFreeze.Done,

		// We only want to do code freeze if we are doing a patch release or RC-1.
		// See RFC https://github.com/vitessio/vitess/issues/15586 which document this process.
		Ignore: state.VitessRelease.GA || state.Issue.RC > 1,
	}
}

type codeFreezeUrl string

func codeFreezeUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(codeFreezeUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.CodeFreeze.URL
	mi.IsDone = mi.State.Issue.CodeFreeze.Done
	return mi, nil
}

func codeFreezeAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, freeze := pre_release.CodeFreeze(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return codeFreezeUrl(freeze())
	}, ui.PushDialog(ui.NewProgressDialog("Code freeze", pl)))
}
