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
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/releaser/state"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

type (
	slackMessage string
	slackAnnouncementType int
)

const (
	slackAnnouncementPostRelease = iota
	slackAnnouncementPreRequisite

	preRequisiteSlackMessage = `📣 The Vitess maintainers are planning on releasing v%s on <DATE>.`
	postReleaseSlackMessage = `📣 We have just released v%s. Check out the release notes on https://github.com/%s/release/tag/v%s`
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
	newRelease, _ := vitess.FindNextRelease(state.MajorRelease)
	return mi, func() tea.Msg {
		return slackMessage(fmt.Sprintf(preRequisiteSlackMessage, newRelease))
	}
}

func slackAnnouncementPostReleaseAct(mi menuItem) (menuItem, tea.Cmd) {
	newRelease, _ := vitess.FindNextRelease(state.MajorRelease)
	return mi, func() tea.Msg {
		return slackMessage(fmt.Sprintf(postReleaseSlackMessage, newRelease, state.VitessRepo, newRelease))
	}
}

func slackAnnouncementUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	slackMsg, ok := msg.(slackMessage)
	if !ok {
		return mi, nil
	}

	mi.state = "Done"

	return mi, push(warningDialog{
		title:   "The following message must be posted on the #general and #releases OSS Slack channels",
		message: []string{string(slackMsg)},
	})
}
