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
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser"
)

func VtopManualUpdateMessage(state *releaser.State) []string {
	var urlVtopReleasePRMsg string
	var vtopHeadReleaseBranch string

	// The steps in the 'release' section are sequential, it is therefor not possible to not have a release PR.
	// Unless, there was a bug/issue or the release team manually modified the release issue.
	// In which case we might fail to find the release PR, and thus defaulting to the following message:
	if state.Issue.VtopCreateReleasePR.URL == "" {
		urlVtopReleasePRMsg = fmt.Sprintf("the '%s' release branch by creating a new PR", state.VtOpRelease.ReleaseBranch)
		vtopHeadReleaseBranch = state.VtOpRelease.ReleaseBranch

	}

	previousVitessRelease := releaser.FindPreviousRelease(state.VitessRelease.Remote, state.VitessRelease.MajorRelease)

	msg := []string{
		"We need to make manual changes to the test files of the vitess-operator to use the newest releases.",
		"Everything happen under the './test/endtoend/' directory.",
		"",
		fmt.Sprintf("Add the following changes to %s:", urlVtopReleasePRMsg),
		fmt.Sprintf("\t- Modify the 'verifyVtGateVersion' function calls to use '%s' as the argument on the following files:", strings.ToLower(state.VitessRelease.Release)),
		"\t\t- backup_restore_test.sh",
		"\t\t- vtorc_vtadmin_test.sh",
		"\t\t- backup_schedule_test.sh",
		"\t\t- unmanaged_tablet_test.sh",
		"",
		"\t- In the file 'upgrade_test.sh' there are two 'verifyVtGateVersion' calls:",
		fmt.Sprintf("\t\t- The first one must use '%s'.", previousVitessRelease),
		fmt.Sprintf("\t\t- The second one must use '%s'.", strings.ToLower(state.VitessRelease.Release)),
	}

	nextVitessMajorRelease := releaser.FindNextMajorRelease(state.VitessRelease.MajorRelease)
	if state.VtOpRelease.IsLatestRelease {
		msg = append(msg, []string{
			"",
			"Add the following changes to 'main' by creating a new PR:",
			fmt.Sprintf("\t- In the file '101_initial_cluster.yaml', we must use 'vitess/lite:v%s'.", strings.ToLower(state.VitessRelease.Release)),
			fmt.Sprintf("\t- Copy the 'operator-latest.yaml' file from the HEAD of '%s' into main's 'operator/operator.yaml'.", vtopHeadReleaseBranch),
			"\t- Once copied, modify 'operator/operator.yaml' with the following:",
			"\t\t- Remove the change that adds 'imagePullPolicy: Never'",
			fmt.Sprintf("\t\t- Update the image 'vitess-operator-pr:latest' to use 'planetscale/vitess-operator:v%s'.", strings.ToLower(state.Issue.VtopRelease)),
			fmt.Sprintf("\t- The 'verifyVtGateVersion' function calls must use '%s' as an argument on the following files:", nextVitessMajorRelease),
			"\t\t- backup_restore_test.sh",
			"\t\t- vtorc_vtadmin_test.sh",
			"\t\t- backup_schedule_test.sh",
			"\t\t- unmanaged_tablet_test.sh",
			"\t- In the file 'upgrade_test.sh' there are two 'verifyVtGateVersion' calls:",
			fmt.Sprintf("\t\t- The first one must use '%s'.", strings.ToLower(state.VitessRelease.Release)),
			fmt.Sprintf("\t\t- The second one must use '%s'.", nextVitessMajorRelease),
		}...)
	}
	return msg
}
