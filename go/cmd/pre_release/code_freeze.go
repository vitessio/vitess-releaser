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
	"fmt"

	"github.com/spf13/cobra"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

// Code Freeze:
// - Checkout the proper branch
// - Find the remote of vitessio/vitess.git
// - Git pull from the remote
// - Run the code freeze script
// - Get the PR URL and prompt it to the user
var codeFreeze = &cobra.Command{
	Use:   "code-freeze",
	Short: "Does the code-freeze of a release",
	Run: func(cmd *cobra.Command, args []string) {
		out := pre_release.CodeFreeze()
		fmt.Println("Please force merge the Pull Request created for code freeze:", out)
	},
}
