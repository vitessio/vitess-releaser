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

package github

import (
	"encoding/json"
	"fmt"
	"github.com/cli/go-gh"
	"log"
	"strings"
	"vitess.io/vitess-releaser/go/releaser/state"
)

type Issue struct {
	Title    string
	Body     string
	Labels   []string
	Assignee string
}

// Create will open the issue on GitHub and return the link of the newly created issue
func (i *Issue) Create() string {
	stdOut, _, err := gh.Exec(
		"issue", "create",
		"--repo", state.VitessRepo,
		"--title", i.Title,
		"--body", i.Body,
		"--label", strings.Join(i.Labels, ","),
		"--assignee", i.Assignee,
	)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ReplaceAll(stdOut.String(), "\n", "")
}

func GetReleaseIssue(majorVersion string) string {
	res, _, err := gh.Exec("issue", "list", "-l", "Type: Release", "--json", "title,url")
	if err != nil {
		log.Fatal(err.Error())
	}

	issues := []map[string]string{}
	err = json.Unmarshal(res.Bytes(), &issues)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, issue := range issues {
		title := issue["title"]
		if strings.HasPrefix(title, fmt.Sprintf("Release of v%s", state.MajorRelease)) {
			return issue["url"]
		}
	}

	return ""
}
