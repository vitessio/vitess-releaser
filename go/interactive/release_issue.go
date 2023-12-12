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
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func createIssueMenuItem(ctx context.Context) *menuItem {
	return &menuItem{
		state:  releaser.UnwrapState(ctx),
		name:   steps.CreateReleaseIssue,
		isDone: state.ToDo,
		act:    createIssue,
		init:   issueInit,
		update: issueUpdate,
	}
}

type releaseIssue struct {
	url string
	nb  int
}

func issueInit(mi *menuItem) tea.Cmd {
	return func() tea.Msg {
		return releaseIssue{
			url: mi.state.IssueLink,
			nb:  mi.state.IssueNbGH,
		}
	}
}

func createIssue(mi *menuItem) (*menuItem, tea.Cmd) {
	// safeguard
	if mi.state.IssueLink != "" {
		return mi, func() tea.Msg {
			return releaseIssue{
				url: mi.state.IssueLink,
				nb:  mi.state.IssueNbGH,
			}
		}
	}

	pl, createIssueFn := releaser.CreateReleaseIssue(mi.state)
	return mi, tea.Batch(
		func() tea.Msg {
			nb, url := createIssueFn()
			return releaseIssue{
				url: url,
				nb:  nb,
			}
		},
		pushDialog(newProgressDialog("Create Release Issue", pl)),
	)
}

func issueUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	ri, ok := msg.(releaseIssue)
	if !ok {
		return mi, nil
	}
	if len(ri.url) != 0 && ri.nb != 0 {
		return gotIssueURL(mi, ri), nil
	}
	return mi, nil
}

func gotIssueURL(item *menuItem, ri releaseIssue) *menuItem {
	item.name = steps.ReleaseIssue
	item.info = ri.url
	item.isDone = state.Done
	item.act = nil // We don't want to accidentally create a second one
	item.state.IssueNbGH = ri.nb
	item.state.IssueLink = ri.url
	return item
}
