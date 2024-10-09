/*
Copyright 2024 The Vitess Authors.

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
	"github.com/vitessio/vitess-releaser/go/interactive/ui"
	"github.com/vitessio/vitess-releaser/go/releaser"
)

type boolMsg struct {
	name string
	msg  []string
}

func newBooleanMenu(ctx context.Context, rawMsg []string, stepName string, setInverse func(), isDone, ignore bool) *ui.MenuItem {
	state := releaser.UnwrapState(ctx)

	act := func(mi *ui.MenuItem) (*ui.MenuItem, tea.Cmd) {
		return mi, func() tea.Msg {
			return boolMsg{name: stepName, msg: rawMsg}
		}
	}

	update := func(mi *ui.MenuItem, msg tea.Msg) (*ui.MenuItem, tea.Cmd) {
		switch msg := msg.(type) {
		case boolMsg:
			if msg.name != stepName {
				return mi, nil
			}
			return mi, ui.PushDialog(&ui.DoneDialog{
				Title:    stepName,
				Message:  msg.msg,
				IsDone:   mi.IsDone,
				StepName: stepName,
			})
		case ui.DoneDialogAction:
			if string(msg) != stepName {
				return mi, nil
			}
			setInverse()
			mi.IsDone = !mi.IsDone
			pl, fn := mi.State.UploadIssue()
			return mi, tea.Batch(func() tea.Msg {
				fn()
				return tea.Msg("")
			}, ui.PushDialog(ui.NewProgressDialog("Updating the Release Issue", pl)))
		}
		return mi, nil
	}

	return &ui.MenuItem{
		State:  state,
		Name:   stepName,
		IsDone: isDone,
		Act:    act,
		Update: update,

		Ignore: ignore,
	}
}
