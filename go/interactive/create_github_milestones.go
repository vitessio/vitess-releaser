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
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type createMilestone string

func createMilestoneMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	act := createMilestoneAct
	if state.Issue.NewGitHubMilestone.Done {
		act = nil
	}
	return &ui.MenuItem{
		State:  state,
		Name:   steps.CreateMilestone,
		Act:    act,
		Update: createMilestoneUpdate,
		Info:   state.Issue.NewGitHubMilestone.URL,
		IsDone: state.Issue.NewGitHubMilestone.Done,
	}
}

func createMilestoneUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	milestoneLink, ok := msg.(createMilestone)
	if !ok || len(milestoneLink) == 0 {
		return mi, nil
	}

	mi.Info = mi.State.Issue.NewGitHubMilestone.URL
	mi.IsDone = mi.State.Issue.NewGitHubMilestone.Done
	mi.Act = nil // We don't want to accidentally create a second one
	return mi, nil
}

func createMilestoneAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, create := pre_release.NewMilestone(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return createMilestone(create())
	}, ui.PushDialog(ui.NewProgressDialog("Creating new GitHub Milestone", pl)))
}
