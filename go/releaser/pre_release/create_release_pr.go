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

package pre_release

import (
	"fmt"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CreateReleasePR(ctx *releaser.Context) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 0,
	}
	return pl, func() string {
		// setup
		git.CorrectCleanRepo(ctx.VitessRepo)
		nextRelease, branchName := releaser.FindNextRelease(ctx.MajorRelease)
		remote := git.FindRemoteName(ctx.VitessRepo)
		git.ResetHard(remote, branchName)

		// find new branch to create the release
		newBranchName := git.FindNewGeneratedBranch(remote, branchName, "create-release")

		// deactivate code freeze
		deactivateCodeFreeze()
		if git.CommitAll(fmt.Sprintf("Unfreeze branch %s", branchName)) {
			// TODO: handle
			return ""
		}
		git.Push(remote, newBranchName)

		// TODO: Generate the release notes

		// TODO: Do the version change throughout the code base

		return ""
	}
}
