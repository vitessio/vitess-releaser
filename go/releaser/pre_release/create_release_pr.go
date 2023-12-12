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
	"log"
	"os/exec"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/logging"
)

func CreateReleasePR(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 0,
	}
	return pl, func() string {
		// setup
		git.CorrectCleanRepo(state.VitessRepo)
		nextRelease, branchName := releaser.FindNextRelease(state.MajorRelease)
		remote := git.FindRemoteName(state.VitessRepo)
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

		generateReleaseNotes(state, nextRelease)
		if git.CommitAll("Addition of release notes") {
			// TODO: handle
			return ""
		}
		git.Push(remote, newBranchName)

		updateExamples(nextRelease, "")
		// TODO: Do the version change throughout the code base

		return ""
	}
}

// if [ "$2" != "" ]; then
//
//	sed -i.bak -E "s/planetscale\/vitess-operator:(.*)/planetscale\/vitess-operator:v$2/g" $vtop_example_files
//
// fi
// rm -f $(find -E $ROOT/examples/operator -regex ".*.(md|yaml).bak")
// rm -f $(find -E $ROOT/examples/compose/* -regex ".*.(go|yml).bak")
// rm -f $(find -E $ROOT/examples/compose/**/* -regex ".*.(go|yml).bak")
func updateExamples(newVersion, vtopNewVersion string) {
	//  compose_example_files=$(find -E $ROOT/examples/compose/* -regex ".*.(go|yml)")
	//  compose_example_sub_files=$(find -E $ROOT/examples/compose/**/* -regex ".*.(go|yml)")
	//  vtop_example_files=$(find -E $ROOT/examples/operator -name "*.yaml")

	// sed -i.bak -E "s/vitess\/lite:(.*)/vitess\/lite:v$1/g" $compose_example_files $compose_example_sub_files $vtop_example_files
	out, err := exec.Command(
		"sed", "-i.bak", "-E",
		fmt.Sprintf("s/vitess\\/lite:(.*)/vitess\\/lite:v%s/g", newVersion),
		"$(find -E $ROOT/examples/compose/* -regex \".*.(go|yml)\")",
		"$(find -E $ROOT/examples/compose/**/* -regex \".*.(go|yml)\")",
		"$(find -E $ROOT/examples/operator -name \"*.yaml\")",
	).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/vtadmin:(.*)/vitess\/vtadmin:v$1/g" $compose_example_files $compose_example_sub_files $vtop_example_files
	out, err = exec.Command(
		"sed", "-i.bak", "-E",
		fmt.Sprintf("\"s/vitess\\/vtadmin:(.*)/vitess\\/vtadmin:v%s/g\"", newVersion),
		"$(find -E $ROOT/examples/compose/* -regex \".*.(go|yml)\")",
		"$(find -E $ROOT/examples/compose/**/* -regex \".*.(go|yml)\")",
		"$(find -E $ROOT/examples/operator -name \"*.yaml\")",
	).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/lite:\${VITESS_TAG:-latest}/vitess\/lite:v$1/g" $compose_example_sub_files $vtop_example_files
	out, err = exec.Command(
		"sed", "-i.bak", "-E",
		fmt.Sprintf("s/vitess\\/lite:\\${VITESS_TAG:-latest}/vitess\\/lite:v%s/g", newVersion),
		"$(find -E $ROOT/examples/compose/**/* -regex \".*.(go|yml)\")",
		"$(find -E $ROOT/examples/operator -name \"*.yaml\")",
	).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// sed -i.bak -E "s/vitess\/lite:(.*)-mysql80/vitess\/lite:v$1-mysql80/g" $(find -E $ROOT/examples/operator -name "*.md")
	out, err = exec.Command(
		"sed", "-i.bak", "-E",
		fmt.Sprintf("s/vitess\\/lite:(.*)-mysql80/vitess\\/lite:v%s-mysql80/g", newVersion),
		"$(find -E $ROOT/examples/operator -name \"*.md\")",
	).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	//  rm -f $(find -E $ROOT/examples/operator -regex ".*.(md|yaml).bak")
	//  rm -f $(find -E $ROOT/examples/compose/* -regex ".*.(go|yml).bak")
	//  rm -f $(find -E $ROOT/examples/compose/**/* -regex ".*.(go|yml).bak")
	out, err = exec.Command(
		"rm", "-f",
		"$(find -E $ROOT/examples/operator -regex \".*.(md|yaml).bak\")",
		"$(find -E $ROOT/examples/compose/* -regex \".*.(go|yml).bak\")",
		"$(find -E $ROOT/examples/compose/**/* -regex \".*.(go|yml).bak\")",
	).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}
