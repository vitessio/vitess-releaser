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

package main

import (
	"fmt"
	"github.com/vitessio/vitess-releaser/go/cmd"
	"os"
	"os/exec"
	"runtime/debug"
)

// On some shells the terminal is left in a bad state possibly because the debug output is large or
// it contains control characters. This function restores the terminal to a sane state on a panic.
func restoreTerminal() {
	cmdSane := exec.Command("stty", "sane")
	cmdSane.Stdin = os.Stdin
	if err := cmdSane.Run(); err != nil {
		fmt.Println("Failed to restore terminal to sane state:", err)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Vitess Releaser Panic:\n", r)
			debug.PrintStack()
			restoreTerminal()
		}
	}()

	cmd.Execute()
}
