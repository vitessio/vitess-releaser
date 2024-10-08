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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vitess.io/vitess-releaser/go/cmd/flags"
	"vitess.io/vitess-releaser/go/interactive"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/git"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

const VERSION = "v1.0"

var (
	releaseVersion     string
	vtopReleaseVersion string
	releaseDate        string
	rcIncrement        int
	live               = true
	help               bool
	version            bool

	rootCmd = &cobra.Command{
		Use:   "vitess-releaser",
		Short: "Tooling used to release new versions of Vitess",
		Run: func(cmd *cobra.Command, args []string) {
			if version {
				printVersionAndExit()
			}
			ctx := cmd.Context()
			state := releaser.UnwrapState(ctx)
			git.CorrectCleanRepo(state.VitessRelease.Repo)

			// TODO: The assumption that the Release Manager won't be
			// modifying the release issue while using vitess-releaser
			// is made here, perhaps there is a better way of doing it
			state.LoadIssue()

			interactive.MainScreen(ctx, state)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&releaseDate, flags.ReleaseDate, "d", "", "Date of the release with the format: YYYY-MM-DD. Required when initiating a release.")
	rootCmd.PersistentFlags().BoolVarP(&help, flags.Help, "h", false, "Displays this help.")
	rootCmd.PersistentFlags().BoolVar(&live, flags.RunLive, false, "If live is true, will run against vitessio/vitess and planetscale/vitess-operator. Otherwise everything is done against your own forks.")
	rootCmd.PersistentFlags().IntVarP(&rcIncrement, flags.RCIncrement, "", 0, "Define the release as an RC release, value is used to determine the number of the RC.")
	rootCmd.PersistentFlags().StringVarP(&releaseVersion, flags.MajorRelease, "r", "", "Number of the major release on which we want to create a new release.")
	rootCmd.PersistentFlags().StringVarP(&vtopReleaseVersion, flags.VtOpRelease, "", "", "Number of the major and minor release on which we want to create a new release, i.e. '2.11', leave empty for no vtop release.")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "v", false, "Prints the version and git commit hash.")

	err := cobra.MarkFlagRequired(rootCmd.PersistentFlags(), flags.MajorRelease)
	if err != nil {
		utils.BailOutE(err)
	}
}

func Execute() {
	err := rootCmd.ParseFlags(os.Args)
	if help {
		_ = rootCmd.Help()
		os.Exit(0)
	}
	if err != nil {
		utils.BailOutE(err)
	}

	if version {
		printVersionAndExit()
	}

	err = rootCmd.ValidateRequiredFlags()
	if err != nil {
		fmt.Println(err)
		_ = rootCmd.Help()
		os.Exit(1)
	}

	resetGHUser := utils.SetGHUser()
	defer resetGHUser()

	s := &releaser.State{}

	vitessRepo, vtopRepo := getGitRepos()

	vitessRelease, issueNb, issueLink := setUpVitessReleaseInformation(s, vitessRepo, rcIncrement)
	vtopRelease := setUpVtOpReleaseInformation(s, vtopRepo, rcIncrement)

	s.VitessRelease = vitessRelease
	s.VtOpRelease = vtopRelease
	s.IssueNbGH = issueNb
	s.IssueLink = issueLink
	s.Issue.RC = rcIncrement
	s.Issue.DoVtOp = s.VtOpRelease.Release != ""
	s.Issue.VtopRelease = s.VtOpRelease.Release
	s.Issue.GA = vitessRelease.GA

	setUpIssueDate(s)

	ctx := releaser.WrapState(context.Background(), s)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func setUpVitessReleaseInformation(s *releaser.State, repo string, rc int) (releaser.ReleaseInformation, int, string) {
	s.GoToVitess()

	git.CorrectCleanRepo(repo)

	remote := git.FindRemoteName(repo)
	release, releaseBranch, isLatestRelease, isFromMain, ga := releaser.FindNextRelease(remote, releaseVersion, false, rc)
	issueNb, issueLink, releaseFromIssue := github.GetReleaseIssueInfo(repo, releaseVersion, rcIncrement)

	// if we want to do an RC-1 release and the branch is different from `main`, something is wrong
	// and if we want to do an >= RC-2 release, the release as to be the latest AKA on the latest release branch
	if rcIncrement >= 1 && !isLatestRelease {
		utils.BailOut(nil, "wanted: RC %d but release branch was %s, latest release was %v and is from main is %v", rcIncrement, releaseBranch, isLatestRelease, isFromMain)
	}

	majorReleaseNb, err := strconv.Atoi(releaseVersion)
	if err != nil {
		utils.BailOut(err, "could not parse the release version")
	}

	vitessRelease := releaser.ReleaseInformation{
		Repo:          repo,
		Remote:        remote,
		ReleaseBranch: releaseBranch,
		// BaseReleaseBranch is the same as ReleaseBranch for Vitess post v21, maybe we can merge these two at a later date
		BaseReleaseBranch: releaseBranch,
		MajorRelease:      releaseVersion,
		MajorReleaseNb:    majorReleaseNb,
		IsLatestRelease:   isLatestRelease,
		Release:           releaseFromIssue,
		GA:                ga,
	}
	if vitessRelease.Release == "" {
		vitessRelease.Release = releaser.AddRCToReleaseTitle(release, rcIncrement)
	}
	return vitessRelease, issueNb, issueLink
}

func setUpVtOpReleaseInformation(s *releaser.State, repo string, rc int) releaser.ReleaseInformation {
	if vtopReleaseVersion == "" {
		return releaser.ReleaseInformation{}
	}

	s.GoToVtOp()
	defer s.GoToVitess()

	git.CorrectCleanRepo(repo)

	remote := git.FindRemoteName(repo)
	release, releaseBranch, isLatestRelease, _, _ := releaser.FindNextRelease(remote, vtopReleaseVersion, true, rc)

	vtopRelease := releaser.ReleaseInformation{
		Repo:            repo,
		Remote:          remote,
		Release:         release,
		ReleaseBranch:   releaseBranch,
		IsLatestRelease: isLatestRelease,
	}
	return vtopRelease
}

func setUpIssueDate(s *releaser.State) {
	// We only require the release date if the release issue does not exist on GH
	// If the issue already exist we ignore the flag, the value will be loaded from the Issue
	if s.IssueLink != "" {
		return
	}
	if releaseDate == "" {
		fmt.Println("--date flag missing")
		_ = rootCmd.Help()
		os.Exit(1)
	}
	parsedReleaseDate, err := time.Parse(time.DateOnly, releaseDate)
	if err != nil {
		utils.BailOutE(err)
	}
	s.Issue.Date = parsedReleaseDate
}

func getGitRepos() (vitessRepo, vtopRepo string) {
	if live {
		vitessRepo = "vitessio/vitess"
		vtopRepo = "planetscale/vitess-operator"
	} else {
		currentGitHubUser := github.CurrentUser()
		vitessRepo = currentGitHubUser + "/vitess"
		vitessRepo = "planetscale/vitess-private"
		fmt.Printf(">>>>>>>>>>> vitess repo is %s\n", vitessRepo)
		vtopRepo = currentGitHubUser + "/vitess-operator"
	}
	return
}

func printVersionAndExit() {
	commit, shortHash := getGitCommit()
	fmt.Printf("Version: %s.%s\n", VERSION, shortHash)
	fmt.Printf("Last Commit: %s\n", commit)
	os.Exit(0)
}

func getGitCommit() (string, string) {
	cmd := exec.Command("git", "show", "-s", "--format=%H%n%an%n%ae%n%ad%n%s", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown", "unknown"
	}
	parts := strings.SplitN(string(out), "\n", 5)
	if len(parts) < 5 {
		return "unknown", "unknown"
	}
	commitHash := parts[0]
	authorName := parts[1]
	authorEmail := parts[2]
	authorDate := parts[3]
	commitMessage := parts[4]
	return fmt.Sprintf("%s\nAuthor: %s <%s>\nDate: %s\n\n    %s", commitHash, authorName, authorEmail, authorDate, commitMessage), commitHash[:7]
}
