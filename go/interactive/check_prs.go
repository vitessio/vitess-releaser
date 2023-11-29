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

type openPRs []string

func checkPRsMenuItem(ctx *releaser.Context) menuItem {
	return menuItem{
		ctx:    ctx,
		name:   "Backport Pull Requests: Check",
		act:    checkPRsAct,
		update: checkPRsUpdate,
	}
}

func checkPRsAct(mi menuItem) (menuItem, tea.Cmd) {
	mi.state = "Checking pull requests..."
	return mi, func() tea.Msg {
		prs := prerequisite.FormatPRs(prerequisite.CheckPRs(mi.ctx))
		return openPRs(prs)
	}
}

func checkPRsUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	prs, ok := msg.(openPRs)
	if !ok {
		return mi, nil
	}
	mi.state = fmt.Sprintf("Done, %d PRs need to be merged.", len(prs))
	if len(prs) == 0 {
		return mi, nil
	}

	return mi, pushDialog(warningDialog{
		title:   "These PRs still need to be closed before we can continue",
		message: prs,
	})
}
