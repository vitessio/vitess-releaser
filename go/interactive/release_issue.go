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
	"vitess.io/vitess-releaser/go/releaser/issue"

	"vitess.io/vitess-releaser/go/releaser/github"
)

func createIssueMenuItem(ctx *releaser.Context) menuItem {
	return menuItem{
		ctx:    ctx,
		name:   "Create Release Issue",
		state:  "Loading...",
		act:    createIssue,
		init:   issueInit,
		update: issueUpdate,
	}
}

type releaseIssue string

func issueInit(ctx *releaser.Context) tea.Cmd {
	return func() tea.Msg {
		url := github.GetReleaseIssue(ctx)
		return releaseIssue(url)
	}
}

func createIssue(mi menuItem) (menuItem, tea.Cmd) {
	mi.state = "Creating issue..."
	pl, createIssueFn := issue.CreateReleaseIssue(mi.ctx)
	issueCreator := func() tea.Msg { return releaseIssue(createIssueFn()) }
	return mi, tea.Batch(
		issueCreator,
		pushDialog(newProgressDialog("Create Release Issue", pl)),
	)
}

func issueUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	url, ok := msg.(releaseIssue)
	if !ok {
		return mi, nil
	}
	if len(url) != 0 {
		return gotIssueURL(mi, string(url)), nil
	}

	mi.state = "TODO"
	return mi, nil
}

func gotIssueURL(item menuItem, url string) menuItem {
	item.name = "Release Issue"
	item.state = url
	item.act = nil // We don't want to accidentally create a second one
	return item
}
