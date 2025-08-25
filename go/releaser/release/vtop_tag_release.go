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

package release

import (
	"fmt"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
)

func VtopTagRelease(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 6,
	}

	return pl, func() string {
		state.GoToVtOp()
		defer state.GoToVitess()

		// 1. Setup of the vtop codebase
		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VtOpRelease.Repo)
		git.ResetHard(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)

		// 2. Tag the latest commit
		releaseNameWithRC := releaser.AddRCToReleaseTitle(state.VtOpRelease.Release, state.Issue.RC)
		lowerReleaseName := strings.ToLower(releaseNameWithRC)
		gitTag := fmt.Sprintf("v%s", lowerReleaseName)
		pl.NewStepf("Tag and push %s", gitTag)
		git.TagAndPush(state.VtOpRelease.Remote, gitTag)

		// 3. Create the release on the GitHub UI
		pl.NewStepf("Create the release on the GitHub UI")

		url := github.CreateRelease(
			state.VtOpRelease.Repo,
			gitTag,
			"",
			state.VtOpRelease.IsLatestRelease && state.Issue.RC == 0,
			state.Issue.RC > 0,
		)
		pl.NewStepf("Done %s", url)

		state.Issue.VtopTagRelease.Done = true
		state.Issue.VtopTagRelease.URL = url
		pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
		_, fn := state.UploadIssue()
		issueLink := fn()

		pl.NewStepf("Issue updated, see: %s", issueLink)

		return url
	}
}
