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

func ReleaseNotesOnMainItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := releaseNotesOnMainAct
	if state.Issue.ReleaseNotesOnMain.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.ReleaseNotesOnMain,
		Act:    act,
		Update: releaseNotesOnMainUpdate,
		Info:   state.Issue.ReleaseNotesOnMain.URL,
		IsDone: state.Issue.ReleaseNotesOnMain.Done,
	}
}

type releaseNotesOnMainUrl string

func releaseNotesOnMainUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(releaseNotesOnMainUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.ReleaseNotesOnMain.URL
	mi.IsDone = mi.State.Issue.ReleaseNotesOnMain.Done
	return mi, nil
}

func releaseNotesOnMainAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, fn := release.CopyReleaseNotesToBranch(mi.State, &mi.State.Issue.ReleaseNotesOnMain, "main")
	return mi, tea.Batch(func() tea.Msg {
		return releaseNotesOnMainUrl(fn())
	}, ui.PushDialog(ui.NewProgressDialog("Release Notes on Main", pl)))
}
