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
	"io/fs"
	"path/filepath"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/code_freeze"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

const (
	examplesOperator = "./examples/operator"
	examplesCompose  = "./examples/compose/"
)

func CreateReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 13,
	}

	// If we are doing a patch release or doing a GA we have two more steps to take: unfreeze the branch and commit
	// RFC https://github.com/vitessio/vitess/issues/15586 defines this new process.
	unfreezeBranch := state.Issue.RC == 0
	if unfreezeBranch {
		pl.TotalSteps += 2
	}

	var done bool
	var url string
	var commitCount int
	return pl, func() string {
		defer func() {
			state.Issue.CreateReleasePR.Done = done
			state.Issue.CreateReleasePR.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		// setup
		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VitessRelease.Repo)
		git.ResetHard(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)

		releasePRName := fmt.Sprintf("[%s] Release of `v%s`", state.VitessRelease.ReleaseBranch, state.VitessRelease.Release)

		// look for existing PRs
		pl.NewStepf("Look for an existing Release Pull Request named '%s'", releasePRName)
		if _, url = github.FindPR(state.VitessRelease.Repo, releasePRName); url != "" {
			pl.TotalSteps = 5 // only 5 total steps in this situation
			pl.NewStepf("An opened Release Pull Request was found: %s", url)
			done = true
			return url
		}

		// find new branch to create the release
		pl.NewStepf("Create temporary branch from %s", state.VitessRelease.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch, "create-release")

		// deactivate code freeze
		if unfreezeBranch {
			pl.NewStepf("Deactivate code freeze on %s", state.VitessRelease.ReleaseBranch)
			code_freeze.DeactivateCodeFreeze()

			pl.NewStepf("Commit unfreezing the branch %s", state.VitessRelease.ReleaseBranch)
			if !git.CommitAll(fmt.Sprintf("Unfreeze branch %s", state.VitessRelease.ReleaseBranch)) {
				commitCount++
				git.Push(state.VitessRelease.Remote, newBranchName)
			}
		}

		pl.NewStepf("Generate the release notes")
		generateReleaseNotes(state, releaser.RemoveRCFromReleaseTitle(state.VitessRelease.Release))

		pl.NewStepf("Commit the release notes")
		if !git.CommitAll("Addition of release notes") {
			commitCount++
			git.Push(state.VitessRelease.Remote, newBranchName)
		}

		lowerRelease := strings.ToLower(state.VitessRelease.Release)
		pl.NewStepf("Update the code examples")
		updateExamples(lowerRelease, strings.ToLower(releaser.AddRCToReleaseTitle(state.VtOpRelease.Release, state.Issue.RC)))

		pl.NewStepf("Update version.go")
		releaser.UpdateVersionGoFile(lowerRelease)

		pl.NewStepf("Update the Java directory")
		releaser.UpdateJavaDir(lowerRelease)

		pl.NewStepf("Commit the update to the codebase for the v%s release", state.VitessRelease.Release)
		if !git.CommitAll(fmt.Sprintf("Update codebase for the v%s release", state.VitessRelease.Release)) {
			commitCount++
			git.Push(state.VitessRelease.Remote, newBranchName)
		}

		if commitCount == 0 {
			pl.TotalSteps = 14
			pl.NewStepf("Nothing was commit and pushed, seems like there is no need to create the Release Pull Request")
			done = true
			return ""
		}

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  releasePRName,
			Body:   fmt.Sprintf("Includes the release notes and release commit for the `v%s` release. Once this PR is merged, we will be able to tag `v%s` on the merge commit.", state.VitessRelease.Release, state.VitessRelease.Release),
			Branch: newBranchName,
			Base:   state.VitessRelease.ReleaseBranch,
			Labels: []github.Label{{Name: "Component: General"}, {Name: "Type: Release"}, {Name: "Do Not Merge"}},
		}
		_, url = pr.Create(state.VitessRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return url
	}
}

// findFilesRecursive will fetch the full list of files that have to be
// updated when modifying the Vitess examples, here is what it looks
// like using the sed command in bash:
//
//	compose_example_files=$(find -E ./examples/compose/* -regex ".*.(go|yml)")
//	compose_example_sub_files=$(find -E ./examples/compose/**/* -regex ".*.(go|yml)")
//	vtop_example_files=$(find -E ./examples/operator -name "*.yaml")
func findFilesRecursive() []string {
	var files []string
	dirs := []string{examplesCompose, examplesOperator}
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			utils.BailOut(err, "failed to find files recursively")
		}
	}
	return files
}

// updateExamples updates the Vitess examples to use the proper tag/version of
// Vitess, according to what we are releasing. Moreover, it changes the vitess-operator
// version used only if we do a new vitess-operator release.
func updateExamples(newVersion, vtopNewVersion string) {
	files := findFilesRecursive()

	// sed -i.bak -E "s/vitess\/lite:(.*)/vitess\/lite:v$1/g" $compose_example_files $compose_example_sub_files $vtop_example_files
	args := append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/lite:(.*)/vitess\\/lite:v%s/g", newVersion)}, files...)
	utils.Exec("sed", args...)

	// sed -i.bak -E "s/vitess\/vtadmin:(.*)/vitess\/vtadmin:v$1/g" $compose_example_files $compose_example_sub_files $vtop_example_files
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/vtadmin:(.*)/vitess\\/vtadmin:v%s/g", newVersion)}, files...)
	utils.Exec("sed", args...)

	// modify the docker image tag used for planetscale/vitess-operator
	// only if we do a new release
	if vtopNewVersion != "" {
		// sed -i.bak -E "s/planetscale\/vitess-operator:(.*)/planetscale\/vitess-operator:v$2/g" $vtop_example_files
		args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/planetscale\\/vitess-operator:(.*)/planetscale\\/vitess-operator:v%s/g", vtopNewVersion)}, files...)
		utils.Exec("sed", args...)
	}

	// remove backup files from sed
	filesBackups := make([]string, 0, len(files))
	for _, file := range files {
		filesBackups = append(filesBackups, fmt.Sprintf("%s.bak", file))
	}
	args = append([]string{"-f"}, filesBackups...)
	utils.Exec("rm", args...)
}
