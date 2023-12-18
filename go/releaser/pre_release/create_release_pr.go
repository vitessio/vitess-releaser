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
	"os/exec"
	"path/filepath"
	"strings"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CreateReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 8,
	}
	return pl, func() string {
		// setup
		git.CorrectCleanRepo(state.VitessRepo)
		nextRelease, branchName := releaser.FindNextRelease(state.MajorRelease)

		pl.NewStepf("Fetch from git remote")
		remote := git.FindRemoteName(state.VitessRepo)
		git.ResetHard(remote, branchName)

		// find new branch to create the release
		pl.NewStepf("Create temporary branch from %s", branchName)
		newBranchName := git.FindNewGeneratedBranch(remote, branchName, "create-release")

		// deactivate code freeze
		pl.NewStepf("Deactivate code freeze on %s", branchName)
		deactivateCodeFreeze()

		pl.NewStepf("Commit unfreezing the branch %s", branchName)
		if git.CommitAll(fmt.Sprintf("Unfreeze branch %s", branchName)) {
			// TODO: handle
			return ""
		}
		git.Push(remote, newBranchName)

		pl.NewStepf("Generate the release notes")
		generateReleaseNotes(state, nextRelease)

		pl.NewStepf("Commit the release notes")
		if git.CommitAll("Addition of release notes") {
			// TODO: handle
			return ""
		}
		git.Push(remote, newBranchName)

		pl.NewStepf("Update the examples")
		// TODO: handle vtop version
		updateExamples(nextRelease, "")

		// TODO: Do the version change throughout the code base

		pl.NewStepf("...")
		return ""
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
	dirs := []string{"./examples/compose/", "./examples/operator"}
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
			log.Fatal(err.Error())
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
	out, err := exec.Command("sed", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/vtadmin:(.*)/vitess\/vtadmin:v$1/g" $compose_example_files $compose_example_sub_files $vtop_example_files
	args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/vitess\\/vtadmin:(.*)/vitess\\/vtadmin:v%s/g", newVersion)}, files...)
	out, err = exec.Command("sed", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// modify the docker image tag used for planetscale/vitess-operator
	// only if we do a new release
	if vtopNewVersion != "" {
		// sed -i.bak -E "s/planetscale\/vitess-operator:(.*)/planetscale\/vitess-operator:v$2/g" $vtop_example_files
		args = append([]string{"-i.bak", "-E", fmt.Sprintf("s/planetscale\\/vitess-operator:(.*)/planetscale\\/vitess-operator:v%s/g", vtopNewVersion)}, files...)
		out, err = exec.Command("sed", args...).CombinedOutput()
		if err != nil {
			log.Fatalf("%s: %s", err, out)
		}
	}

	// remove backup files from sed
	filesBackups := make([]string, 0, len(files))
	for _, file := range files {
		filesBackups = append(filesBackups, fmt.Sprintf("%s.bak", file))
	}
	args = append([]string{"-f"}, filesBackups...)
	out, err = exec.Command("rm", args...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}
