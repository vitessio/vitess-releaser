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

package release

import (
	"fmt"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CloseMilestone(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	return pl, func() string {
		milestone := fmt.Sprintf("v%s", state.VitessRelease.Release)
		nextRelease := releaser.FindVersionAfterNextRelease(state)
		nextMilestone := fmt.Sprintf("v%s", nextRelease)

		pl.NewStepf("Get opened Pull Requests for Milestone %s", milestone)
		prs := github.GetOpenedPRsByMilestone(state.VitessRelease.Repo, milestone)

		if len(prs) > 0 {
			pl.NewStepf("Move %d Pull Requests to the %s Milestone", len(prs), nextMilestone)
			github.AssignMilestoneToPRs(state.VitessRelease.Repo, nextMilestone, prs)
		} else {
			pl.TotalSteps--
		}

		pl.NewStepf("Close Milestone %s", milestone)
		url := github.CloseMilestone(state.VitessRelease.Repo, milestone)

		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		state.Issue.CloseMilestone.Done = true
		state.Issue.CloseMilestone.URL = url
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return url
	}
}
