/*
Copyright 2023 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pre_release

import (
	"errors"
	"fmt"
	"log"
	"os/exec"

	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/state"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

type codeFreezeStatus int

const (
	codeFreezeDeactivated codeFreezeStatus = iota
	codeFreezeActivated

	codeFreezeWorkflowFile = "./.github/workflows/code_freeze.yml"
)

// CodeFreeze will freeze the branch of the next release we want to release.
// The function returns the URL of the code freeze Pull Request, this Pull
// Request must be forced-merged by a Vitess maintainer, this step cannot be automated.
func CodeFreeze() string {
	vitess.CorrectCleanRepo()

	nextRelease, branchName := vitess.FindNextRelease(state.MajorRelease)
	log.Printf("next release is %s on %s", nextRelease, branchName)

	remote := git.FindRemoteName(state.VitessRepo)
	git.Pull(remote, branchName)

	newBranchName := findNewBranchForCodeFreeze(remote, branchName)
	log.Printf("code freeze branch is %s", newBranchName)
	activateCodeFreeze()

	git.CommitAll(fmt.Sprintf("Code Freeze of %s", branchName))
	git.Push(remote, newBranchName)
	log.Printf("pushed code freeze changes to workflow")

	pr := github.PR{
		Title:  fmt.Sprintf("[%s] Code Freeze for `v%s`", branchName, nextRelease),
		Body:   fmt.Sprintf("This Pull Request freezes the branch `%s` for `v%s`", branchName, nextRelease),
		Branch: newBranchName,
		Base:   branchName,
		Labels: []string{"Component: General", "Type: Release"},
	}
	return pr.Create()
}

func findNewBranchForCodeFreeze(remote, baseBranch string) string {
	remoteAndBase := fmt.Sprintf("%s/%s", remote, baseBranch)

	var newBranch string
	for i := 1; ; i++ {
		newBranch = fmt.Sprintf("%s-code-freeze-%d", baseBranch, i)
		err := git.CreateBranchAndCheckout(newBranch, remoteAndBase)
		if err != nil {
			if errors.Is(err, git.ErrBranchExists) {
				continue
			}
			log.Fatal(err)
		}
		break
	}
	return newBranch
}

func activateCodeFreeze() {
	changeCodeFreezeWorkflow(codeFreezeActivated)
}

func deactivateCodeFreeze() {
	changeCodeFreezeWorkflow(codeFreezeDeactivated)
}

func changeCodeFreezeWorkflow(s codeFreezeStatus) {
	// sed -i.bak -E "s/exit (.*)/exit 0/g" $code_freeze_workflow
	out, err := exec.Command("sed", "-i.bak", "-E", fmt.Sprintf("s/exit (.*)/exit %d/g", s), codeFreezeWorkflowFile).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}

	// remove backup file left by the sed command
	out, err = exec.Command("rm", "-f", fmt.Sprintf("%s.bak", codeFreezeWorkflowFile)).CombinedOutput()
	if err != nil {
		log.Fatalf("%s: %s", err, out)
	}
}
