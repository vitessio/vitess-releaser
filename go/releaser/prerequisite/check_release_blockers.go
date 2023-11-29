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

package prerequisite

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	gh "github.com/cli/go-gh/v2"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

type Issue struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Labels      []label `json:"labels"`
}

func FormatIssues(issues []Issue) []string {
	var prFmt []string
	for _, issue := range issues {
		prFmt = append(prFmt, fmt.Sprintf(" -> %s  %s", issue.URL, issue.Title))
	}
	return prFmt
}

func CheckReleaseBlockerIssues(ctx *releaser.Context) []Issue {
	vitess.CorrectCleanRepo(ctx.VitessRepo)

	byteRes, _, err := gh.Exec("issue", "list", "--json", "title,url,labels", "--repo", ctx.VitessRepo)
	if err != nil {
		log.Fatalf(err.Error())
	}
	var issues []Issue
	err = json.Unmarshal(byteRes.Bytes(), &issues)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var mustClose []Issue

	branchName := fmt.Sprintf("release-%s.0", ctx.MajorRelease)
	for _, i := range issues {
		for _, l := range i.Labels {
			if strings.HasPrefix(l.Name, "Release Blocker: ") && strings.Contains(l.Name, branchName) {
				mustClose = append(mustClose, i)
			}
		}
	}
	return mustClose
}
