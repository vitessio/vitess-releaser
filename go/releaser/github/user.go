package github

import (
	"encoding/json"
	"github.com/cli/go-gh"
	"log"
)

func CurrentUser() string {
	exec, _, err := gh.Exec("api", "user")
	if err != nil {
		log.Fatalf(err.Error())
	}
	x := map[string]any{}

	err = json.Unmarshal(exec.Bytes(), &x)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return x["login"].(string)
}
