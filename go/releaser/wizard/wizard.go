package wizard

import (
	"fmt"
	"github.com/briandowns/spinner"
	"math/rand"
	"time"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
	"vitess.io/vitess-releaser/go/releaser/slack"
)

func Gogo(ctx *releaser.Context) error {
	stepSpinner("Finding the current release issue", func() {
		ctx.IssueNbGH, ctx.IssueLink = github.GetReleaseIssueInfo(ctx.VitessRepo, ctx.MajorRelease)
	})

	if ctx.IssueNbGH == 0 {
		fmt.Printf("No release issue exists. Create one by running: \n> vitess-releaser -r %s prerequisite create-issue\n", ctx.MajorRelease)
		return nil
	}

	stepSpinner("Reading it", func() {
		ctx.LoadIssue()
	})

	if !ctx.Issue.SlackPreRequisite {
		message := `Release has not yet been announced on slack. This is next.
The following message must be posted on the #general and #releases OSS Slack channels:
%s
`
		fmt.Printf(message, slack.PostReleaseMessage(ctx))
		return nil
	}

	var msg string
	stepSpinner("Check PRs and Issues", func() {
		_, f := prerequisite.CheckAndAddPRsIssues(ctx)
		msg = f()
	})

	if !(ctx.Issue.CheckBackport.Done() && ctx.Issue.ReleaseBlocker.Done()) {
		fmt.Println(msg)
		fmt.Println("Close PRs and issues and run this again")
	}

	return nil
}

var r = rand.New(rand.NewSource(time.Now().Unix()))

func stepSpinner(suffix string, f func()) {
	s := spinner.New(
		spinner.CharSets[r.Intn(len(spinner.CharSets))],
		100*time.Millisecond,
		func(s *spinner.Spinner) {
			s.Prefix = suffix + " "
		},
	)
	s.Start()
	f()
	s.Stop()
	return
}
