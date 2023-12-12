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
	"github.com/charmbracelet/lipgloss"
)

var (
	cellStyle     = lipgloss.NewStyle().Foreground(lightGray)
	selectedStyle = lipgloss.NewStyle().Foreground(vitessOrange).Bold(true)
	headerStyle   = lipgloss.NewStyle().Foreground(white).AlignHorizontal(lipgloss.Center).Bold(true)
	bgStyle       = lipgloss.NewStyle().Background(darkGray).Foreground(lightGray)
	borderStyle   = lipgloss.NewStyle().Foreground(lightGray)
)

const (
	vitessOrange = lipgloss.Color("#DE6E39")
	white        = lipgloss.Color("#FFFFFF")
	darkGray     = lipgloss.Color("#767676")
	lightGray    = lipgloss.Color("#c7c7c7")
)
