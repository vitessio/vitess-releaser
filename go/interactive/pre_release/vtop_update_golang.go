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

func VtopUpdateGolangMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopUpdateGolangAct
	if state.Issue.VtopUpdateGolang.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopUpdateGolang,
		Act:    act,
		Update: vtopUpdateGolangUpdate,
		IsDone: state.Issue.VtopUpdateGolang.Done,
		Info:   state.Issue.VtopUpdateGolang.URL,

		Ignore: state.VtOpRelease.Release == "",
	}
}

type vtopUpdateGolangUrl string

func vtopUpdateGolangUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(vtopUpdateGolangUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.VtopUpdateGolang.Done
	mi.Info = mi.State.Issue.VtopUpdateGolang.URL
	return mi, nil
}

func vtopUpdateGolangAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, freeze := pre_release.VtopUpdateGolang(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return vtopUpdateGolangUrl(freeze())
	}, ui.PushDialog(ui.NewProgressDialog(steps.VtopUpdateGolang, pl)))
}
