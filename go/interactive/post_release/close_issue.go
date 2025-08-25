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

package post_release

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/steps"
)

func CloseIssueItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)

	return &ui.MenuItem{
		State:  state,
		Name:   steps.CloseIssue,
		Act:    closeIssueAct,
		Update: closeIssueUpdate,
		Info:   state.IssueLink,
		IsDone: state.Issue.CloseIssue,
	}
}

type closeIssueUrl string

func closeIssueUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	_, ok := msg.(closeIssueUrl)
	if !ok {
		return mi, nil
	}

	mi.IsDone = mi.State.Issue.CloseIssue

	return mi, nil
}

func closeIssueAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	pl, fn := releaser.CloseReleaseIssue(mi.State)

	return mi, tea.Batch(func() tea.Msg {
		return closeIssueUrl(fn())
	}, ui.PushDialog(ui.NewProgressDialog("Close Issue", pl)))
}
