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

package logging

import (
	"fmt"
	"slices"
	"sync"
)

type ProgressLogging struct {
	mu               sync.Mutex
	Done, TotalSteps int
	StepsDone        []string
}

func (pl *ProgressLogging) GetDone() int {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	return pl.Done
}

func (pl *ProgressLogging) GetTotal() int {
	return pl.TotalSteps
}

func (pl *ProgressLogging) NewStepf(msg string, args ...any) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.Done++
	msgF := fmt.Sprintf(msg, args...)
	pl.StepsDone = append(pl.StepsDone, fmt.Sprintf("%d/%d - %s", pl.Done, pl.TotalSteps, msgF))
}

func (pl *ProgressLogging) GetStepInProgress() []string {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	return slices.Clone(pl.StepsDone)
}

func (pl *ProgressLogging) SetTotalStep(v int) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.TotalSteps = v
}