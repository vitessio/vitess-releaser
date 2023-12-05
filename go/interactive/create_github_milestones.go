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
	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/pre_release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type createMilestone string

func createMilestoneMenuItem(ctx *releaser.Context) *menuItem {
	return &menuItem{
		ctx:    ctx,
		name:   steps.CreateMilestone,
		act:    createMilestoneAct,
		update: createMilestoneUpdate,
		isDone: state.ToDo,
	}
}

func createMilestoneUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	milestoneLink, ok := msg.(createMilestone)
	if !ok {
		return mi, nil
	}

	mi.info = string(milestoneLink)
	mi.isDone = state.Done
	return mi, nil
}

func createMilestoneAct(mi *menuItem) (*menuItem, tea.Cmd) {
	mi.info = "running..."
	pl, create := pre_release.NewMilestone(mi.ctx)
	return mi, tea.Batch(func() tea.Msg {
		return createMilestone(create())
	}, pushDialog(newProgressDialog("Creating new GitHub Milestone", pl)))
}
