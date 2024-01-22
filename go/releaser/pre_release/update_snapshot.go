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

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func UpdateSnapshotOnMain(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 10,
	}

	var done bool
	var url string
	return pl, func() string {
		defer func() {
			state.Issue.UpdateSnapshotOnMain.Done = done
			state.Issue.UpdateSnapshotOnMain.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRepo)
		git.Checkout("main")
		git.ResetHard(state.Remote, "main")

		nextNextRelease := releaser.FindVersionAfterNextRelease(state)
		snapshotRelease := fmt.Sprintf("%s-SNAPSHOT", nextNextRelease)

		snapshotUpdatePRName := fmt.Sprintf("Bump to `v%s` after the `v%s` release", snapshotRelease, state.Release)

		// look for existing PRs
		pl.NewStepf("Look for an existing Pull Request named '%s'", snapshotUpdatePRName)
		if _, url = github.FindPR(state.VitessRepo, snapshotUpdatePRName); url != "" {
			pl.TotalSteps = 5 // only 5 total steps in this situation
			pl.NewStepf("An opened Pull Request was found: %s", url)
			done = true
			return url
		}

		pl.NewStepf("Create new branch based on %s/%s", state.Remote, "main")
		newBranchName := git.FindNewGeneratedBranch(state.Remote, "main", "snapshot-update")

		pl.NewStepf("Update version.go")
		UpdateVersionGoFile(snapshotRelease)

		pl.NewStepf("Update the Java directory")
		UpdateJavaDir(snapshotRelease)

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Snapshot update: %s", snapshotUpdatePRName)) {
			pl.TotalSteps = 9 // only 9 total steps in this situation
			pl.NewStepf("Nothing to commit, seems like back to dev mode is already done")
			done = true
			return ""
		}
		git.Push(state.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  snapshotUpdatePRName,
			Body:   fmt.Sprintf("Includes the changes required to update the SNAPSHOT version (v%s) after the release of v%s.", snapshotRelease, state.Release),
			Branch: newBranchName,
			Base:   "main",
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}},
		}
		_, url = pr.Create(state.VitessRepo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return ""
	}
}
