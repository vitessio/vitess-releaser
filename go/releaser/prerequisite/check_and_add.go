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

package prerequisite

import (
	"fmt"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CheckAndAddPRsIssues(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 4,
	}

	return pl, func() string {
		pl.NewStepf("Check and add Pull Requests")
		prsOnGH := github.CheckBackportToPRs(state.VitessRepo, state.MajorRelease)
	outerPR:
		for _, pr := range prsOnGH {
			// separate the PR number from the URL
			nb := pr.URL[strings.LastIndex(pr.URL, "/")+1:]
			markdownURL := fmt.Sprintf("#%s", nb)
			for _, pri := range state.Issue.CheckBackport.Items {
				if pri.URL == markdownURL {
					continue outerPR
				}
			}
			state.Issue.CheckBackport.Items = append(state.Issue.CheckBackport.Items, releaser.ItemWithLink{
				URL: markdownURL,
			})
		}

		pl.NewStepf("Check and add Release Blocker Issues")
		issuesOnGH := github.CheckReleaseBlockerIssues(state.VitessRepo, state.MajorRelease)
	outerRBI:
		for _, i := range issuesOnGH {
			// separate the Issue number from the URL
			nb := i.URL[strings.LastIndex(i.URL, "/")+1:]
			markdownURL := fmt.Sprintf("#%s", nb)
			for _, rbi := range state.Issue.ReleaseBlocker.Items {
				if rbi.URL == markdownURL {
					continue outerRBI
				}
			}
			state.Issue.ReleaseBlocker.Items = append(state.Issue.ReleaseBlocker.Items, releaser.ItemWithLink{
				URL: markdownURL,
			})
		}

		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		fn()

		msg := GetCheckAndAddInfoMsg(state)
		pl.NewStepf(msg)
		return msg
	}
}

func GetCheckAndAddInfoMsg(state *releaser.State) string {
	nbPRs, nbIssues := state.Issue.CheckBackport.ItemsLeft(), state.Issue.ReleaseBlocker.ItemsLeft()

	msg := fmt.Sprintf("Up to date, see: %s", state.IssueLink)
	if nbPRs > 0 || nbIssues > 0 {
		msg = fmt.Sprintf("Found %d PRs and %d issues, see: %s", nbPRs, nbIssues, state.IssueLink)
	}
	return msg
}
