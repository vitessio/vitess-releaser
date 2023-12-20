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
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type checkAndAdd string

func checkAndAddMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return &ui.MenuItem{
		State:               state,
		Name:                steps.CheckAndAdd,
		Update:              checkAndAddUpdate,
		IsDone:              state.IssueNbGH != 0 && state.Issue.CheckBackport.Done() && state.Issue.ReleaseBlocker.Done(),
		Info:                "Loading ...",
		Init:                initCheckAndAdd,
		Act:                 actCheckAndAdd,
		DontCountInProgress: true,
	}
}

func initCheckAndAdd(mi *ui.MenuItem) tea.Cmd {
	if mi.State.IssueLink == "" {
		return nil
	}
	_, add := prerequisite.CheckAndAddPRsIssues(mi.State)
	return func() tea.Msg {
		return checkAndAdd(add())
	}
}

func actCheckAndAdd(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	if mi.State.IssueLink == "" {
		return mi, nil
	}
	pl, add := prerequisite.CheckAndAddPRsIssues(mi.State)
	return mi, tea.Batch(func() tea.Msg {
		return checkAndAdd(add())
	}, ui.PushDialog(ui.NewProgressDialog("", pl)))
}

func checkAndAddUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	out, ok := msg.(checkAndAdd)
	if !ok {
		return mi, nil
	}

	outStr := string(out)
	mi.Info = outStr
	mi.IsDone = mi.State.Issue.CheckBackport.Done() && mi.State.Issue.ReleaseBlocker.Done()
	return mi, nil
}
