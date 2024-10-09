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

package prerequisite

import (
	"fmt"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
)

func CheckAndAddPRsIssues(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 6,
	}

	return pl, func() string {
		pl.NewStepf("Read Release Issue")
		state.LoadIssue()

		pl.NewStepf("Check and add Pull Requests")
		prsOnGH := github.CheckBackportToPRs(state.VitessRelease.Repo, state.VitessRelease.ReleaseBranch)
		state.Issue.CheckBackport = addLinksToParentOfItems(state.Issue.CheckBackport, prsOnGH)

		pl.NewStepf("Check and add Release Blocker Issues")
		releaseBlockerIssuesOnGH := github.CheckReleaseBlockerIssues(state.VitessRelease.Repo, state.VitessRelease.MajorRelease)
		pl.NewStepf("Check and add Release Blocker PRs")
		releaseBlockerPRsOnGH := github.CheckReleaseBlockerPRs(state.VitessRelease.Repo, state.VitessRelease.MajorRelease)

		// Merge the two maps together and add them to the issue
		releaseBlockers := map[string]any{}
		for key, _ := range releaseBlockerIssuesOnGH {
			releaseBlockers[key] = nil
		}
		for key, _ := range releaseBlockerPRsOnGH {
			releaseBlockers[key] = nil
		}
		state.Issue.ReleaseBlocker = addLinksToParentOfItems(state.Issue.ReleaseBlocker, releaseBlockers)

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
	nbPRs, releaseBlockerItems := state.Issue.CheckBackport.ItemsLeft(), state.Issue.ReleaseBlocker.ItemsLeft()

	msg := ""
	if nbPRs > 0 || releaseBlockerItems > 0 {
		msg = fmt.Sprintf("Found %d PRs and %d release blocker items", nbPRs, releaseBlockerItems)
	}
	return msg
}
