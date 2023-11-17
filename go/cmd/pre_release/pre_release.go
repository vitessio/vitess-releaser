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

import "github.com/spf13/cobra"

func PreRelease() *cobra.Command {
	cmd := &cobra.Command{
		Use: "pre-release",
		Short: "Runs the pre-release steps of a release",
	}

	cmd.AddCommand(codeFreeze)
	return cmd
}