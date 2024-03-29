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

	"vitess.io/vitess-releaser/go/releaser"
)

func CopyBranchProtectionRules(state *releaser.State) []string {
	return []string{
		fmt.Sprintf("Since we have created the new branch %s, we need to copy the branch protection rules from main into %s", state.VitessRelease.ReleaseBranch, state.VitessRelease.ReleaseBranch),
		fmt.Sprintf("To do this, head over to https://github.com/%s/settings/branches and create a new rule for branch %s", state.VitessRelease.Repo, state.VitessRelease.ReleaseBranch),
	}
}
