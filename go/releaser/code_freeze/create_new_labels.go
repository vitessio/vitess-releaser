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

const (
	releaseBlockerLabelName  = "Release Blocker: "
	releaseBlockerLabelColor = "B60205"
	releaseBlockerLabelDesc  = "This item blocks the release on branch "

	backportToLabelName  = "Backport to: "
	backportToLabelColor = "D4C5F9"
	backportToLabelDesc  = "Needs to be backport to "
)

func CreateNewLabels(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	return pl, func() string {
		// Create the label for the base release branch i.e. "Backport to: release-20.0"
		labelBaseBranch := backportToLabelName + state.VitessRelease.BaseReleaseBranch
		pl.NewStepf("Creating '%s' label", labelBaseBranch)
		github.CreateLabel(state.VitessRelease.Repo, labelBaseBranch, backportToLabelColor, backportToLabelDesc+state.VitessRelease.BaseReleaseBranch)

		releaseBlockerLabel := releaseBlockerLabelName + state.VitessRelease.BaseReleaseBranch
		pl.NewStepf("Creating '%s' label", releaseBlockerLabel)
		github.CreateLabel(state.VitessRelease.Repo, releaseBlockerLabel, releaseBlockerLabelColor, releaseBlockerLabelDesc+state.VitessRelease.BaseReleaseBranch)

		// Let's use the base branch for the link as that label will also match the label of the rc branch
		labelURL := fmt.Sprintf("https://github.com/%s/labels?q=%s", state.VitessRelease.Repo, state.VitessRelease.BaseReleaseBranch)
		pl.NewStepf("Label created, see: %s", labelURL)

		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		state.Issue.CreateNewLabels.Done = true
		state.Issue.CreateNewLabels.URL = labelURL
		_, fn := state.UploadIssue()
		issueLink := fn()
		pl.NewStepf("Issue updated, see: %s", issueLink)
		return ""
	}
}
