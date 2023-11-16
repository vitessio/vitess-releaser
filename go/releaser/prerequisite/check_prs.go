package prerequisite

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cli/go-gh/v2"
	"vitess.io/vitess-releaser/go/releaser/vitess"
)

type PR struct {
	BaseRefName string `json:"baseRefName"`
	Title       string `json:"title"`
	Url         string `json:"url"`
}

func CheckPRs(majorRelease string) []string {
	vitess.CorrectCleanRepo()

	byteRes, _, err := gh.Exec("pr", "list", "--json", "title,baseRefName,url")
	if err != nil {
		log.Fatalf(err.Error())
	}
	var prs []PR
	err = json.Unmarshal(byteRes.Bytes(), &prs)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var mustClose []string

	branchName := fmt.Sprintf("release-%s.0", majorRelease)
	for _, pr := range prs {
		if pr.BaseRefName == branchName {
			mustClose = append(mustClose, fmt.Sprintf(" -> %s  %s", pr.Url, pr.Title))
		}
	}
	return mustClose
}