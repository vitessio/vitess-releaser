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

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/cmd/flags"
	"vitess.io/vitess-releaser/go/cmd/interactive"
	"vitess.io/vitess-releaser/go/cmd/post_release"
	"vitess.io/vitess-releaser/go/cmd/pre_release"
	"vitess.io/vitess-releaser/go/cmd/prerequisite"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
)

var (
	releaseVersion string
	releaseDate    string
	live           = true
)

var rootCmd = &cobra.Command{
	Use:   "vitess-releaser",
	Short: "vitess-releaser - a tool for releasing vitess",
	Long:  "vitess-releaser - a tool for releasing vitess",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&releaseVersion, flags.MajorRelease, "r", "", "Number of the major release on which we want to create a new release.")
	rootCmd.PersistentFlags().StringVarP(&releaseDate, flags.ReleaseDate, "d", "", "Date of the release with the format: YYYY-MM-DD. Required when initiating a release.")
	rootCmd.PersistentFlags().BoolVar(&live, flags.RunLive, false, "If live is true, will run against vitessio/vitess. Otherwise everything is done against your personal repository")

	err := cobra.MarkFlagRequired(rootCmd.PersistentFlags(), flags.MajorRelease)
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(prerequisite.Prerequisite())
	rootCmd.AddCommand(pre_release.PreRelease())
	rootCmd.AddCommand(post_release.PostRelease())
	rootCmd.AddCommand(interactive.Command())
}

func Execute() {
	err := rootCmd.ParseFlags(os.Args)
	if err != nil {
		panic(err)
	}

	err = rootCmd.ValidateRequiredFlags()
	if err != nil {
		fmt.Println(err)
		_ = rootCmd.Help()
		os.Exit(1)
	}

	var s releaser.State

	if live {
		s.VitessRepo = "vitessio/vitess"
	} else {
		s.VitessRepo = github.CurrentUser() + "/vitess"
	}
	s.MajorRelease = releaseVersion

	git.CorrectCleanRepo(s.VitessRepo)

	remote := git.FindRemoteName(s.VitessRepo)
	release, releaseBranch, isLatestRelease := releaser.FindNextRelease(remote, s.MajorRelease)
	issueNb, issueLink, releaseFromIssue := github.GetReleaseIssueInfo(s.VitessRepo, s.MajorRelease)

	s.Remote = remote
	s.ReleaseBranch = releaseBranch
	s.IsLatestRelease = isLatestRelease
	s.IssueNbGH = issueNb
	s.IssueLink = issueLink
	s.Release = releaseFromIssue
	if releaseFromIssue == "" {
		s.Release = release
	}

	// We only require the release date if the release issue does not exist on GH
	// If the issue already exist we ignore the flag, the value will be loaded from the Issue
	if s.IssueLink == "" {
		if releaseDate == "" {
			fmt.Println("--date flag missing")
			_ = rootCmd.Help()
			os.Exit(1)
		}
		parsedReleaseDate, err := time.Parse(time.DateOnly, releaseDate)
		if err != nil {
			panic(err)
		}
		s.Issue.Date = parsedReleaseDate
	}

	ctx := releaser.WrapState(context.Background(), &s)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
