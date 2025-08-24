package actions

import "os"

func ActionName() string {
	return os.Getenv("GITHUB_ACTION")
}

func IsRunsOn() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}
