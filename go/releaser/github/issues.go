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

package github

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	gh "github.com/cli/go-gh"
	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
)

func execGh(args ...string) string {
	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		cmd := append([]string{"gh"}, strings.Join(args, " "))
		utils.BailOut(err, "failed to execute: %s, got: %s", strings.Join(cmd, " "), stdOut.String()+stdErr.String())
	}
	return stdOut.String()
}

func execGhWithError(args ...string) (string, error) {
	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		cmd := append([]string{"gh"}, strings.Join(args, " "))
		return stdOut.String(), fmt.Errorf("%w: failed to execute: %s, got: %s", err, strings.Join(cmd, " "), stdOut.String()+stdErr.String())
	}
	return stdOut.String(), nil
}

type Issue struct {
	Title    string  `json:"title"`
	Body     string  `json:"body"`
	URL      string  `json:"url"`
	Labels   []Label `json:"labels"`
	Assignee string  `json:"assignee"`
	Number   int     `json:"number"`
}

func CloseReleaseIssue(repo string, nb int) {
	execGh(
		"issue", "close",
		"--repo", repo,
		strconv.Itoa(nb),
		"--reason", "completed",
		"--comment", fmt.Sprintf("Release completed."),
	)
}

// Create will open the issue on GitHub and return the link of the newly created issue
func (i *Issue) Create(repo string) string {
	var labels []string
	for _, label := range i.Labels {
		labels = append(labels, label.Name)
	}
	stdOut := execGh(
		"issue", "create",
		"--repo", repo,
		"--title", i.Title,
		"--body", i.Body,
		"--label", strings.Join(labels, ","),
		"--assignee", i.Assignee,
	)
	return strings.ReplaceAll(stdOut, "\n", "")
}

func (i *Issue) UpdateBody(repo string) string {
	stdOut := execGh(
		"issue", "edit",
		"--repo", repo,
		strconv.Itoa(i.Number), "-b", i.Body,
	)
	return strings.ReplaceAll(stdOut, "\n", "")
}

func GetIssueTitleAndBody(repo string, nb int) (string, string) {
	stdOut := execGh(
		"issue", "view",
		strconv.Itoa(nb),
		"--repo", repo,
		"--json",
		"title,body",
	)
	var i Issue
	err := json.Unmarshal([]byte(stdOut), &i)
	if err != nil {
		utils.BailOut(err, "failed to parse the issue number %d, got: %s", nb, stdOut)
	}
	return i.Title, i.Body
}

func GetReleaseIssue(repo, release string, rcIncrement int) (string, string) {
	stdOut := execGh(
		"issue", "list",
		"-l", "Type: Release",
		"--json", "title,url",
		"--repo", repo,
	)

	var issues []map[string]string
	err := json.Unmarshal([]byte(stdOut), &issues)
	if err != nil {
		utils.BailOut(err, "failed to parse the release issue, got: %s", stdOut)
	}

	for _, issue := range issues {
		title := issue["title"]
		prefix := "Release of `v"
		if strings.HasPrefix(title, fmt.Sprintf("%s%s", prefix, release)) {
			// If we have an RC increment but the title does not match the RC increment we skip this issue
			if rcIncrement > 0 && !strings.Contains(title, fmt.Sprintf("-RC%d", rcIncrement)) {
				continue
			}

			return issue["url"], strings.ReplaceAll(title[len(prefix):], "`", "")
		}
	}
	return "", ""
}

func GetReleaseIssueInfo(repo, majorRelease string, rcIncrement int) (nb int, url, release string) {
	url, release = GetReleaseIssue(repo, majorRelease, rcIncrement)
	if url == "" {
		// no issue found
		return 0, "", ""
	}
	nb = URLToNb(url)
	return
}

func URLToNb(url string) int {
	lastIdx := strings.LastIndex(url, "/")
	issueNbStr := url[lastIdx+1:]
	nb, err := strconv.Atoi(issueNbStr)
	if err != nil {
		utils.BailOut(err, "failed to convert the end of the GitHub URL to a number, got: %s", issueNbStr)
	}
	return nb
}

func FormatIssues(issues []Issue) []string {
	var prFmt []string
	for _, issue := range issues {
		prFmt = append(prFmt, fmt.Sprintf(" -> %s  %s", issue.URL, issue.Title))
	}
	return prFmt
}

func CheckReleaseBlockerIssues(repo, majorRelease string) map[string]any {
	git.CorrectCleanRepo(repo)

	stdOut := execGh("issue", "list", "--json", "title,url,labels", "--repo", repo)
	var issues []Issue
	err := json.Unmarshal([]byte(stdOut), &issues)
	if err != nil {
		utils.BailOut(err, "failed to parse the release blocker issue, got: %s", stdOut)
	}

	var mustClose []Issue

	branchName := fmt.Sprintf("release-%s.0", majorRelease)
	for _, i := range issues {
		for _, l := range i.Labels {
			if strings.HasPrefix(l.Name, "Release Blocker: ") && strings.Contains(l.Name, branchName) {
				mustClose = append(mustClose, i)
			}
		}
	}

	m := make(map[string]any, len(mustClose))
	for _, pr := range mustClose {
		nb := pr.URL[strings.LastIndex(pr.URL, "/")+1:]
		markdownURL := fmt.Sprintf("#%s", nb)
		m[markdownURL] = nil
	}
	return m
}

func LoadKnownIssues(repo, majorRelease string) []Issue {
	label := fmt.Sprintf("Known issue: %s", majorRelease)

	stdOut := execGh(
		"issue", "list",
		"--repo", repo,
		"--label", label,
		"--json", "title,number",
	)

	var knownIssues []Issue
	err := json.Unmarshal([]byte(stdOut), &knownIssues)
	if err != nil {
		utils.BailOut(err, "failed to parse known issues, got: %s", stdOut)
	}
	return knownIssues
}
