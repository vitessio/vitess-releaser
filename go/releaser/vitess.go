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

package releaser

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"vitess.io/vitess-releaser/go/releaser/git"
)

// FindNextRelease finds the next release that needs to be released for the given
// major release increment. And also tell whether this release is going to be the
// latest release of Vitess or not.
//
// First, it tries to figure out if the major release we want to use is on main, if
// it is, it returns the SNAPSHOT version of the main branch.
//
// Secondly, if the release we want to use is not on the main branch, it checks out
// to a release branch matching the given major release number. The SNAPSHOT version
// on that release branch is then returned.
func FindNextRelease(remote, majorRelease string) (string, string, bool) {
	git.Checkout("main")
	git.ResetHard(remote, "main")

	currentRelease := getCurrentRelease()
	mainMajor := releaseToMajor(currentRelease)

	if mainMajor == majorRelease {
		return currentRelease, "main", true
	}

	// main branch does not match, let's try release branches
	branchName := fmt.Sprintf("release-%s.0", majorRelease)
	git.Checkout(branchName)
	git.ResetHard(remote, branchName)

	currentRelease = getCurrentRelease()
	major := releaseToMajor(currentRelease)

	// if the current release and the wanted release are different, it means there is an
	// error, we were not able to find the proper branch / corresponding release
	if major != majorRelease {
		log.Fatalf("on branch '%s', could not find the corresponding major release '%s'", branchName, majorRelease)
	}

	mainMajorNb, err := strconv.Atoi(mainMajor)
	if err != nil {
		log.Fatal(err)
	}
	majorNb, err := strconv.Atoi(major)
	if err != nil {
		log.Fatal(err)
	}
	return currentRelease, branchName, mainMajorNb-1 == majorNb
}

func getCurrentRelease() string {
	// Execute the following command to find the version from the `version.go` file:
	// sed -n 's/.*versionName.*\"\([[:digit:]\.]*\).*\"/\1/p' ./go/vt/servenv/version.go
	out, err := exec.Command("sed", "-n", "s/.*versionName.*\"\\([[:digit:]\\.]*\\).*\"/\\1/p", "./go/vt/servenv/version.go").CombinedOutput()
	if err != nil {
		log.Fatalf("%v: %s", err, out)
	}

	outStr := string(out)
	return strings.ReplaceAll(outStr, "\n", "")
}

func releaseToMajor(release string) string {
	return release[:strings.Index(release, ".")]
}

func CorrectCleanRepo(repo string) {
	if !git.CheckCurrentRepo(repo + ".git") {
		log.Fatalf("the tool should be run from the %s repository directory", repo)
	}
	if !git.CleanLocalState() {
		log.Fatal("the vitess repository should have a clean state")
	}
}
