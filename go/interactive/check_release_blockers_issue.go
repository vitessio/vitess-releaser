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

package interactive

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/prerequisite"
)

type releaseBlockerIssues []string

func releaseBlockerIssuesMenuItem(ctx *releaser.Context) menuItem {
	return menuItem{
		ctx:    ctx,
		name:   "Ensure all release blocker Issues are closed",
		act:    releaseBlockerIssuesAct,
		update: releaseBlockerIssuesUpdate,
	}
}

func releaseBlockerIssuesAct(mi menuItem) (menuItem, tea.Cmd) {
	mi.state = "Checking Issues..."
	return mi, func() tea.Msg {
		issues := prerequisite.FormatIssues(prerequisite.CheckReleaseBlockerIssues(mi.ctx))
		return releaseBlockerIssues(issues)
	}
}

func releaseBlockerIssuesUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	is, ok := msg.(releaseBlockerIssues)
	if !ok {
		return mi, nil
	}
	mi.state = fmt.Sprintf("Done, %d Issues need to be closed.", len(is))
	if len(is) == 0 {
		return mi, nil
	}

	return mi, pushDialog(warningDialog{
		title:   "These Issues still need to be closed before we can continue",
		message: is,
	})
}
