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

func TagReleaseItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := tagReleaseAct
	if state.Issue.TagRelease.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.TagRelease,
		Act:    act,
		Update: tagReleaseUpdate,
		Info:   state.Issue.TagRelease.URL,
		IsDone: state.Issue.TagRelease.Done,
	}
}

type tagReleaseUrl string

func tagReleaseUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(tagReleaseUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.TagRelease.URL
	mi.IsDone = mi.State.Issue.TagRelease.Done
	return mi, nil
}

func tagReleaseAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, tag := release.TagRelease(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return tagReleaseUrl(tag())
	}, ui.PushDialog(ui.NewProgressDialog("Tag Release", pl)))
}
