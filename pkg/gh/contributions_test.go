package gh

import (
	"reflect"
	"testing"
)

func TestMergeContributionChunks_SingleTruncatedChunkKeepsServerTotal(t *testing.T) {
	// A single chunk where the server reports more contributing repositories than
	// the 100-capped breakdown returns. The merged total must keep the accurate
	// server value (not the length of the truncated list) and flag truncation.
	repos := make([]RepositoryContributions, 100)
	for i := range repos {
		repos[i] = RepositoryContributions{NameWithOwner: string(rune('a'+i%26)) + string(rune('0'+i/26)), Contributions: 1}
	}
	chunk := &ContributionsCollection{
		TotalCommitContributions:                 500,
		TotalRepositoriesWithContributedCommits:  150,
		CommitContributionsByRepository:          repos,
		CommitContributionsByRepositoryTruncated: true,
	}

	got := mergeContributionChunks([]*ContributionsCollection{chunk})

	if got.TotalRepositoriesWithContributedCommits != 150 {
		t.Errorf("TotalRepositoriesWithContributedCommits = %d, want 150", got.TotalRepositoriesWithContributedCommits)
	}
	if !got.CommitContributionsByRepositoryTruncated {
		t.Error("CommitContributionsByRepositoryTruncated = false, want true")
	}
	if len(got.CommitContributionsByRepository) != 100 {
		t.Errorf("len(CommitContributionsByRepository) = %d, want 100", len(got.CommitContributionsByRepository))
	}
}

func TestMergeContributionChunks_MultipleChunksMergeAndSum(t *testing.T) {
	chunks := []*ContributionsCollection{
		{
			TotalCommitContributions:                10,
			TotalIssueContributions:                 1,
			TotalRepositoriesWithContributedCommits: 2,
			CommitContributionsByRepository: []RepositoryContributions{
				{NameWithOwner: "o/a", Contributions: 3},
				{NameWithOwner: "o/b", Contributions: 4},
			},
			ContributionCalendar: ContributionCalendar{
				TotalContributions: 11,
				Days:               []ContributionDay{{Date: "2024-01-01", Count: 5}},
			},
		},
		{
			TotalCommitContributions:                20,
			TotalIssueContributions:                 2,
			TotalRepositoriesWithContributedCommits: 2,
			CommitContributionsByRepository: []RepositoryContributions{
				{NameWithOwner: "o/b", Contributions: 6},
				{NameWithOwner: "o/c", Contributions: 7},
			},
			ContributionCalendar: ContributionCalendar{
				TotalContributions: 22,
				Days:               []ContributionDay{{Date: "2025-01-01", Count: 9}},
			},
		},
	}

	got := mergeContributionChunks(chunks)

	if got.TotalCommitContributions != 30 {
		t.Errorf("TotalCommitContributions = %d, want 30", got.TotalCommitContributions)
	}
	if got.TotalIssueContributions != 3 {
		t.Errorf("TotalIssueContributions = %d, want 3", got.TotalIssueContributions)
	}
	if got.ContributionCalendar.TotalContributions != 33 {
		t.Errorf("ContributionCalendar.TotalContributions = %d, want 33", got.ContributionCalendar.TotalContributions)
	}
	// Distinct repositories (o/a, o/b, o/c) exceed any single chunk's server total (2).
	if got.TotalRepositoriesWithContributedCommits != 3 {
		t.Errorf("TotalRepositoriesWithContributedCommits = %d, want 3", got.TotalRepositoriesWithContributedCommits)
	}
	if got.CommitContributionsByRepositoryTruncated {
		t.Error("CommitContributionsByRepositoryTruncated = true, want false")
	}
	want := []RepositoryContributions{
		{NameWithOwner: "o/a", Contributions: 3},
		{NameWithOwner: "o/b", Contributions: 10},
		{NameWithOwner: "o/c", Contributions: 7},
	}
	if !reflect.DeepEqual(got.CommitContributionsByRepository, want) {
		t.Errorf("CommitContributionsByRepository = %+v, want %+v", got.CommitContributionsByRepository, want)
	}
	wantDays := []ContributionDay{{Date: "2024-01-01", Count: 5}, {Date: "2025-01-01", Count: 9}}
	if !reflect.DeepEqual(got.ContributionCalendar.Days, wantDays) {
		t.Errorf("ContributionCalendar.Days = %+v, want %+v", got.ContributionCalendar.Days, wantDays)
	}
}

func TestMergeContributionChunks_TruncationPropagatesAcrossChunks(t *testing.T) {
	chunks := []*ContributionsCollection{
		{TotalRepositoriesWithContributedCommits: 1, CommitContributionsByRepository: []RepositoryContributions{{NameWithOwner: "o/a", Contributions: 1}}},
		{TotalRepositoriesWithContributedCommits: 120, CommitContributionsByRepositoryTruncated: true},
	}

	got := mergeContributionChunks(chunks)

	if !got.CommitContributionsByRepositoryTruncated {
		t.Error("CommitContributionsByRepositoryTruncated = false, want true")
	}
	// max(len(distinct)=1, maxServerTotal=120) == 120.
	if got.TotalRepositoriesWithContributedCommits != 120 {
		t.Errorf("TotalRepositoriesWithContributedCommits = %d, want 120", got.TotalRepositoriesWithContributedCommits)
	}
}
