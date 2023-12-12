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
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/steps"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func codeFreezeMenuItem(ctx context.Context) *menuItem {
	state := releaser.UnwrapState(ctx)
	return &menuItem{
		state:  state,
		name:   steps.CodeFreeze,
		act:    codeFreezeAct,
		update: codeFreezeUpdate,
		info:   state.Issue.CodeFreeze.URL,
		isDone: state.Issue.CodeFreeze.Done,
	}
}

type codeFreezeUrl string

func codeFreezeUpdate(mi *menuItem, msg tea.Msg) (*menuItem, tea.Cmd) {
	_, ok := msg.(codeFreezeUrl)
	if !ok {
		return mi, nil
	}

	mi.info = mi.state.Issue.CodeFreeze.URL
	mi.isDone = mi.state.Issue.CodeFreeze.Done
	return mi, nil
}

func codeFreezeAct(mi *menuItem) (*menuItem, tea.Cmd) {
	pl, freeze := pre_release.CodeFreeze(mi.state)
	return mi, tea.Batch(func() tea.Msg {
		return codeFreezeUrl(freeze())
	}, pushDialog(newProgressDialog("Code freeze", pl)))
}
