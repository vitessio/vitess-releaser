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

package interactive

import (
	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/logging"
	"vitess.io/vitess-releaser/go/releaser/steps"
)

func createReleasePRMenuItem(ctx *releaser.Context) *menuItem {
	return &menuItem{
		ctx:    ctx,
		name:   steps.CreateReleasePR,
		act:    createReleasePRAct,
		update: createReleasePRUpdate,
		info:   ctx.Issue.CreateReleasePR.URL,
		isDone: ctx.Issue.CreateReleasePR.Done,
	}
}

type createReleasePRUrl string

func createReleasePRUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	_, ok := msg.(createReleasePRUrl)
	if !ok {
		return mi, nil
	}

	mi.info = mi.ctx.Issue.CreateReleasePR.URL
	mi.isDone = mi.ctx.Issue.CreateReleasePR.Done
	return mi, nil
}

func createReleasePRAct(mi *menuItem) (*menuItem, tea.Cmd) {
	pl := &logging.ProgressLogging{TotalSteps: 0}
	return mi, tea.Batch(func() tea.Msg {
		return createReleasePRUrl("done")
	}, pushDialog(newProgressDialog("Create the Release Pull Request", pl)))
}
