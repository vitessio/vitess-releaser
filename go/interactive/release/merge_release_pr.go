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

func MergeReleasePRItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := mergeReleasePRAct

	if state.Issue.MergeReleasePR.Done {
		act = nil
	}

	info := "Run this step once the Release Pull Request was created."
	if state.Issue.CreateReleasePR.URL != "" {
		info = state.Issue.CreateReleasePR.URL
	}

	return &ui.MenuItem{
		State:  state,
		Name:   steps.MergeReleasePR,
		Act:    act,
		Update: mergeReleasePRUpdate,
		Info:   info,
		IsDone: state.Issue.MergeReleasePR.Done,
	}
}

type mergeReleasePRUrl string

func mergeReleasePRUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	url, ok := msg.(mergeReleasePRUrl)
	if !ok {
		return mi, nil
	}

	if url != "" {
		mi.Info = string(url)
	}

	mi.IsDone = mi.State.Issue.MergeReleasePR.Done

	return mi, nil
}

func mergeReleasePRAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	// If the Release PR was not found in the Release Issue, do nothing
	if mi.State.Issue.CreateReleasePR.URL == "" {
		return mi, nil
	}

	pl, m := release.MergeReleasePR(mi.State)

	return mi, tea.Batch(func() tea.Msg {
		return mergeReleasePRUrl(m())
	}, ui.PushDialog(ui.NewProgressDialog("Merge Release Pull Request", pl)))
}
