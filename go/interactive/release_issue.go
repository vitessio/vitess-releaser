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
	s := releaser.UnwrapState(ctx)
	i := &menuItem{
		state:  s,
		name:   steps.CreateReleaseIssue,
		isDone: state.ToDo,
		init:   createIssue,
		update: issueUpdate,
	}
	if s.IssueLink != "" {
		gotIssueURL(i)
	}
	return i
}

type releaseIssue struct {
	url string
	nb  int
}

func createIssue(mi *menuItem) tea.Cmd {
	_, createIssueFn := releaser.CreateReleaseIssue(mi.state)
	nb, url := createIssueFn()
	return func() tea.Msg {
		return releaseIssue{
			url: url,
			nb:  nb,
		}
	}
}

func issueUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	ri, ok := msg.(releaseIssue)
	if !ok {
		return mi, nil
	}
	if len(ri.url) != 0 && ri.nb != 0 {
		return gotIssueURL(mi), nil
	}
	return mi, nil
}

func gotIssueURL(item *menuItem) *menuItem {
	item.name = steps.ReleaseIssue
	item.info = item.state.IssueLink
	item.isDone = state.Done
	item.act = nil  // We don't want to accidentally create a second one
	item.init = nil // So we cancel the init function if we already have the issue
	return item
}
