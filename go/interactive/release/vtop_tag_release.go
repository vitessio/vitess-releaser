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

func VtopTagReleaseMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopTagReleaseAct
	if state.Issue.VtopTagRelease.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopTagRelease,
		Act:    act,
		Update: vtopTagReleaseUpdate,
		IsDone: state.Issue.VtopTagRelease.Done,
		Info:   state.Issue.VtopTagRelease.URL,

		Ignore: state.VtOpRelease.Release == "",
	}
}

type vtopTagReleaseUrl string

func vtopTagReleaseUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(vtopTagReleaseUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.VtopTagRelease.Done
	mi.Info = mi.State.Issue.VtopTagRelease.URL
	return mi, nil
}

func vtopTagReleaseAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, act := release.VtopTagRelease(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return vtopTagReleaseUrl(act())
	}, ui.PushDialog(ui.NewProgressDialog(steps.VtopTagRelease, pl)))
}
