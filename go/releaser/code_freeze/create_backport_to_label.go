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

const (
	backportToLabelName  = "Backport to: "
	backportToLabelColor = "D4C5F9"
	backportToLabelDesc  = "Needs to be backport to "
)

func CreateBackportToLabel(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 5,
	}

	return pl, func() string {
		// Create the label for the rc release branch i.e. "Backport to: release-20.0-rc"
		labelRcBranch := backportToLabelName + state.VitessRelease.ReleaseBranch
		pl.NewStepf("Creating '%s' label", labelRcBranch)
		github.CreateLabel(state.VitessRelease.Repo, labelRcBranch, backportToLabelColor, backportToLabelDesc+state.VitessRelease.ReleaseBranch)

		// Create the label for the base release branch i.e. "Backport to: release-20.0"
		labelBaseBranch := backportToLabelName + state.VitessRelease.BaseReleaseBranch
		pl.NewStepf("Creating '%s' label", labelBaseBranch)
		github.CreateLabel(state.VitessRelease.Repo, labelBaseBranch, backportToLabelColor, backportToLabelDesc+state.VitessRelease.BaseReleaseBranch)

		// Let's use the base branch for the link as that label will also match the label of the rc branch
		labelURL := fmt.Sprintf("https://github.com/%s/labels?q=Backport+to%%3A+%s", state.VitessRelease.Repo, state.VitessRelease.BaseReleaseBranch)
		pl.NewStepf("Label created, see: %s", labelURL)

		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		state.Issue.CreateBackportToLabel.Done = true
		state.Issue.CreateBackportToLabel.URL = labelURL
		_, fn := state.UploadIssue()
		issueLink := fn()
		pl.NewStepf("Issue updated, see: %s", issueLink)
		return ""
	}
}
