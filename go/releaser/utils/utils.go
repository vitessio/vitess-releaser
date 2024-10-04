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

package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
)

func BailOut(err error, msg string, args ...interface{}) {
	fullMsg := fmt.Sprintf(msg, args...)
	if err == nil {
		log.Println(fullMsg)
	} else {
		log.Println(err.Error(), "\n", fullMsg, "\n", string(debug.Stack()))
	}
	os.Exit(1)
}

func BailOutE(err error) {
	log.Println(err, "\n", debug.Stack())
	os.Exit(1)
}

func Exec(cmd string, args ...string) string {
	command := exec.Command(cmd, args...)
	out, err := command.CombinedOutput()
	if err != nil {
		BailOut(err, "failed to execute: %s, got: %s", command.String(), string(out))
	}
	return string(out)
}

func ExecWithError(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	out, err := command.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%w: failed to execute: %s", err, command.String())
	}
	return string(out), nil
}

func SetGHUser() func() {
	ghToken := os.Getenv("VITESS_RELEASER_GH_TOKEN")
	if ghToken != "" {
		currentGHToken := os.Getenv("GH_TOKEN")
		os.Setenv("GH_TOKEN", ghToken)
		fmt.Println("Setting GH_TOKEN to VITESS_RELEASER_GH_TOKEN")
		return func() {
			os.Setenv("GH_TOKEN", currentGHToken)
		}
	}
	return func() {}
}
