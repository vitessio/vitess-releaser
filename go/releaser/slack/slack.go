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

package slack

import (
	"fmt"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
)

const (
	preRequisiteSlackMessage = `📣 The Vitess maintainers are planning on releasing v%s on %s.`
	postReleaseSlackMessage  = `📣 We have just released v%s. Check out the release notes on https://github.com/%s/releases/tag/v%s`
)

func AnnouncementMessage(state *releaser.State) string {
	return fmt.Sprintf(preRequisiteSlackMessage, state.VitessRelease.Release, state.Issue.Date.Format("Mon _2 Jan"))
}

func PostReleaseMessage(state *releaser.State) string {
	return fmt.Sprintf(postReleaseSlackMessage, state.VitessRelease.Release, state.VitessRelease.Repo, strings.ToLower(state.VitessRelease.Release))
}
