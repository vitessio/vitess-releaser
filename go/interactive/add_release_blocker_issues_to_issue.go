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
	"vitess.io/vitess-releaser/go/releaser/issue"
)

type addReleaseBlockerIssues string

func addReleaseBlockerIssuesMenuItem(ctx *releaser.Context) *menuItem {
	return &menuItem{
		ctx:    ctx,
		name:   "Release Blocker Issues: Add to Release Issue",
		act:    addReleaseBlockerIssuesAct,
		update: addReleaseBlockerIssuesUpdate,
	}
}

func addReleaseBlockerIssuesUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	releaseIssueLink, ok := msg.(addReleaseBlockerIssues)
	if !ok {
		return mi, nil
	}

	mi.state = string(releaseIssueLink)
	return mi, nil
}

func addReleaseBlockerIssuesAct(mi *menuItem) (*menuItem, tea.Cmd) {
	mi.state = "running..."
	pl, add := issue.AddReleaseBlockerIssues(mi.ctx)
	return mi, tea.Batch(func() tea.Msg {
		return addReleaseBlockerIssues(add())
	}, pushDialog(newProgressDialog("Adding opened Release Blocker Issues to Release Issue", pl)))
}
