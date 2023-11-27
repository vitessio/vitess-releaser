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

package prerequisite

import (
	"encoding/json"
	"fmt"
	"log"

	gh "github.com/cli/go-gh/v2"

	"vitess.io/vitess-releaser/go/releaser/state"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

type PR struct {
	BaseRefName string `json:"baseRefName"`
	Title       string `json:"title"`
	Url         string `json:"url"`
}

func FormatPRs(prs []PR) []string {
	var prFmt []string
	for _, pr := range prs {
		prFmt = append(prFmt, fmt.Sprintf(" -> %s  %s", pr.Url, pr.Title))
	}
	return prFmt
}

func CheckPRs(majorRelease string) []PR {
	vitess.CorrectCleanRepo()

	byteRes, _, err := gh.Exec("pr", "list", "--json", "title,baseRefName,url", "--repo", state.VitessRepo)
	if err != nil {
		log.Fatalf(err.Error())
	}
	var prs []PR
	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var mustClose []PR

	branchName := fmt.Sprintf("release-%s.0", majorRelease)
	for _, pr := range prs {
		if pr.BaseRefName == branchName {
			mustClose = append(mustClose, pr)
		}
	}
	return mustClose
}
