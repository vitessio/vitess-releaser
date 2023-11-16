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
	"fmt"
	"log"
	"strings"

	"github.com/cli/go-gh"
)

type Issue struct {
	Title  string
	Body   string
	Labels []string
}

func (i *Issue) CreateIssue() {
	stdOut, _, err := gh.Exec(
		"issue", "create",
		"--repo", "frouioui/vitess",
		"--title", i.Title,
		"--body", i.Body,
		"--label", strings.Join(i.Labels, ","),
		"--assignee", "@me",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stdOut.String())
}
