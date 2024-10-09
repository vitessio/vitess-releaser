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

package code_freeze

import (
	"fmt"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
)

func NewMilestone(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	// Two extra steps if we are doing an RC-1 release
	if state.Issue.RC == 1 {
		pl.TotalSteps += 2
	}

	return pl, func() string {
		var link string
		defer func() {
			if link == "" {
				return
			}
			state.Issue.NewGitHubMilestone.Done = true
			state.Issue.NewGitHubMilestone.URL = link

			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Finding the next Milestone")
		nextNextRelease := releaser.FindVersionAfterNextRelease(state)
		newMilestone := fmt.Sprintf("v%s", nextNextRelease)

		ms := github.GetMilestonesByName(state.VitessRelease.Repo, newMilestone)
		if len(ms) > 0 {
			pl.SetTotalStep(4) // we do one lest step if the milestone already exist
			link = ms[0].URL
			pl.NewStepf("Found an existing Milestone: %s", link)
			return link
		}

		pl.NewStepf("Creating Milestone %s on GitHub", newMilestone)
		link = github.CreateNewMilestone(state.VitessRelease.Repo, newMilestone)
		pl.NewStepf("New Milestone %s created: %s", newMilestone, link)

		// Let's assume we release v20.0.0, the current milestone is v20.0.0, and new milestone which we
		// just created is v21.0.0:
		//
		// During the RC-1, we want to move all opened PRs from the v20.0.0 milestone to the v21.0.0 milestone.
		// All the PRs that are still opened on the v20.0.0 milestone are based on main, and thus need to be
		// assigned the new milestone for main: v21.0.0. We might decide to backport an opened PR to the release
		// branch, in which case GitHub Actions will assign the v20.0.0 milestone to the backport.
		//
		// This only applies to RC-1 releases. For patch releases, since the branch is frozen, the risk of
		// merging a PR in a release that don't match the milestone is very slim.
		if state.Issue.RC == 1 {
			currentMilestone := fmt.Sprintf("v%s", releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release))
			pl.NewStepf("Get opened Pull Requests for Milestone %s", currentMilestone)
			prs := github.GetOpenedPRsByMilestone(state.VitessRelease.Repo, currentMilestone)

			if len(prs) > 0 {
				pl.NewStepf("Move %d Pull Requests to the %s Milestone", len(prs), newMilestone)
				github.AssignMilestoneToPRs(state.VitessRelease.Repo, newMilestone, prs)
			} else {
				pl.NewStepf("No opened Pull Request found for Milestone %s, nothing to move", currentMilestone)
			}
		}
		return link
	}
}
