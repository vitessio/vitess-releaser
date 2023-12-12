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
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type checkAndAdd string

func checkAndAddMenuItem(ctx context.Context) *menuItem {
	state := releaser.UnwrapState(ctx)
	return &menuItem{
		state:  state,
		name:   steps.CheckAndAdd,
		act:    checkAndAddAct,
		update: checkAndAddUpdate,
		isDone: state.IssueNbGH != 0 && state.Issue.CheckBackport.Done() && state.Issue.ReleaseBlocker.Done(),
		info:   prerequisite.GetCheckAndAddInfoMsg(state),
	}
}

func checkAndAddUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	out, ok := msg.(checkAndAdd)
	if !ok {
		return mi, nil
	}

	outStr := string(out)
	mi.info = outStr
	mi.isDone = mi.state.Issue.CheckBackport.Done() && mi.state.Issue.ReleaseBlocker.Done()
	return mi, nil
}

func checkAndAddAct(mi *menuItem) (*menuItem, tea.Cmd) {
	pl, add := prerequisite.CheckAndAddPRsIssues(mi.state)
	return mi, tea.Batch(func() tea.Msg {
		return checkAndAdd(add())
	}, pushDialog(newProgressDialog("Check and add pending PRs and release blocker Issues to Release Issue", pl)))
}
