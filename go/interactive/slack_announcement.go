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
	"vitess.io/vitess-releaser/go/releaser/slack"
)

type (
	slackMessage          string
	slackAnnouncementType int
)

const (
	slackAnnouncementPostRelease = iota
	slackAnnouncementPreRequisite
)

func slackAnnouncementMenuItem(announcementType slackAnnouncementType) menuItem {
	var act func(menuItem) (menuItem, tea.Cmd)
	switch announcementType {
	case slackAnnouncementPostRelease:
		act = slackAnnouncementPostReleaseAct
	case slackAnnouncementPreRequisite:
		act = slackAnnouncementPreRequisiteAct
	}

	return menuItem{
		name:   "Announce the release on Slack",
		act:    act,
		update: slackAnnouncementUpdate,
	}
}

func slackAnnouncementPreRequisiteAct(mi menuItem) (menuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return slackMessage(slack.AnnouncementMessage())
	}
}

func slackAnnouncementPostReleaseAct(mi menuItem) (menuItem, tea.Cmd) {
	return mi, func() tea.Msg {
		return slackMessage(slack.PostReleaseMessage())
	}
}

func slackAnnouncementUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	slackMsg, ok := msg.(slackMessage)
	if !ok {
		return mi, nil
	}

	mi.state = "Done"

	return mi, pushDialog(warningDialog{
		title:   "The following message must be posted on the #general and #releases OSS Slack channels",
		message: []string{string(slackMsg)},
	})
}
