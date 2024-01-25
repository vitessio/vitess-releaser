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
	"vitess.io/vitess-releaser/go/releaser/pre_release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func CreateReleasePRMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := createReleasePRAct
	if state.Issue.CreateReleasePR.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.CreateReleasePR,
		Act:    act,
		Update: createReleasePRUpdate,
		Info:   state.Issue.CreateReleasePR.URL,
		IsDone: state.Issue.CreateReleasePR.Done,
	}
}

type createReleasePRUrl string

func createReleasePRUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(createReleasePRUrl)
	if !ok {
		return mi, nil
	}

	mi.Info = mi.State.Issue.CreateReleasePR.URL
	mi.IsDone = mi.State.Issue.CreateReleasePR.Done
	return mi, nil
}

func createReleasePRAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, fn := pre_release.CreateReleasePR(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return createReleasePRUrl(fn())
	}, ui.PushDialog(ui.NewProgressDialog("Create the Release Pull Request", pl)))
}
