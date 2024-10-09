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

package pre_release

import (
	"fmt"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
)

func VtopCreateBranch(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 3,
	}

	return pl, func() string {
		state.GoToVtOp()
		defer state.GoToVitess()

		git.CorrectCleanRepo(state.VtOpRelease.Repo)
		pl.NewStepf("Create branch %s", state.VtOpRelease.ReleaseBranch)
		err := git.CreateBranchAndCheckout(state.VtOpRelease.ReleaseBranch, fmt.Sprintf("%s/main", state.VtOpRelease.Remote))
		if err != nil {
			git.Checkout(state.VtOpRelease.ReleaseBranch)
			git.ResetHard(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)
		} else {
			git.Push(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)
		}
		state.Issue.VtopCreateBranch = true
		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)
		return ""
	}
}
