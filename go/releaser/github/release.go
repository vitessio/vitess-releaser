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

package github

import (
	"fmt"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser/git"
)

func CreateRelease(repo, tag, notesFilePath string, latest, prerelease bool) (url string) {
	target := git.GetSHAForGitRef(tag)

	args := []string{
		"release", "create",
		"--repo", repo,
		"--title", fmt.Sprintf("Vitess %s", tag),
		"--target", target,
		"--verify-tag",
	}

	if notesFilePath != "" {
		args = append(args, "-F", notesFilePath)
	} else {
		args = append(args, "--generate-notes")
	}

	if latest {
		args = append(args, "--latest=true")
	} else {
		args = append(args, "--latest=false")
	}

	if prerelease {
		args = append(args, "--prerelease")
	}

	args = append(args, tag)
	stdOut, err := execGhWithError(args...)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, tag)
		}
		utils.BailOutE(err)
	}
	return strings.ReplaceAll(stdOut, "\n", "")
}
