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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

const (
	regexpFindGolangVersionInVitess = `(?i).*goversion_min[[:space:]]*([0-9.]+).*`
	regexpFindGolangVersionInVtop   = `^go[[:space:]]*([0-9.]+).*`
)

func VtopUpdateGolang(state *releaser.State) (*logging.ProgressLogging, func() string) {
	pl := &logging.ProgressLogging{
		TotalSteps: 12,
	}

	var done bool
	var url string
	return pl, func() string {
		defer func() {
			state.Issue.VtopUpdateGolang.Done = done
			state.Issue.VtopUpdateGolang.URL = url
			pl.NewStepf("Update Issue %s on GitHub", state.IssueLink)
			_, fn := state.UploadIssue()
			issueLink := fn()

			pl.NewStepf("Issue updated, see: %s", issueLink)
		}()

		pl.NewStepf("Fetch from git remote vitess repository")
		git.CorrectCleanRepo(state.VitessRelease.Repo)
		git.ResetHard(state.VitessRelease.Remote, state.VitessRelease.ReleaseBranch)

		pl.NewStepf("Get Go version of vitess")
		vitessGoVersion := currentGolangVersionInVitess()

		state.GoToVtOp()
		defer state.GoToVitess()

		pl.NewStepf("Fetch from git remote vitess-operator repository")
		git.CorrectCleanRepo(state.VtOpRelease.Repo)
		git.ResetHard(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)

		pl.NewStepf("Get Go version of vitess-operator")
		vtopGoVersion := currentGolangVersionInVtop()

		if len(vitessGoVersion.Segments()) < 2 || len(vtopGoVersion.Segments()) < 2 {
			pl.TotalSteps = 7
			pl.NewStepf("Unable to use the golang version, vitess=%s, vtop=%s", vitessGoVersion.String(), vtopGoVersion.String())
			done = true
			url = "Unable to parse the Golang version"
			return ""
		}

		if vitessGoVersion.Original() == vtopGoVersion.Original() {
			pl.TotalSteps = 7
			pl.NewStepf("Nothing to update, both Golang version share the same minor version: vitess=%s, vtop=%s", vitessGoVersion.String(), vtopGoVersion.String())
			done = true
			return ""
		}

		goUpdatePRName := fmt.Sprintf("[%s] Update Golang version to `v%s`", state.VtOpRelease.ReleaseBranch, vitessGoVersion.String())

		// look for existing code freeze PRs
		pl.NewStepf("Look for an existing Go update Pull Request named '%s'", goUpdatePRName)
		if _, url = github.FindPR(state.VtOpRelease.Repo, goUpdatePRName); url != "" {
			pl.TotalSteps = 8
			pl.NewStepf("An opened Go update Request was found: %s", url)
			done = true
			return url
		}

		pl.NewStepf("Create new branch based on %s/%s", state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch)
		newBranchName := git.FindNewGeneratedBranch(state.VtOpRelease.Remote, state.VtOpRelease.ReleaseBranch, "go-upgrade")

		pl.NewStepf("Updating the Go version of the operator to %s", vitessGoVersion.String())
		updateGolangVersionForVtop(vitessGoVersion)

		pl.NewStepf("Commit and push to branch %s", newBranchName)
		if git.CommitAll(fmt.Sprintf("Update Go version to %s", vitessGoVersion.String())) {
			pl.TotalSteps = 11
			pl.NewStepf("Nothing to commit, seems like the update is already done")
			done = true
			return ""
		}
		git.Push(state.VtOpRelease.Remote, newBranchName)

		pl.NewStepf("Create Pull Request")
		pr := github.PR{
			Title:  goUpdatePRName,
			Body:   fmt.Sprintf("This Pull Request updates the Golang version to %s.", vitessGoVersion.String()),
			Branch: newBranchName,
			Base:   state.VtOpRelease.ReleaseBranch,
		}
		_, url = pr.Create(state.VtOpRelease.Repo)
		pl.NewStepf("Pull Request created %s", url)
		done = true
		return url
	}
}

func updateGolangVersionForVtop(targetGoVersion *version.Version) {
	utils.Exec("sed", "-i.bak", "-E", fmt.Sprintf("s/^go (.*)/go %s/g", targetGoVersion.String()), "go.mod")
	utils.Exec("rm", "-f", "go.mod.bak")

	utils.Exec("sed", "-i.bak", "-E", fmt.Sprintf("s/^FROM golang:(.*) AS build/FROM golang:%s AS build/g", targetGoVersion.String()), "build/Dockerfile.release")
	utils.Exec("rm", "-f", "build/Dockerfile.release.bak")

	utils.Exec("sed", "-i.bak", "-E", fmt.Sprintf("s/go(.*).linux-amd64.tar.gz/go%s.linux-amd64.tar.gz/g", targetGoVersion.String()), ".buildkite/pipeline.yml")
	utils.Exec("rm", "-f", ".buildkite/pipeline.yml.bak")

	workflowFiles := findVtopWorkflowFiles()
	args := append([]string{"-i.bak", "-E", fmt.Sprintf("s/go-version: (.*)/go-version: %s/g", targetGoVersion.String())}, workflowFiles...)
	utils.Exec("sed", args...)
	for _, file := range workflowFiles {
		utils.Exec("rm", "-f", fmt.Sprintf("%s.bak", file))
	}
}

func findVtopWorkflowFiles() []string {
	var files []string
	root := ".github/workflows/"
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".yaml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		utils.BailOut(err, "failed to walk the directory %s", root)
	}
	return files
}

func currentGolangVersionInVitess() *version.Version {
	buildFile := "build.env"
	contentRaw, err := os.ReadFile(buildFile)
	if err != nil {
		utils.BailOut(err, "failed to read the file %s", buildFile)
	}
	content := string(contentRaw)

	versre := regexp.MustCompile(regexpFindGolangVersionInVitess)
	versionStr := versre.FindStringSubmatch(content)
	if len(versionStr) != 2 {
		utils.BailOut(nil, "malformatted error, got: %v", versionStr)
	}
	v, err := version.NewVersion(versionStr[1])
	if err != nil {
		utils.BailOut(err, "failed to create new version with %s", versionStr[1])
	}
	return v
}

func currentGolangVersionInVtop() *version.Version {
	gomodFile := "go.mod"
	contentRaw, err := os.ReadFile(gomodFile)
	if err != nil {
		utils.BailOut(err, "failed to read file %s", gomodFile)
	}
	content := string(contentRaw)

	versre := regexp.MustCompile(regexpFindGolangVersionInVtop)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		versionStr := versre.FindStringSubmatch(line)
		if len(versionStr) != 2 {
			continue
		}
		v, err := version.NewVersion(versionStr[1])
		if err != nil {
			utils.BailOut(err, "failed to create new version with %s", versionStr[1])
		}
		return v
	}
	utils.BailOut(nil, "could not parse the %s", gomodFile)
	return nil
}
