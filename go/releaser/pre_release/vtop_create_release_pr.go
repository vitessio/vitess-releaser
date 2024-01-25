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
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

const (
	vtopDefaultsFile       = "./pkg/apis/planetscale/v2/defaults.go"
	vtopInitialClusterFile = "./test/endtoend/operator/101_initial_cluster.yaml"

	vtopVersionGoFile = "./go/vt/servenv/version.go"
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

func VtopCreateReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 15,
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

		state.GoToVtOp()
		defer state.GoToVitess()

		// setup
		pl.NewStepf("Fetch from git remote")
		git.CorrectCleanRepo(state.VtOpRelease.Repo)
		git.ResetHard(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)

		// We must ensure that, if any, the golang upgrade PR has been merged
		// otherwise we cannot proceed with the creation of the Release PR.
		if strings.HasPrefix(state.Issue.VtopUpdateGolang.URL, "https://") {
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
					pl.NewStepf("This step has timeout, please merge the Pull Request %s and try again.", state.Issue.VtopUpdateGolang.URL)
					return ""
				}
			}
			pl.NewStepf("PR has been merged")
		}

		releasePRName := fmt.Sprintf("[%s] Release of `v%s`", state.VtOpRelease.ReleaseBranch, state.VtOpRelease.Release)
		if state.Issue.RC > 0 {
			releasePRName = fmt.Sprintf("%s-RC%d", releasePRName, state.Issue.RC)
		}

		// look for existing PRs
		pl.NewStepf("Look for an existing Release Pull Request named '%s'", releasePRName)
		if _, url = github.FindPR(state.VtOpRelease.Repo, releasePRName); url != "" {
			pl.TotalSteps = 5 // only 5 total steps in this situation
			pl.NewStepf("An opened Release Pull Request was found: %s", url)
			done = true
			return url
		}

		// find new branch to create the release
		pl.NewStepf("Create temporary branch from %s", state.VtOpRelease.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch, "create-release")

		// set vitess go deps to the new tag we created
		updateVitessDeps(state)
		if !git.CommitAll(fmt.Sprintf("Set vitess golang dependencies to %s", strings.ToLower(state.VitessRelease.Release))) {
			git.Push(state.VitessRelease.Remote, newBranchName)
		}

		releaseNameWithRC := releaser.AddRCToReleaseTitle(state.VtOpRelease.Release, state.Issue.RC)
		lowerReleaseName := strings.ToLower(releaseNameWithRC)
		updateVtOpVersionGoFile(lowerReleaseName)

		state.GoToVitess()
		vitessPreviousRelease := releaser.FindPreviousRelease(state.VitessRelease.Remote, state.VitessRelease.MajorRelease)
		state.GoToVtOp()

		updateVtopTests(vitessPreviousRelease, strings.ToLower(state.VitessRelease.Release))

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  releasePRName,
			Body:   fmt.Sprintf(""),
			Branch: newBranchName,
			Base:   state.VtOpRelease.ReleaseBranch,
			Labels: []github.Label{},
		}
		_, url = pr.Create(state.VtOpRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return url
	}
}

func updateVitessDeps(state *releaser.State) {
	out, err := exec.Command("go", "get", "-u", fmt.Sprintf("vitess.io/vitess@%s", strings.ToLower(state.VitessRelease.Release))).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	out, err = exec.Command("go", "mod", "tidy").CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func updateVtOpVersionGoFile(newVersion string) {
	err := os.WriteFile(vtopVersionGoFile, []byte(fmt.Sprintf(vtopVersionGo, time.Now().Year(), newVersion)), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
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
	out, err := exec.Command("sed", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/vtadmin:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/vtadmin:v$new_vitess_version\3/g" $operator_files
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/vtadmin:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\\/vtadmin:v%s\\3/g", vitessNewVersion)}, testFiles...)
	out, err = exec.Command("sed", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$new_vitess_version\3\"/g" $ROOT/pkg/apis/planetscale/v2/defaults.go
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\\/lite:v%s\\3/g", vitessNewVersion)}, vtopDefaultsFile)
	out, err = exec.Command("sed", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\/lite:v$old_vitess_version\3/g" $ROOT/test/endtoend/operator/101_initial_cluster.yaml
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/lite:([^-]*)(-rc[0-9]*)?(-mysql.*)?/vitess\\/lite:v%s\\3/g", vitessPreviousVersion)}, vtopInitialClusterFile)
	out, err = exec.Command("sed", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	filesBackups := make([]string, 0, len(testFiles)+1)
	for _, file := range testFiles {
		filesBackups = append(filesBackups, fmt.Sprintf("%s.bak", file))
	}
	filesBackups = append(filesBackups, vtopInitialClusterFile+".bak")
	filesBackups = append(filesBackups, vtopDefaultsFile+".bak")
	args = append([]string{"-f"}, filesBackups...)
	out, err = exec.Command("rm", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}

func vtopTestFiles() []string {
	var files []string
	err := filepath.WalkDir("./test/endtoend/operator/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".yaml") && !strings.Contains(path, "101_initial_cluster.yaml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	return files
}
