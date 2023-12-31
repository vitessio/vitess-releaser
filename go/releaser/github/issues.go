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

package github

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	gh "github.com/cli/go-gh"
	"vitess.io/vitess-releaser/go/releaser/git"
)

type Issue struct {
	Title    string  `json:"title"`
	Body     string  `json:"body"`
	URL      string  `json:"url"`
	Labels   []Label `json:"labels"`
	Assignee string  `json:"assignee"`
	Number   int     `json:"number"`
}

func CloseReleaseIssue(repo string, nb int) {
	_, _, err := gh.Exec(
		"issue", "close",
		"--repo", repo,
		strconv.Itoa(nb),
		"--reason", "completed",
		"--comment", fmt.Sprintf("Release completed."),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// Create will open the issue on GitHub and return the link of the newly created issue
func (i *Issue) Create(repo string) string {
	var labels []string
	for _, label := range i.Labels {
		labels = append(labels, label.Name)
	}
	stdOut, _, err := gh.Exec(
		"issue", "create",
		"--repo", repo,
		"--title", i.Title,
		"--body", i.Body,
		"--label", strings.Join(labels, ","),
		"--assignee", i.Assignee,
	)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ReplaceAll(stdOut.String(), "\n", "")
}

func (i *Issue) UpdateBody(repo string) string {
	stdOut, _, err := gh.Exec(
		"issue", "edit",
		"--repo", repo,
		strconv.Itoa(i.Number), "-b", i.Body,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return strings.ReplaceAll(stdOut.String(), "\n", "")
}

func GetIssueBody(repo string, nb int) string {
	var i Issue
	stdOut, _, err := gh.Exec(
		"issue", "view",
		strconv.Itoa(nb),
		"--repo", repo,
		"--json",
		"body",
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = json.Unmarshal(stdOut.Bytes(), &i)
	if err != nil {
		log.Fatal(err.Error())
	}

	return i.Body
}

func GetReleaseIssue(repo, release string) (string, string) {
	res, _, err := gh.Exec(
		"issue", "list",
		"-l", "Type: Release",
		"--json", "title,url",
		"--repo", repo,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	var issues []map[string]string
	err = json.Unmarshal(res.Bytes(), &issues)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, issue := range issues {
		title := issue["title"]
		prefix := "Release of v"
		if strings.HasPrefix(title, fmt.Sprintf("%s%s", prefix, release)) {
			return issue["url"], title[len(prefix):]
		}
	}
	return "", ""
}

func GetReleaseIssueInfo(repo, majorRelease string) (nb int, url string, release string) {
	url, release = GetReleaseIssue(repo, majorRelease)
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
		log.Fatal(err.Error())
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

	byteRes, _, err := gh.Exec("issue", "list", "--json", "title,url,labels", "--repo", repo)
	if err != nil {
		log.Fatalf(err.Error())
	}
	var issues []Issue
	err = json.Unmarshal(byteRes.Bytes(), &issues)
	if err != nil {
		log.Fatalf(err.Error())
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

	byteRes, _, err := gh.Exec(
		"issue", "list",
		"--repo", repo,
		"--label", label,
		"--json", "title,number",
	)
	if err != nil {
		log.Fatal(err)
	}

	var knownIssues []Issue
	err = json.Unmarshal(byteRes.Bytes(), &knownIssues)
	if err != nil {
		log.Fatal(err)
	}
	return knownIssues
}
