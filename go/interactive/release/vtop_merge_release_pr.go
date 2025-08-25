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

func VtopMergeReleasePRItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := vtopMergeReleasePRAct

	if state.Issue.VtopMergeReleasePR.Done {
		act = nil
	}

	info := "Run this step once the Release Pull Request was created."
	if state.Issue.VtopCreateReleasePR.URL != "" {
		info = state.Issue.VtopCreateReleasePR.URL
	}

	return &ui.MenuItem{
		State:  state,
		Name:   steps.VtopMergeReleasePR,
		Act:    act,
		Update: vtopMergeReleasePRUpdate,
		IsDone: state.Issue.VtopMergeReleasePR.Done,
		Info:   info,

		Ignore: state.VtOpRelease.Release == "",
	}
}

type vtopMergeReleasePRUrl string

func vtopMergeReleasePRUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(vtopMergeReleasePRUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.VtopMergeReleasePR.Done
	mi.Info = mi.State.Issue.VtopMergeReleasePR.URL

	return mi, nil
}

func vtopMergeReleasePRAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	if mi.State.Issue.VtopCreateReleasePR.URL == "" {
		return mi, nil
	}

	pl, act := release.VtopMergeReleasePR(mi.State)

	return mi, tea.Batch(func() tea.Msg {
		return vtopMergeReleasePRUrl(act())
	}, ui.PushDialog(ui.NewProgressDialog(steps.VtopMergeReleasePR, pl)))
}
