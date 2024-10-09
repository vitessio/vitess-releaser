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
	"os"
	"time"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

const (
	vtopVersionGoFile = "./version/version.go"
	vtopVersionGo     = `/*
Copyright %d PlanetScale Inc.

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

package version

// DO NOT EDIT
// THIS FILE IS AUTO-GENERATED DURING NEW RELEASES BY THE VITESS-RELEASER

var (
	Version = "%s"
)
`
)

func VtopBumpMainVersion(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 8,
	}

	var done bool
	var url string
	return pl, func() string {
		state.GoToVtOp()
		defer state.GoToVitess()

		defer func() {
			state.Issue.VtopBumpMainVersion.Done = done
			state.Issue.VtopBumpMainVersion.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VtOpRelease.Repo)
		git.ResetHard(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)

		bumpPRName := fmt.Sprintf("[main] Bump version.go to %s", state.VtOpRelease.Release)
		pl.NewStepf("Look for an existing Release Pull Request named '%s'", bumpPRName)
		if _, url = github.FindPR(state.VtOpRelease.Repo, bumpPRName); url != "" {
			pl.TotalSteps = 5
			pl.NewStepf("An opened Release Pull Request was found: %s", url)
			done = true
			return url
		}

		pl.NewStepf("Create temporary branch from main")
		newBranchName := git.FindNewGeneratedBranch(state.VtOpRelease.Remote, "main", "bump-main-version")

		pl.NewStepf("Bump version.go to %s", state.VtOpRelease.Release)
		UpdateVtOpVersionGoFile(state.VtOpRelease.Release)
		if !git.CommitAll(fmt.Sprintf("Go back to dev mode")) {
			git.Push(state.VtOpRelease.Remote, newBranchName)

			pl.NewStepf("Create Pull Request")
			pr := github.PR{
				Title:  bumpPRName,
				Body:   fmt.Sprintf("This Pull Request bumps the version/version.go file to %s", state.VtOpRelease.Release),
				Branch: newBranchName,
				Base:   "main",
				Labels: []github.Label{},
			}
			_, url = pr.Create(state.VtOpRelease.Repo)
			pl.NewStepf("Pull Request created %s", url)
		} else {
			pl.TotalSteps -= 2
		}

		done = true
		return ""
	}
}

func UpdateVtOpVersionGoFile(newVersion string) {
	err := os.WriteFile(vtopVersionGoFile, []byte(fmt.Sprintf(vtopVersionGo, time.Now().Year(), newVersion)), os.ModePerm)
	if err != nil {
		utils.BailOut(err, "failed to write to file %s", vtopVersionGoFile)
	}
}
