package actions

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func GetEventName() string {
	return os.Getenv("GITHUB_EVENT_NAME")
}

func GetEventJsonPath() string {
	return os.Getenv("GITHUB_EVENT_PATH")
}

type EventContext struct {
	Name   string `json:"name"`
	Action string `json:"action,omitempty"`
}

func GetEventContext() (*EventContext, error) {
	path := GetEventJsonPath()
	if path == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}
	var event EventContext
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint
	if err := yaml.NewDecoder(f).Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}
