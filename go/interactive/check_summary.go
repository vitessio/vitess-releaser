/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package interactive

import (
	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
)

type checkSummary []string

func checkSummaryMenuItem(ctx *releaser.Context) menuItem {
	return menuItem{
		ctx:    ctx,
		name:   "Check release note summary",
		act:    checkSummaryAct,
		update: checkSummaryUpdate,
	}
}

func checkSummaryUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	l, ok := msg.(checkSummary)
	if !ok {
		return mi, nil
	}

	// TODO: i think the goal here is to mark this as done only when the action is marked as done in the release issue
	// mi.state = fmt.Sprintf("Done.")

	return mi, pushDialog(infoDialog{
		title:   "Check release note summary",
		message: l,
	})
}

func checkSummaryAct(mi menuItem) (menuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return checkSummary(prerequisite.CheckSummary(mi.ctx))
	}
}
