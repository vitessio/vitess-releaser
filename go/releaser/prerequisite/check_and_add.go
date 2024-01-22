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

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CheckAndAddPRsIssues(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	return pl, func() string {
		pl.NewStepf("Read Release Issue")
		state.LoadIssue()

		pl.NewStepf("Check and add Pull Requests")
		prsOnGH := github.CheckBackportToPRs(state.VitessRelease.Repo, state.MajorRelease)
		state.Issue.CheckBackport = addLinksToParentOfItems(state.Issue.CheckBackport, prsOnGH)

		pl.NewStepf("Check and add Release Blocker Issues")
		issuesOnGH := github.CheckReleaseBlockerIssues(state.VitessRelease.Repo, state.MajorRelease)
		state.Issue.ReleaseBlocker = addLinksToParentOfItems(state.Issue.ReleaseBlocker, issuesOnGH)

		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		fn()

		msg := GetCheckAndAddInfoMsg(state)
		if msg == "" {
			pl.NewStepf("Done")
		} else {
			pl.NewStepf(msg)
		}
		return msg
	}
}

func addLinksToParentOfItems(parent releaser.ParentOfItems, set map[string]any) releaser.ParentOfItems {
	for i, item := range parent.Items {
		if _, ok := set[item.URL]; !ok {
			parent.Items[i].Done = true
		}
	}

outerPR:
	for url := range set {
		for _, pri := range parent.Items {
			if pri.URL == url {
				continue outerPR
			}
		}
		parent.Items = append(parent.Items, releaser.ItemWithLink{
			URL: url,
		})
	}
	return parent
}

func GetCheckAndAddInfoMsg(state *releaser.State) string {
	nbPRs, nbIssues := state.Issue.CheckBackport.ItemsLeft(), state.Issue.ReleaseBlocker.ItemsLeft()

	msg := ""
	if nbPRs > 0 || nbIssues > 0 {
		msg = fmt.Sprintf("Found %d PRs and %d issues", nbPRs, nbIssues)
	}
	return msg
}
