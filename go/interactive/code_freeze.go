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
	"vitess.io/vitess-releaser/go/interactive/state"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func codeFreezeMenuItem(ctx *releaser.Context) *menuItem {
	return &menuItem{
		ctx:    ctx,
		name:   steps.CodeFreeze,
		act:    codeFreezeAct,
		update: codeFreezeUpdate,
		isDone: state.ToDo, // TODO: read the initial state from the Release Issue on GitHub
	}
}

type codeFreezeUrl string

func codeFreezeUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	url, ok := msg.(codeFreezeUrl)
	if !ok {
		return mi, nil
	}
	mi.info = string(url)
	mi.isDone = state.Done
	return mi, nil
}

func codeFreezeAct(mi *menuItem) (*menuItem, tea.Cmd) {
	mi.info = "running..."
	pl, freeze := pre_release.CodeFreeze(mi.ctx)
	return mi, tea.Batch(func() tea.Msg {
		return codeFreezeUrl(freeze())
	}, pushDialog(newProgressDialog("Code freeze", pl)))
}
