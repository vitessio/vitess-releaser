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

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func NewMilestone(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
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
		return link
	}
}
