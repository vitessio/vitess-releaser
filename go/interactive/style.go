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
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	cellStyle     = lipgloss.NewStyle().Foreground(darkGray)
	selectedStyle = lipgloss.NewStyle().Foreground(hotPink)
	headerStyle   = lipgloss.NewStyle().Foreground(black)
	bgStyle       = lipgloss.NewStyle().Background(darkGray).Foreground(lightGray)
)

const (
	hotPink   = lipgloss.Color("#FF06B7")
	darkGray  = lipgloss.Color("#767676")
	black     = lipgloss.Color("#00000")
	lightGray = lipgloss.Color("#cccccc")
)
