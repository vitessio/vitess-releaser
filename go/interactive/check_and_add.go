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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
)

type checkAndAdd string

func checkAndAddMenuItem(ctx *releaser.Context) *menuItem {
	return &menuItem{
		ctx:    ctx,
		name:   "Check and add pending PRs and release blocker Issues to Release Issue",
		act:    checkAndAddAct,
		update: checkAndAddUpdate,
		status: state.ToDo, // TODO: read the initial state from the Release Issue on GitHub
	}
}

func checkAndAddUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	out, ok := msg.(checkAndAdd)
	if !ok {
		return mi, nil
	}

	outStr := string(out)
	mi.info = outStr
	if strings.Contains(outStr, "Found") {
		mi.status = state.ToDo
	} else {
		mi.status = state.Done
	}
	return mi, nil
}

func checkAndAddAct(mi *menuItem) (*menuItem, tea.Cmd) {
	mi.info = "running..."
	pl, add := prerequisite.CheckAndAddPRsIssues(mi.ctx)
	return mi, tea.Batch(func() tea.Msg {
		return checkAndAdd(add())
	}, pushDialog(newProgressDialog("Check and add pending PRs and release blocker Issues to Release Issue", pl)))
}
