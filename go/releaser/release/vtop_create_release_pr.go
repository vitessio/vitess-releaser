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
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vitessio/vitess-releaser/go/releaser"
	"github.com/vitessio/vitess-releaser/go/releaser/code_freeze"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/github"
	"github.com/vitessio/vitess-releaser/go/releaser/logging"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

const (
	vtopDefaultsFile       = "./pkg/apis/planetscale/v2/defaults.go"
	vtopInitialClusterFile = "./test/endtoend/operator/101_initial_cluster.yaml"
)

func VtopCreateReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	hasGoUpgradePR := strings.HasPrefix(state.Issue.VtopUpdateGolang.URL, "https://")

	pl := &logging.ProgressLogging{
		TotalSteps: 15,
	}

	if hasGoUpgradePR {
		pl.TotalSteps += 2
	}

	var done bool
	var urls []string
	var commitCount int
	return pl, func() string {
		defer func() {
			state.Issue.VtopCreateReleasePR.Done = done
			state.Issue.VtopCreateReleasePR.URLs = urls
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

		// 2. Check Go Upgrade PR
		// We must ensure that, if any, the golang upgrade PR has been merged
		// otherwise we cannot proceed with the creation of the Release PR.
		if hasGoUpgradePR {
			prNb := github.URLToNb(state.Issue.VtopUpdateGolang.URL)
			pl.NewStepf("Checking if %s is merged. Please merge it if not already done. This step will timeout in 2 minutes.", state.Issue.VtopUpdateGolang.URL)
			timeout := time.After(2 * time.Minute)
		outer:
			for {
				select {
				case <-time.After(5 * time.Second):
					if github.IsPRMerged(state.VtOpRelease.Repo, prNb) {
						break outer
					}
				case <-timeout:
					pl.TotalSteps = 5
					pl.NewStepf("This step has timeout, please merge the Pull Request %s and try again.", state.Issue.VtopUpdateGolang.URL)
					return ""
				}
			}
			pl.NewStepf("PR has been merged")
		}

		// 3. Figuring out the name of the release PR
		releasePRName := fmt.Sprintf("[%s] Release of `v%s`", state.VtOpRelease.ReleaseBranch, state.VtOpRelease.Release)
		if state.Issue.RC > 0 {
			releasePRName = fmt.Sprintf("[%s] Release of `v%s-RC%d`", state.VtOpRelease.ReleaseBranch, state.VtOpRelease.Release, state.Issue.RC)
		}

		// 4. Look for existing PRs
		pl.NewStepf("Look for an existing Release Pull Request named '%s'", releasePRName)
		if _, url := github.FindPR(state.VtOpRelease.Repo, releasePRName); url != "" {
			pl.TotalSteps = 5
			if hasGoUpgradePR {
				pl.TotalSteps += 2
			}
			pl.NewStepf("An opened Release Pull Request was found: %s", url)
			done = true
			urls = append(urls, url)
			return url
		}

		// 5. Create temporary branch for the release
		pl.NewStepf("Create temporary branch from %s", state.VtOpRelease.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch, "create-release")

		// 6. Update the vitess golang dependency with the new vitess tag
		pl.NewStepf("Update the golang dependency of vitess to tag %s", strings.ToLower(state.VitessRelease.Release))
		updateVitessDeps(state)
		if !git.CommitAll(fmt.Sprintf("Set vitess golang dependencies to %s", strings.ToLower(state.VitessRelease.Release))) {
			commitCount++
			git.Push(state.VtOpRelease.Remote, newBranchName)
		}

		releaseNameWithRC := releaser.AddRCToReleaseTitle(state.VtOpRelease.Release, state.Issue.RC)
		lowerReleaseName := strings.ToLower(releaseNameWithRC)

		// 7. Update the version file of vtop
		pl.NewStepf("Update version file to %s", lowerReleaseName)
		code_freeze.UpdateVtOpVersionGoFile(lowerReleaseName)
		if !git.CommitAll(fmt.Sprintf("Update the version file to %s", lowerReleaseName)) {
			commitCount++
			git.Push(state.VtOpRelease.Remote, newBranchName)
		}

		// 8. Find out what is the previous release of vitess
		pl.NewStepf("Figuring out what the previous release of vitess is")
		state.GoToVitess()
		vitessPreviousRelease := releaser.FindPreviousRelease(state.VitessRelease.Remote, state.VitessRelease.MajorRelease)
		state.GoToVtOp()

		// 9. Update test code with proper images
		pl.NewStepf("Update vitess-operator test code to use proper images")
		updateVtopTests(vitessPreviousRelease, strings.ToLower(state.VitessRelease.Release))
		if !git.CommitAll(fmt.Sprintf("Update test code to use proper image")) {
			commitCount++
			git.Push(state.VtOpRelease.Remote, newBranchName)
		}

		// 10. Tag the latest commit
		gitTag := fmt.Sprintf("v%s", lowerReleaseName)
		pl.NewStepf("Tag and push %s", gitTag)
		git.TagAndPush(state.VitessRelease.Remote, gitTag)

		// 11. Figure out what is the next vtop release for this branch
		nextRelease := findNextVtOpVersion(state.VtOpRelease.Release, state.Issue.RC)
		pl.NewStepf("Go back to dev mode with version = %s", nextRelease)
		code_freeze.UpdateVtOpVersionGoFile(nextRelease)
		if !git.CommitAll(fmt.Sprintf("Go back to dev mode")) {
			commitCount++
			git.Push(state.VtOpRelease.Remote, newBranchName)
		}

		if commitCount > 0 {
			// 12. Create the Pull Request
			pl.NewStepf("Create Pull Request")
			pr := github.PR{
				Title:  releasePRName,
				Body:   fmt.Sprintf("This Pull Request contains all the code for the %s release of vtop + the back to dev mode. Warning: the tag is made on one of the commits of this PR, you must **not** squash merge this PR.", lowerReleaseName),
				Branch: newBranchName,
				Base:   state.VtOpRelease.ReleaseBranch,
				Labels: []github.Label{},
			}
			_, url := pr.Create(state.VtOpRelease.Repo)
			pl.NewStepf("Pull Request created %s", url)
			urls = append(urls, url)
		} else {
			pl.TotalSteps -= 2
		}

		// 13. Create the release on the GitHub UI
		pl.NewStepf("Create the release on the GitHub UI")
		url := github.CreateRelease(state.VtOpRelease.Repo, gitTag, "", state.VtOpRelease.IsLatestRelease && state.Issue.RC == 0, state.Issue.RC > 0)
		pl.NewStepf("Done %s", url)
		urls = append(urls, url)

		done = true
		return url
	}
}

func updateVitessDeps(state *releaser.State) {
	if !strings.HasPrefix(state.VitessRelease.Repo, "vitessio/vitess") {
		// bailing out here, since we are doing a release on a fork / testing the vitess releaser
		// the release we did on vitess is not on vitessio/vitess and thus updating the deps of
		// vtop to the new release of vitess will fail
		return
	}

	currentReleaseSlice := strings.Split(state.VitessRelease.Release, ".")
	if len(currentReleaseSlice) != 3 {
		utils.BailOut(nil, "could not parse the version.go in vitessio/vitess, output: %s", state.VitessRelease.Release)
	}

	utils.Exec("go", "get", "-u", fmt.Sprintf("vitess.io/vitess@v0.%s.%s", currentReleaseSlice[0], strings.ToLower(currentReleaseSlice[2])))
	utils.Exec("go", "mod", "tidy")
}

/*
	function updateVitessImages() {
	  old_vitess_version=$1
	  new_vitess_version=$2

	  operator_files=$(find -E $ROOT/test/endtoend/operator/* -name "*.yaml" | grep -v "101_initial_cluster.yaml")
	  sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$new_vitess_version\3/g" $operator_files
	  sed -i.bak -E "s/vitess\/vtadmin:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/vtadmin:v$new_vitess_version\3/g" $operator_files
	  sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$new_vitess_version\3\"/g" $ROOT/pkg/apis/planetscale/v2/defaults.go
	  sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$old_vitess_version\3/g" $ROOT/test/endtoend/operator/101_initial_cluster.yaml

	  rm -f $(find -E $ROOT/test/endtoend/operator/ -name "*.yaml.bak") $ROOT/pkg/apis/planetscale/v2/defaults.go.bak
	}
*/
func updateVtopTests(vitessPreviousVersion, vitessNewVersion string) {
	testFiles := vtopTestFiles()

	// sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$new_vitess_version\3/g" $operator_files
	args := append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\\/lite:v%s\\3/g", vitessNewVersion)}, testFiles...)
	utils.Exec("sed", args...)

	// sed -i.bak -E "s/vitess\/vtadmin:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/vtadmin:v$new_vitess_version\3/g" $operator_files
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/vtadmin:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\\/vtadmin:v%s\\3/g", vitessNewVersion)}, testFiles...)
	utils.Exec("sed", args...)

	// sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$new_vitess_version\3\"/g" $ROOT/pkg/apis/planetscale/v2/defaults.go
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?(.*)/vitess\\/lite:v%s\\3\"/g", vitessNewVersion)}, vtopDefaultsFile)
	utils.Exec("sed", args...)

	// sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$old_vitess_version\3/g" $ROOT/test/endtoend/operator/101_initial_cluster.yaml
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\\/lite:v%s\\3/g", vitessPreviousVersion)}, vtopInitialClusterFile)
	utils.Exec("sed", args...)

	filesBackups := make([]string, 0, len(testFiles)+1)
	for _, file := range testFiles {
		filesBackups = append(filesBackups, fmt.Sprintf("%s.bak", file))
	}
	filesBackups = append(filesBackups, vtopInitialClusterFile+".bak")
	filesBackups = append(filesBackups, vtopDefaultsFile+".bak")
	args = append([]string{"-f"}, filesBackups...)
	utils.Exec("rm", args...)
}

func vtopTestFiles() []string {
	var files []string
	root := "./test/endtoend/operator/"
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".yaml") && !strings.Contains(path, "101_initial_cluster.yaml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		utils.BailOut(err, "failed to walk directory %s", root)
	}
	return files
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
