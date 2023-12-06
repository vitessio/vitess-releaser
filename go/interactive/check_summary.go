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
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type checkSummary []string

func checkSummaryMenuItem(ctx *releaser.Context) *menuItem {
	return &menuItem{
		ctx:    ctx,
		name:   steps.CheckSummary,
		isDone: ctx.Issue.CheckSummary,
		act:    checkSummaryAct,
		update: checkSummaryUpdate,
	}
}

func checkSummaryUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	switch msg := msg.(type) {
	case checkSummary:
		return mi, pushDialog(&doneDialog{
			stepName: mi.name,
			title:    "Check release note summary",
			message:  msg,
			isDone:   mi.isDone,
		})
	case doneDialogAction:
		if string(msg) != mi.name {
			return mi, nil
		}
		mi.ctx.Issue.CheckSummary = !mi.ctx.Issue.CheckSummary
		mi.isDone = !mi.isDone
		pl, fn := mi.ctx.UploadIssue()
		return mi, tea.Batch(func() tea.Msg {
			fn()
			return tea.Msg("")
		}, pushDialog(newProgressDialog("Updating the Release Issue", pl)))
	}
	return mi, nil
}

func checkSummaryAct(mi *menuItem) (*menuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return checkSummary(prerequisite.CheckSummary(mi.ctx))
	}
}
