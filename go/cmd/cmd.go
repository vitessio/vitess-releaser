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

package cmd

import (
	"fmt"
	"os"

	"vitess.io/vitess-releaser/go/cmd/flags"
	"vitess.io/vitess-releaser/go/cmd/prerequisite"

	"github.com/spf13/cobra"
)

var (
	releaseVersion string
)

var rootCmd = &cobra.Command{
	Use:   "vitess-releaser",
	Short: "vitess-releaser - a tool for releasing vitess",
	Long:  "vitess-releaser - a tool for releasing vitess",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&releaseVersion, flags.MajorRelease, "r", "", "Number of the major release on which we want to create a new release.")
	err := rootCmd.MarkPersistentFlagRequired(flags.MajorRelease)
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(prerequisite.Prerequisite())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
