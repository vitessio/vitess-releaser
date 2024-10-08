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
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CopyBranchProtectionRules(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 4,
	}

	return pl, func() string {
		defer func() {
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			state.Issue.CopyBranchProtectionRules = true
			_, fn := state.UploadIssue()
			issueLink := fn()
			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		if state.VitessRelease.Repo != "vitessio/vitess" {
			pl.TotalSteps--
			pl.NewStepf("Skipping as we are not running on vitessio/vitess.")
			return ""
		}
		pl.NewStepf("Duplicating the branch protection rules for %s", state.VitessRelease.ReleaseBranch)
		github.CopyBranchProtectionRules(state.VitessRelease.Repo, "vitessio", "main", state.VitessRelease.ReleaseBranch)

		pl.NewStepf("Duplicating the branch protection rules for %s", state.VitessRelease.BaseReleaseBranch)
		github.CopyBranchProtectionRules(state.VitessRelease.Repo, "vitessio", "main", state.VitessRelease.BaseReleaseBranch)
		return ""
	}
}
