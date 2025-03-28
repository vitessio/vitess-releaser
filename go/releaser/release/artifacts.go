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

	"github.com/vitessio/vitess-releaser/go/releaser"
)

func CheckArtifacts(state *releaser.State) []string {
	return []string{
		fmt.Sprintf("Check that release artifacts were generated: at bottom of https://github.com/vitessio/vitess/releases/tag/%s.", state.GetTag()),
		"",
		"The workflow that builds the artifacts can be found here: https://github.com/vitessio/vitess/actions/workflows/create_release.yml",
		"This workflow must be green.",
	}
}
