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

package beforerelease

import (
	"fmt"

	"vitess.io/vitess-releaser/go/git"

	"github.com/spf13/cobra"
)

var createIssue = &cobra.Command{
	Use:   "create-issue",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.CheckCurrentRepo("vitessio/vitess.git") {
			return fmt.Errorf("The tool should be run from the vitessio/vitess repository directory")
		}
		if !git.CleanLocalState() {
			return fmt.Errorf("The vitess repository should have a clean state")
		}
		return nil
	},
}
