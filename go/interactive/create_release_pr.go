/*
Copyright 2023 The Vitess Authors.

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

package interactive

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func createReleasePRMenuItem(ctx context.Context) *menuItem {
	state := releaser.UnwrapState(ctx)
	act := createMilestoneAct
	if state.Issue.CreateReleasePR.Done {
		act = nil
	}
	return &menuItem{
		state:  state,
		name:   steps.CreateReleasePR,
		act:    act,
		update: createReleasePRUpdate,
		info:   state.Issue.CreateReleasePR.URL,
		isDone: state.Issue.CreateReleasePR.Done,
	}
}

type createReleasePRUrl string

func createReleasePRUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	_, ok := msg.(createReleasePRUrl)
	if !ok {
		return mi, nil
	}

	mi.info = mi.state.Issue.CreateReleasePR.URL
	mi.isDone = mi.state.Issue.CreateReleasePR.Done
	return mi, nil
}

func createReleasePRAct(mi *menuItem) (*menuItem, tea.Cmd) {
	pl, fn := pre_release.CreateReleasePR(mi.state)
	return mi, tea.Batch(func() tea.Msg {
		return createReleasePRUrl(fn())
	}, pushDialog(newProgressDialog("Create the Release Pull Request", pl)))
}
