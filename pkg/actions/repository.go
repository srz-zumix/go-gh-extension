package actions

import "os"

func GetRepositoryFullName() string {
	return os.Getenv("GITHUB_REPOSITORY")
}

func GetRepositoryFullNameWithHost() string {
	return os.Getenv("GITHUB_SERVER_URL") + "/" + GetRepositoryFullName()
}
