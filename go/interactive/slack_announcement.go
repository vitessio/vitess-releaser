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

package interactive

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/slack"
	"github.com/vitessio/vitess-releaser/go/releaser/steps"
)

type (
	slackMessage          string
	slackAnnouncementType int
)

const (
	slackAnnouncementPostRelease = iota
	slackAnnouncementPreRequisite
)

func slackAnnouncementMenuItem(ctx context.Context, announcementType slackAnnouncementType) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)

	var name string
	var act func(*ui.MenuItem) (*ui.MenuItem, tea.Cmd)
	var isDone bool
	switch announcementType {
	case slackAnnouncementPreRequisite:
		act = slackAnnouncementPreRequisiteAct
		name = steps.SlackAnnouncement
		isDone = state.Issue.SlackPreRequisite
	case slackAnnouncementPostRelease:
		act = slackAnnouncementPostReleaseAct
		name = steps.SlackAnnouncementPost
		isDone = state.Issue.SlackPostRelease
	}

	return &ui.MenuItem{
		State:  state,
		Name:   name,
		Act:    act,
		Update: slackAnnouncementUpdate,
		IsDone: isDone,
	}
}

func slackAnnouncementPreRequisiteAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return slackMessage(slack.AnnouncementMessage(mi.State))
	}
}

func slackAnnouncementPostReleaseAct(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return slackMessage(slack.PostReleaseMessage(mi.State))
	}
}

func slackAnnouncementUpdate(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
	switch msg := msg.(type) {
	case slackMessage:
		return mi, ui.PushDialog(&ui.DoneDialog{
			StepName: mi.Name,
			Title:    "The following message must be posted on the #general and #releases OSS Slack channels",
			Message:  []string{string(msg)},
			IsDone:   mi.IsDone,
		})
	case ui.DoneDialogAction:
		if string(msg) != mi.Name {
			return mi, nil
		}
		if mi.Name == steps.SlackAnnouncement {
			mi.State.Issue.SlackPreRequisite = !mi.State.Issue.SlackPreRequisite
		} else if mi.Name == steps.SlackAnnouncementPost {
			mi.State.Issue.SlackPostRelease = !mi.State.Issue.SlackPostRelease
		}
		mi.IsDone = !mi.IsDone
		pl, fn := mi.State.UploadIssue()
		return mi, tea.Batch(func() tea.Msg {
			fn()
			return tea.Msg("")
		}, ui.PushDialog(ui.NewProgressDialog("Updating the Release Issue", pl)))
	}
	return mi, nil
}
