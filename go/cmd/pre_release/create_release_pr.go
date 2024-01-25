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

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/releaser"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

var createReleasePR = &cobra.Command{
	Use:   "release-pr",
	Short: "Create the Release Pull Request",
	Run: func(cmd *cobra.Command, args []string) {
		state := releaser.UnwrapState(cmd.Context())
		_, freeze := pre_release.CreateReleasePR(state)
		out := freeze()
		fmt.Println("Release Pull Request created:", out)
	},
}
