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
	"vitess.io/vitess-releaser/go/interactive/ui"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/post_release"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type twitterMsg []string

func TwitterMenuItem(ctx context.Context) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)
	return &ui.MenuItem{
		State:  state,
		Name:   steps.Twitter,
		IsDone: state.Issue.Twitter,
		Act:    twitterAct,
		Update: twitterUpdate,
	}
}

func twitterUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	switch msg := msg.(type) {
	case twitterMsg:
		return mi, ui.PushDialog(&ui.DoneDialog{
			StepName: mi.Name,
			Title:    "Twitter Announcement",
			Message:  msg,
			IsDone:   mi.IsDone,
		})
	case ui.DoneDialogAction:
		if string(msg) != mi.Name {
			return mi, nil
		}
		mi.State.Issue.Twitter = !mi.State.Issue.Twitter
		mi.IsDone = !mi.IsDone
		pl, fn := mi.State.UploadIssue()
		return mi, tea.Batch(func() tea.Msg {
			fn()
			return tea.Msg("")
		}, ui.PushDialog(ui.NewProgressDialog("Twitter Announcement", pl)))
	}
	return mi, nil
}

func twitterAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return twitterMsg(post_release.TwitterAnnouncement())
	}
}
