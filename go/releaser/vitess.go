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

package releaser

import (
	"fmt"
	"strconv"
	"strings"

	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/utils"
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
func FindNextRelease(remote, majorRelease string, isVtOp bool, rc int) (currentRelease, releaseBranchName string, isLatestRelease, isFromMain, ga bool) {
	fnGetCurrentRelease := getCurrentReleaseVitess
	fnReleaseToMajor := releaseToMajorVitess
	releaseBranchName = fmt.Sprintf("release-%s.0", majorRelease)
	if isVtOp {
		fnGetCurrentRelease = getCurrentReleaseVtOp
		fnReleaseToMajor = releaseToMajorVtOp
		releaseBranchName = fmt.Sprintf("release-%s", majorRelease)
	}

	git.Checkout("main")
	git.ResetHard(remote, "main")

	currentRelease = fnGetCurrentRelease()
	mainMajor := fnReleaseToMajor(currentRelease)

	if isVtOp {
		mainMajorParts := strings.Split(mainMajor, ".")
		majorParts := strings.Split(majorRelease, ".")
		if len(mainMajorParts) == 2 && len(majorParts) == 2 {
			mainMajorNb, err := strconv.Atoi(mainMajorParts[1])
			if err != nil {
				utils.LogPanic(err, "failed to convert main minor release increment to an int (%s)", mainMajorParts[1])
			}
			majorNb, err := strconv.Atoi(majorParts[1])
			if err != nil {
				utils.LogPanic(err, "failed to convert CLI release argument's minor release increment to an int (%s)", majorParts[1])
			}
			if rc > 0 && mainMajorNb == majorNb || mainMajorNb+1 == majorNb {
				return fmt.Sprintf("%s.%d.0", mainMajorParts[0], majorNb), releaseBranchName, true, true, ga
			}
		}
	} else if mainMajor == majorRelease {
		return currentRelease, releaseBranchName, true, true, ga
	}

	// main branch does not match, let's try release branches
	git.Checkout(releaseBranchName)
	git.ResetHard(remote, releaseBranchName)

	currentRelease = fnGetCurrentRelease()
	major := fnReleaseToMajor(currentRelease)

	// if the current release and the wanted release are different, it means there is an
	// error, we were not able to find the proper branch / corresponding release
	if major != majorRelease {
		utils.LogPanic(nil, "on branch '%s', could not find the corresponding major release '%s'", releaseBranchName, majorRelease)
	}

	mainMajorNb, err := strconv.ParseFloat(mainMajor, 64)
	if err != nil {
		utils.LogPanic(err, "could not parse main branch major release number as a float (%s)", mainMajor)
	}
	majorNb, err := strconv.ParseFloat(major, 64)
	if err != nil {
		utils.LogPanic(err, "could not parse CLI major release argument as a float (%s)", major)
	}
	releaseParts := strings.Split(currentRelease, ".")
	if len(releaseParts) != 3 {
		utils.LogPanic(nil, "could not parse the found release: %s", currentRelease)
	}
	isLatest := mainMajorNb-1 == majorNb
	ga = releaseParts[1] == "0" && releaseParts[2] == "0"
	if isVtOp {
		isLatest = mainMajorNb == majorNb
		ga = releaseParts[2] == "0"
	}
	return currentRelease, releaseBranchName, isLatest, false, ga
}

func FindPreviousRelease(remote, currentMajor string) string {
	majorNb, err := strconv.Atoi(currentMajor)
	if err != nil {
		utils.LogPanic(err, "failed to convert the CLI major release argument to an int (%s)", currentMajor)
	}

	previousMajor := majorNb - 1
	previousReleaseBranch := fmt.Sprintf("release-%d.0", previousMajor)
	git.Checkout(previousReleaseBranch)
	git.ResetHard(remote, previousReleaseBranch)

	return getCurrentReleaseVitess()
}

func getCurrentReleaseVitess() string {
	// Execute the following command to find the version from the `version.go` file:
	// sed -n 's/.*versionName.*\"\([[:digit:]\.]*\).*\"/\1/p' ./go/vt/servenv/version.go
	out := utils.Exec("sed", "-n", "s/.*versionName.*\"\\([[:digit:]\\.]*\\).*\"/\\1/p", "./go/vt/servenv/version.go")
	return strings.ReplaceAll(out, "\n", "")
}

func getCurrentReleaseVtOp() string {
	// Execute the following command to find the version from the `version.go` file:
	// sed -n 's/.*Version.*\"\([[:digit:]\.]*\).*\"/\1/p' ./version/version.go
	out := utils.Exec("sed", "-n", "s/.*Version =.*\"\\([[:digit:]\\.]*\\).*\"/\\1/p", "./version/version.go")
	return strings.ReplaceAll(out, "\n", "")
}

func releaseToMajorVitess(release string) string {
	return release[:strings.Index(release, ".")]
}

func releaseToMajorVtOp(release string) string {
	parts := strings.Split(release, ".")
	if len(parts) != 3 {
		utils.LogPanic(nil, "expected the vtop version to have format x.x.x but was %s", release)
	}
	return fmt.Sprintf("%s.%s", parts[0], parts[1])
}
