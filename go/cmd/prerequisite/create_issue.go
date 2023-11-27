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

package prerequisite

import (
	"fmt"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/releaser/issue"
)

// Create issue:
// - Make sure we are in the vitess repo
// - Make sure the git state is clean
// - Figure out the new release number
// - Create the issue for the corresponding release number
var createIssue = &cobra.Command{
	Use:   "create-issue",
	Short: "Create the release issue",
	Run: func(cmd *cobra.Command, args []string) {
		_, createIssueFn := issue.CreateReleaseIssue()
		link := createIssueFn()
		fmt.Println("Link to the new GitHub Issue: ", link)
	},
}
