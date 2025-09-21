package actions

import "os"

func ActionName() string {
	return os.Getenv("GITHUB_ACTION")
}

func IsRunsOn() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func GetRunURL() string {
	return GetRepositoryFullNameWithHost() + "/actions/runs/" + os.Getenv("GITHUB_RUN_ID")
}
