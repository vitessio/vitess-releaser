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

var addPendingPRsToIssue = &cobra.Command{
	Use:   "add-pending-prs",
	Short: "Add all pending Pull Requests to the Release Issue",
	Run: func(cmd *cobra.Command, args []string) {
		_, add := issue.AddBackportPRs()
		link := add()
		fmt.Println("Done, link to the existing GitHub Issue: ", link)
	},
}