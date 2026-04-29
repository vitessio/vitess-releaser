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
	"sort"
	"strconv"
	"strings"

	"github.com/vitessio/vitess-releaser/go/releaser/git"
	"github.com/vitessio/vitess-releaser/go/releaser/utils"
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

func (p *PR) Create(issueLink string, repo string) (nb int, url string) {
	var labels []string
	for _, label := range p.Labels {
		labels = append(labels, label.Name)
	}

	p.Body = fmt.Sprintf("%s\n\n> This Pull Request is part of %s", p.Body, issueLink)

	stdOut := execGh(
		"pr", "create",
		"--repo", repo,
		"--title", p.Title,
		"--body", p.Body,
		"--label", strings.Join(labels, ","),
		"--head", p.Branch,
		"--base", p.Base,
	)
	url = strings.ReplaceAll(stdOut, "\n", "")
	nb = URLToNb(url)

	return nb, url
}

func IsPRMerged(repo string, nb int) bool {
	stdOut := execGh(
		"pr", "view", strconv.Itoa(nb),
		"--repo", repo,
		"--json", "mergedAt",
	)

	// If the PR is not merged, the output of the gh command will be:
	// {
	//  "mergedAt": null
	// }
	//
	// We can then grep for "null", if present, the PR has not been merged yet.
	return !strings.Contains(stdOut, "null")
}

func CheckBackportToPRs(repo, branch string) map[string]any {
	git.CorrectCleanRepo(repo)

	stdOut := execGh("pr", "list", "--json", "title,baseRefName,url,labels", "--repo", repo)

	var prs []PR

	err := json.Unmarshal([]byte(stdOut), &prs)
	if err != nil {
		utils.BailOut(err, "failed to parse backport PRs, got: %s", stdOut)
	}

	var mustClose []PR

	for _, pr := range prs {
		if pr.Base == branch {
			mustClose = append(mustClose, pr)
		}

		for _, l := range pr.Labels {
			if strings.HasPrefix(l.Name, "Backport to: ") && strings.Contains(l.Name, branch) {
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

func CheckReleaseBlockerPRs(repo, majorRelease string) map[string]any {
	git.CorrectCleanRepo(repo)

	stdOut := execGh("pr", "list", "--json", "title,url,labels", "--repo", repo)

	var prs []PR

	err := json.Unmarshal([]byte(stdOut), &prs)
	if err != nil {
		utils.BailOut(err, "failed to parse the release blocker PRs, got: %s", stdOut)
	}

	var mustClose []PR

	branchName := fmt.Sprintf("release-%s.0", majorRelease)

	for _, i := range prs {
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

func FindPR(repo, prTitle string) (nb int, url string) {
	stdOut := execGh(
		"pr", "list",
		"--json", "url,title",
		"--repo", repo,
		"--search", prTitle,
		"--state", "open",
	)

	var prs []PR

	err := json.Unmarshal([]byte(stdOut), &prs)
	if err != nil {
		utils.BailOut(err, "failed to parse PRs, got: %s", stdOut)
	}

	for _, pr := range prs {
		if pr.Title == prTitle {
			return URLToNb(pr.URL), pr.URL
		}
	}

	return 0, ""
}

func GetMergedPRsAndAuthorsByMilestone(repo, milestone string) (prs []PR, authors []string) {
	stdOut := execGh(
		"pr", "list",
		"-s", "merged",
		"-S", fmt.Sprintf("milestone:%s", milestone),
		"--json", "number,title,labels,author",
		"--limit", "5000",
		"--repo", repo,
	)

	err := json.Unmarshal([]byte(stdOut), &prs)
	if err != nil {
		utils.BailOut(err, "failed to parse PRs, got: %s", stdOut)
	}

	// GitHub's Search API (used by `gh pr list -S "milestone:..."`) has a separately
	// indexed corpus that can silently return incomplete results. We've observed it
	// drop ~50 PRs from a milestone that contained ~525, with no error reported.
	// Cross-check against the milestone REST API (authoritative) and bail out if
	// they disagree, so the regression is caught at generation time rather than
	// after the changelog is committed.
	expected := getMergedPRNumbersInMilestoneViaREST(repo, milestone)

	got := make(map[int]bool, len(prs))
	for _, p := range prs {
		got[p.Number] = true
	}

	var missing []int
	for n := range expected {
		if !got[n] {
			missing = append(missing, n)
		}
	}

	if len(missing) > 0 {
		sort.Ints(missing)
		utils.BailOut(nil,
			"GitHub Search API returned %d PRs for milestone %s but the milestone REST API has %d. "+
				"The Search API index is incomplete and would silently corrupt the changelog. "+
				"Missing PR numbers: %v. Investigate and retry.",
			len(prs), milestone, len(expected), missing)
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

// getMergedPRNumbersInMilestoneViaREST returns the set of merged PR numbers
// belonging to the given milestone, using the issues REST API rather than the
// Search API. The issues endpoint is authoritative for milestone membership;
// the Search API can silently lag or drop entries.
func getMergedPRNumbersInMilestoneViaREST(repo, milestoneTitle string) map[int]bool {
	msNumber := getMilestoneNumberByTitle(repo, milestoneTitle)

	// `gh api --paginate --jq` runs the jq filter against each page of results
	// and emits one line per match. Paginates 100 issues at a time.
	stdOut := execGh(
		"api",
		fmt.Sprintf("repos/%s/issues?milestone=%d&state=closed&per_page=100", repo, msNumber),
		"--paginate",
		"--jq", ".[] | select(.pull_request != null and .pull_request.merged_at != null) | .number",
	)

	nums := map[int]bool{}
	for _, line := range strings.Split(stdOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		n, err := strconv.Atoi(line)
		if err != nil {
			utils.BailOut(err, "could not parse PR number %q from milestone REST API", line)
		}

		nums[n] = true
	}

	return nums
}

// getMilestoneNumberByTitle resolves a milestone title (e.g. "v24.0.0") to its
// numeric ID via the REST API. We avoid the `gh milestone` extension here so
// the sanity check works on a plain `gh` install.
func getMilestoneNumberByTitle(repo, title string) int {
	stdOut := execGh(
		"api",
		fmt.Sprintf("repos/%s/milestones?state=all&per_page=100", repo),
		"--paginate",
		"--jq", fmt.Sprintf(`.[] | select(.title == %q) | .number`, title),
	)

	line := strings.TrimSpace(stdOut)
	if line == "" {
		utils.BailOut(nil, "milestone %q not found in repo %s", title, repo)
	}

	// If multiple lines came back the title is ambiguous — bail.
	if strings.Contains(line, "\n") {
		utils.BailOut(nil, "milestone title %q matched multiple milestones in repo %s: %s", title, repo, line)
	}

	n, err := strconv.Atoi(line)
	if err != nil {
		utils.BailOut(err, "could not parse milestone number %q for title %s", line, title)
	}

	return n
}

func GetOpenedPRsByMilestone(repo, milestone string) []PR {
	stdOut := execGh(
		"pr", "list",
		"-s", "open",
		"-S", fmt.Sprintf("milestone:%s", milestone),
		"--json", "number,title,labels,author",
		"--limit", "5000",
		"--repo", repo,
	)

	var prs []PR

	err := json.Unmarshal([]byte(stdOut), &prs)
	if err != nil {
		utils.BailOut(err, "failed to parse PRs, got: %s", stdOut)
	}

	return prs
}

func AssignMilestoneToPRs(repo, milestone string, prs []PR) {
	for _, pr := range prs {
		execGh(
			"pr", "edit",
			strconv.Itoa(pr.Number),
			"--milestone", milestone,
			"--repo", repo,
		)
	}
}
