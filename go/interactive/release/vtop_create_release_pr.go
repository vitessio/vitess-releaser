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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func VtopCreateReleasePRMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopCreateReleasePRAct
	if state.Issue.VtopCreateReleasePR.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopCreateReleasePR,
		Act:    act,
		Update: vtopCreateReleasePRUpdate,
		IsDone: state.Issue.VtopCreateReleasePR.Done,
		Info:   strings.Join(state.Issue.VtopCreateReleasePR.URLs, " | "),

		Ignore: state.VtOpRelease.Release == "",
	}
}

type vtopCreateReleasePRUrl string

func vtopCreateReleasePRUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(vtopCreateReleasePRUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.VtopCreateReleasePR.Done
	mi.Info = strings.Join(mi.State.Issue.VtopCreateReleasePR.URLs, " | ")
	return mi, nil
}

func vtopCreateReleasePRAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, freeze := release.VtopCreateReleasePR(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return vtopCreateReleasePRUrl(freeze())
	}, ui.PushDialog(ui.NewProgressDialog(steps.VtopCreateReleasePR, pl)))
}
