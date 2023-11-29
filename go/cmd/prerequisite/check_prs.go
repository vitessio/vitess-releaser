/*
Copyright 2023 The Vitess Authors.

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

package prerequisite

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/releaser"

	"vitess.io/vitess-releaser/go/releaser/prerequisite"
)

var checkPRs = &cobra.Command{
	Use:     "check-prs",
	Aliases: []string{"pr"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := releaser.UnwrapCtx(cmd.Context())

		mustClose := prerequisite.FormatPRs(prerequisite.CheckPRs(ctx))

		if len(mustClose) == 0 {
			return
		}
		log.Fatalf(fmt.Sprintf("Still open PRs against the release branch:\n%s", strings.Join(mustClose, "\n")))
	},
}
