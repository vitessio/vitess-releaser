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

func VtopBumpMainVersionMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopBumpMainVersionAct
	if state.Issue.VtopBumpMainVersion.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopBumpMainVersion,
		Act:    act,
		Update: vtopBumpMainVersionUpdate,
		Info:   state.Issue.VtopBumpMainVersion.URL,
		IsDone: state.Issue.VtopBumpMainVersion.Done,

		Ignore: state.VtOpRelease.Release == "" || state.Issue.RC != 1,
	}
}

type vtopBumpMainVersionUrl string

func vtopBumpMainVersionUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(vtopBumpMainVersionUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.VtopBumpMainVersion.Done
	mi.Info = mi.State.Issue.VtopBumpMainVersion.URL
	return mi, nil
}

func vtopBumpMainVersionAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, fn := pre_release.VtopBumpMainVersion(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return vtopBumpMainVersionUrl(fn())
	}, ui.PushDialog(ui.NewProgressDialog(steps.VtopBumpMainVersion, pl)))
}
