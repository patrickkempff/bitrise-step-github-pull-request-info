package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/bitrise-tools/go-steputils/tools"
)

// Config - variables should be defined in bitrise secrets.
type Config struct {
	AuthToken     string `env:"personal_access_token"`
	RepositoryURL string `env:"repository_url"`
	IssueNumber   string `env:"issue_number"`
	APIBaseURL    string `env:"api_base_url"`
}

// PR - used for pull request response from github
type PR struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func ownerAndRepo(url string) (string, string) {
	url = strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "git@")
	paths := strings.FieldsFunc(url, func(r rune) bool { return r == '/' || r == ':' })
	return paths[1], strings.TrimSuffix(paths[2], ".git")
}

func main() {

	var conf Config
	// Parse the environment variables to a config instance.
	if err := stepconf.Parse(&conf); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	stepconf.Print(conf)

	owner, repo := ownerAndRepo(conf.RepositoryURL)
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%s", conf.APIBaseURL, owner, repo, conf.IssueNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}

	req.SetBasicAuth(conf.AuthToken, "x-oauth-basic")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}

	var result PR
	json.NewDecoder(resp.Body).Decode(&result)

	if err := tools.ExportEnvironmentWithEnvman("BITRISE_GITHUB_PULL_REQUEST_INFO_TITLE", result.Title); err != nil {
		log.Errorf("Failed to generate output")
	}

	if err := tools.ExportEnvironmentWithEnvman("BITRISE_GITHUB_PULL_REQUEST_INFO_BODY", result.Body); err != nil {
		log.Errorf("Failed to generate output")
	}

	defer resp.Body.Close()
	os.Exit(0)
}
