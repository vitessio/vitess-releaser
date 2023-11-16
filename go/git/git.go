package git

import (
	"log"
	"os/exec"
	"strings"
)

func CheckCurrentRepo(repoWanted string) bool {
	out, err := exec.Command("git", "remote", "-v").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	outStr := string(out)
	return strings.Contains(outStr, repoWanted)
}

func CleanLocalState() bool {
	out, err := exec.Command("git", "status", "-s").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return len(out) == 0
}
