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
	"log"
	"sort"
	"strconv"
	"strings"

	gh "github.com/cli/go-gh"
	"vitess.io/vitess-releaser/go/releaser/git"
)

type Label struct {
	Name string `json:"name"`
}

type Author struct {
	Login string `json:"login"`
}

type PR struct {
	Title  string  `json:"title"`
	Body   string  `json:"body,omitempty"`
	Branch string  `json:"branch,omitempty"`
	Base   string  `json:"baseRefName"`
	URL    string  `json:"url"`
	Labels []Label `json:"labels"`
	Author Author  `json:"author"`
	Number int     `json:"number"`
}

func (p *PR) Create(repo string) (nb int, url string) {
	var labels []string
	for _, label := range p.Labels {
		labels = append(labels, label.Name)
	}
	stdOut, _, err := gh.Exec(
		"pr", "create",
		"--repo", repo,
		"--title", p.Title,
		"--body", p.Body,
		"--label", strings.Join(labels, ","),
		"--head", p.Branch,
		"--base", p.Base,
	)
	if err != nil {
		log.Panic(err)
	}
	url = strings.ReplaceAll(stdOut.String(), "\n", "")
	nb = URLToNb(url)
	return nb, url
}

func IsPRMerged(repo string, nb int) bool {
	stdOut, _, err := gh.Exec(
		"pr", "view", strconv.Itoa(nb),
		"--repo", repo,
		"--json", "mergedAt",
	)
	if err != nil {
		log.Panic(err)
	}

	// If the PR is not merged, the output of the gh command will be:
	// {
	//  "mergedAt": null
	// }
	//
	// We can then grep for "null", if present, the PR has not been merged yet.
	return !strings.Contains(stdOut.String(), "null")
}

func CheckBackportToPRs(repo, majorRelease string) map[string]any {
	git.CorrectCleanRepo(repo)

	byteRes, _, err := gh.Exec("pr", "list", "--json", "title,baseRefName,url,labels", "--repo", repo)
	if err != nil {
		log.Panicf(err.Error())
	}
	var prs []PR
	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Panicf(err.Error())
	}

	var mustClose []PR

	branchName := fmt.Sprintf("release-%s.0", majorRelease)
	for _, pr := range prs {
		if pr.Base == branchName {
			mustClose = append(mustClose, pr)
		}
		for _, l := range pr.Labels {
			if strings.HasPrefix(l.Name, "Backport to: ") && strings.Contains(l.Name, branchName) {
				mustClose = append(mustClose, pr)
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

func FindPR(repo, prTitle string) (nb int, url string) {
	byteRes, _, err := gh.Exec(
		"pr", "list",
		"--json", "url",
		"--repo", repo,
		"--search", prTitle,
		"--state", "open",
	)
	if err != nil {
		log.Panicf(err.Error())
	}
	var prs []PR
	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Panicf(err.Error())
	}
	if len(prs) != 1 {
		return 0, ""
	}
	url = prs[0].URL
	return URLToNb(url), url
}

func GetMergedPRsAndAuthorsByMilestone(repo, milestone string) (prs []PR, authors []string) {
	byteRes, _, err := gh.Exec(
		"pr", "list",
		"-s", "merged",
		"-S", fmt.Sprintf("milestone:%s", milestone),
		"--json", "number,title,labels,author",
		"--limit", "5000",
		"--repo", repo,
	)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Panic(err)
	}

	// Get the full list of distinct PRs authors and sort them
	authorMap := map[string]bool{}
	for _, pr := range prs {
		login := pr.Author.Login
		if ok := authorMap[login]; !ok {
			if !strings.HasPrefix(login, "@app/") {
				authors = append(authors, login)
			}
			authorMap[login] = true
		}
	}
	sort.Strings(authors)
	return prs, authors
}

func GetOpenedPRsByMilestone(repo, milestone string) []PR {
	byteRes, _, err := gh.Exec(
		"pr", "list",
		"-s", "open",
		"-S", fmt.Sprintf("milestone:%s", milestone),
		"--json", "number,title,labels,author",
		"--limit", "5000",
		"--repo", repo,
	)
	if err != nil {
		log.Panic(err)
	}

	var prs []PR
	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Panic(err)
	}
	return prs
}

func AssignMilestoneToPRs(repo, milestone string, prs []PR) {
	for _, pr := range prs {
		_, _, err := gh.Exec(
			"pr", "edit",
			strconv.Itoa(pr.Number),
			"--milestone", milestone,
			"--repo", repo,
		)
		if err != nil {
			log.Panic(err)
		}
	}
}
