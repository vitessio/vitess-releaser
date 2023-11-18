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

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/cmd/pre_release"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/state"

	"vitess.io/vitess-releaser/go/cmd/flags"
	"vitess.io/vitess-releaser/go/cmd/interactive"
	"vitess.io/vitess-releaser/go/cmd/prerequisite"
)

var (
	releaseVersion string
	live           = true
)

var rootCmd = &cobra.Command{
	Use:   "vitess-releaser",
	Short: "vitess-releaser - a tool for releasing vitess",
	Long:  "vitess-releaser - a tool for releasing vitess",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&releaseVersion, flags.MajorRelease, "r", "", "Number of the major release on which we want to create a new release.")
	rootCmd.PersistentFlags().BoolVar(&live, flags.RunLive, false, "If live is true, will run against vitessio/vitess. Otherwise everything is done against your personal repository")
	err := rootCmd.MarkPersistentFlagRequired(flags.MajorRelease)
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(prerequisite.Prerequisite())
	rootCmd.AddCommand(pre_release.PreRelease())
	rootCmd.AddCommand(interactive.Command())
}

func Execute() {
	err := rootCmd.ParseFlags(os.Args)
	if err != nil {
		panic(err)
	}

	if live {
		state.VitessRepo = "vitessio/vitess"
	} else {
		state.VitessRepo = github.CurrentUser() + "/vitess"
	}
	state.MajorRelease = releaseVersion

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
