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

func VtopUpdateCompatibilityTable(state *releaser.State) []string {
	return []string{
		fmt.Sprintf("You open a Pull Request that updates the compatibility table found in the README of https://github.com/%s", state.VtOpRelease.Repo),
		fmt.Sprintf("Add a new row before the last row. This new row should include the v%s.* vitess-operator release and the v%s.0.*, along with the matching K8S version.", state.VtOpRelease.Release, state.VitessRelease.MajorRelease),
		fmt.Sprintf("Once the Pull Request, you may bypass the branch protection rules by changing the settings in https://github.com/%s/settings/branches", state.VtOpRelease.Repo),
	}
}
