package render

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/fatih/color"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

type DiffCommandBuilder func(left, right string, target string) string

func colorizeDiff(diff string) string {
	var result string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+ ") {
			result += color.GreenString(line) + "\n"
		} else if strings.HasPrefix(line, "- ") {
			result += color.RedString(line) + "\n"
		} else {
			result += line + "\n"
		}
	}
	return result
}

func (r *Renderer) RenderDiff(diff any, left, right any, commandBuilder DiffCommandBuilder) {
	if r.exporter != nil {
		r.RenderExportedData(diff)
		return
	}

	diffLines := getDiffLines(diff, left, right, commandBuilder)
	if r.Color {
		diffLines = colorizeDiff(diffLines)
	}
	r.writeLine(diffLines)
}

func getDiffLines(diff any, left, right any, commandBuilder DiffCommandBuilder) string {
	if d, ok := diff.(gh.RepositoryPermissionsDiffs); ok {
		return getRepositoryPermissionDiffsLines(d, left.(string), right.(string), commandBuilder)
	}
	if d, ok := diff.(gh.TeamPermissionsDiffs); ok {
		return getTeamPermissionDiffsLines(d, left.(repository.Repository), right.(repository.Repository), commandBuilder)
	}
	return ""
}

func getRepositoryPermissionDiffLines(d gh.RepositoryPermissionsDiff, leftTeamSlug, rightTeamSlug string, commandBuilder DiffCommandBuilder) string {
	var diff string
	fullName := d.GetFullName()
	leftPerm := gh.GetRepositoryPermissions(d.Left)
	rightPerm := gh.GetRepositoryPermissions(d.Right)
	command := commandBuilder(leftTeamSlug, rightTeamSlug, fullName)
	diff += fmt.Sprintf("diff --%s\n", command)
	if d.Left != nil && d.Right != nil {
		diff += fmt.Sprintf("--- %s %s\n", *d.Left.FullName, leftTeamSlug)
		diff += fmt.Sprintf("+++ %s %s\n", *d.Right.FullName, rightTeamSlug)
		diff += fmt.Sprintf("- %s\n", leftPerm)
		diff += fmt.Sprintf("+ %s\n", rightPerm)
	} else if d.Left != nil {
		diff += fmt.Sprintf("--- %s %s\n", *d.Left.FullName, leftTeamSlug)
		diff += "+++ /dev/null\n"
		diff += fmt.Sprintf("- %s\n", leftPerm)
	} else if d.Right != nil {
		diff += "--- /dev/null\n"
		diff += fmt.Sprintf("+++ %s %s\n", *d.Right.FullName, rightTeamSlug)
		diff += fmt.Sprintf("+ %s\n", rightPerm)
	}
	return diff
}
func getRepositoryPermissionDiffsLines(d gh.RepositoryPermissionsDiffs, leftTeamSlug, rightTeamSlug string, commandBuilder DiffCommandBuilder) string {
	var diffLines string
	for _, diff := range d {
		diffLines += getRepositoryPermissionDiffLines(diff, leftTeamSlug, rightTeamSlug, commandBuilder)
	}
	return diffLines
}

func getTeamPermissionDiffLines(d gh.TeamPermissionsDiff, leftRepo, rightRepo repository.Repository, commandBuilder DiffCommandBuilder) string {
	var diff string

	leftOwnerRepo := fmt.Sprintf("%s/%s", leftRepo.Owner, leftRepo.Name)
	rightOwnerRepo := fmt.Sprintf("%s/%s", rightRepo.Owner, rightRepo.Name)
	teamSlug := d.GetSlug()

	command := commandBuilder(leftOwnerRepo, rightOwnerRepo, teamSlug)
	diff += fmt.Sprintf("diff --%s\n", command)
	if d.Left != nil && d.Right != nil {
		diff += fmt.Sprintf("--- %s %s\n", teamSlug, leftOwnerRepo)
		diff += fmt.Sprintf("+++ %s %s\n", teamSlug, rightOwnerRepo)
		diff += fmt.Sprintf("- %s\n", *d.Left.Permission)
		diff += fmt.Sprintf("+ %s\n", *d.Right.Permission)
	} else if d.Left != nil {
		diff += fmt.Sprintf("--- %s %s\n", teamSlug, leftOwnerRepo)
		diff += "+++ /dev/null\n"
		diff += fmt.Sprintf("- %s\n", *d.Left.Permission)
	} else if d.Right != nil {
		diff += "--- /dev/null\n"
		diff += fmt.Sprintf("+++ %s %s\n", teamSlug, rightOwnerRepo)
		diff += fmt.Sprintf("+ %s\n", *d.Right.Permission)
	}

	return diff
}

// GetDiffLines generates a diff representation for all team permissions differences.
func getTeamPermissionDiffsLines(d gh.TeamPermissionsDiffs, leftRepo, rightRepo repository.Repository, commandBuilder DiffCommandBuilder) string {
	var result string
	for _, diff := range d {
		result += getTeamPermissionDiffLines(diff, leftRepo, rightRepo, commandBuilder)
	}
	return result
}
