package model

import (
	"log"
	"vitess.io/vitess-releaser/go/git"
)

func CorrectCleanRepo() {
	if !git.CheckCurrentRepo("vitessio/vitess.git") {
		log.Fatal("the tool should be run from the vitessio/vitess repository directory")
	}
	if !git.CleanLocalState() {
		log.Fatal("the vitess repository should have a clean state")
	}
}
