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

	"github.com/cli/go-gh/v2"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

type PR struct {
	BaseRefName string `json:"baseRefName"`
	Title       string `json:"title"`
	Url         string `json:"url"`
}

func CheckPRs(majorRelease string) []string {
	vitess.CorrectCleanRepo()

	byteRes, _, err := gh.Exec("pr", "list", "--json", "title,baseRefName,url")
	if err != nil {
		log.Fatalf(err.Error())
	}
	var prs []PR
	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var mustClose []string

	branchName := fmt.Sprintf("release-%s.0", majorRelease)
	for _, pr := range prs {
		if pr.BaseRefName == branchName {
			mustClose = append(mustClose, fmt.Sprintf(" -> %s  %s", pr.Url, pr.Title))
		}
	}
	return mustClose
}