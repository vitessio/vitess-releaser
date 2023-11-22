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
	tea "github.com/charmbracelet/bubbletea"

	"vitess.io/vitess-releaser/go/releaser/pre_release"
)

func codeFreezeMenuItem() menuItem {
	return menuItem{
		name:   "Code freeze",
		act:    codeFreezeAct,
		init:   nil,
		update: codeFreezeUpdate,
	}
}

type codeFreezeUrl string

func codeFreezeUpdate(mi menuItem, msg tea.Msg) (menuItem, tea.Cmd) {
	url, ok := msg.(codeFreezeUrl)
	if !ok {
		return mi, nil
	}
	mi.state = string(url)
	return mi, nil
}

func codeFreezeAct(mi menuItem) (menuItem, tea.Cmd) {
	mi.state = "running..."
	pl, freeze := pre_release.CodeFreeze()
	return mi, tea.Batch(func() tea.Msg {
		return codeFreezeUrl(freeze())
	}, push(newProgressDialog("Code freeze", pl)))
}
