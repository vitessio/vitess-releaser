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
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/slack"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

type (
	slackMessage          string
	slackAnnouncementType int
)

const (
	slackAnnouncementPostRelease = iota
	slackAnnouncementPreRequisite
)

func slackAnnouncementMenuItem(ctx *releaser.Context, announcementType slackAnnouncementType) *menuItem {
	var name string
	var act func(*menuItem) (*menuItem, tea.Cmd)
	var isDone bool
	switch announcementType {
	case slackAnnouncementPreRequisite:
		act = slackAnnouncementPreRequisiteAct
		name = steps.SlackAnnouncement
		isDone = ctx.Issue.SlackPreRequisite
	case slackAnnouncementPostRelease:
		act = slackAnnouncementPostReleaseAct
		name = steps.SlackAnnouncementPost
		isDone = ctx.Issue.SlackPostRelease
	}

	return &menuItem{
		ctx:    ctx,
		name:   name,
		act:    act,
		update: slackAnnouncementUpdate,
		isDone: isDone,
	}
}

func slackAnnouncementPreRequisiteAct(mi *menuItem) (*menuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return slackMessage(slack.AnnouncementMessage(mi.ctx))
	}
}

func slackAnnouncementPostReleaseAct(mi *menuItem) (*menuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return slackMessage(slack.PostReleaseMessage(mi.ctx))
	}
}

func slackAnnouncementUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	slackMsg, ok := msg.(slackMessage)
	if !ok {
		return mi, nil
	}

	return mi, pushDialog(doneDialog{
		title:   "The following message must be posted on the #general and #releases OSS Slack channels",
		message: []string{string(slackMsg)},
		isDone:  &mi.isDone,
		onDoneAsync: func() (*logging.ProgressLogging, func()) {
			return releaser.InverseStepStatus(mi.name)
		},
	})
}
