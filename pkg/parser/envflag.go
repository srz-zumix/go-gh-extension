package parser

import (
	"os"
	"strings"
)

func IsEnableEnvFlag(name string) bool {
	flag := os.Getenv(name)
	if flag == "" {
		return false
	}
	value := strings.ToLower(strings.TrimSpace(flag))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func IsDisableEnvFlag(name string) bool {
	flag := os.Getenv(name)
	if flag == "" {
		return false
	}
	value := strings.ToLower(strings.TrimSpace(flag))
	return value == "0" || value == "false" || value == "no" || value == "off"
}
