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
	"strconv"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/code_freeze"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

func VtopBackToDev(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 8,
	}

	var done bool
	var url string
	return pl, func() string {
		defer func() {
			state.Issue.VtopBackToDevMode.Done = done
			state.Issue.VtopBackToDevMode.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		state.GoToVtOp()
		defer state.GoToVitess()

		// 1. Setup of the vtop codebase
		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VtOpRelease.Repo)
		git.ResetHard(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)

		// 2. Create temporary branch for the release
		pl.NewStepf("Create temporary branch from %s", state.VtOpRelease.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch, "back-to-dev")

		releaseNameWithRC := releaser.AddRCToReleaseTitle(state.VtOpRelease.Release, state.Issue.RC)
		lowerReleaseName := strings.ToLower(releaseNameWithRC)

		// 3. Figure out what is the next vtop release for this branch
		nextRelease := findNextVtOpVersion(state.VtOpRelease.Release, state.Issue.RC)

		// 4. Back to dev mode and commit
		pl.NewStepf("Go back to dev mode with version = %s", nextRelease)
		code_freeze.UpdateVtOpVersionGoFile(nextRelease)
		noCommit := git.CommitAll(fmt.Sprintf("Go back to dev mode"))
		if noCommit {
			done = true
			pl.TotalSteps -= 3
			return ""
		}

		// 5. Push back to dev mode
		pl.NewStepf("Pushing back to dev mode to %s", newBranchName)
		git.Push(state.VtOpRelease.Remote, newBranchName)

		// 6. Create the Pull Request
		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  fmt.Sprintf("[%s] Back to dev mode after the release of `v%s`", state.VtOpRelease.ReleaseBranch, lowerReleaseName),
			Body:   fmt.Sprintf("This Pull Request updates the %s branch to go back to dev mode after the release of v%s.", state.VtOpRelease.ReleaseBranch, lowerReleaseName),
			Branch: newBranchName,
			Base:   state.VtOpRelease.ReleaseBranch,
			Labels: []github.Label{},
		}
		_, url = pr.Create(state.IssueLink, state.VtOpRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)

		done = true
		return url
	}
}

func findNextVtOpVersion(version string, rc int) string {
	if rc > 0 {
		return version
	}
	segments := strings.Split(version, ".")
	if len(segments) != 3 {
		utils.BailOut(nil, "expected three segments when looking at the vtop version, got: %s", version)
	}

	segmentInts := make([]int, 0, len(segments))
	for _, segment := range segments {
		v, err := strconv.Atoi(segment)
		if err != nil {
			utils.BailOut(err, "failed to convert segment of the vtop version to an int: %s", segment)
		}
		segmentInts = append(segmentInts, v)
	}
	return fmt.Sprintf("%d.%d.%d", segmentInts[0], segmentInts[1], segmentInts[2]+1)
}
