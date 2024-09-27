package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/v65/github"
	"io"
	"log"
	"os"
)

const (
	// GITHUB_EVENT_NAME_ENV is the environment variable that contains the name of the event
	GITHUB_EVENT_NAME_ENV = "GITHUB_EVENT_NAME"
	// GITHUB_EVENT_PATH_ENV is the environment variable that contains the path to the GitHub event JSON file
	GITHUB_EVENT_PATH_ENV = "GITHUB_EVENT_PATH"
	// Commit status context value used for tests
	GITHUB_RELEASE_CONTEXT_COMMIT_STATUS = "release/tests"
)

func isIssueEvent() bool {
	eventName := os.Getenv(GITHUB_EVENT_NAME_ENV)
	log.Printf("event type: %s", eventName)
	return eventName == "issues"
}

func getGithubIssuesEventFromEnv(pathName *string) (event *github.IssuesEvent, err error) {
	path := os.Getenv(GITHUB_EVENT_PATH_ENV)
	if pathName != nil {
		path = *pathName
	}

	if len(path) == 0 {
		return nil, errors.New(GITHUB_EVENT_PATH_ENV + " is empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v, err: %v", path, err)
	}
	defer file.Close()

	jsonBytes, err := io.ReadAll(file)

	if err := json.Unmarshal(jsonBytes, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %v", err)
	}
	return
}
