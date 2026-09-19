package gh

import (
	"reflect"
	"testing"
	"time"
)

func TestContributionWindows_SingleWindowWhenWithinLimit(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := since.Add(30 * 24 * time.Hour)

	windows := contributionWindows(since, until)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}
	if !windows[0].from.Equal(since) || !windows[0].to.Equal(until) {
		t.Errorf("window = [%s, %s], want [%s, %s]", windows[0].from, windows[0].to, since, until)
	}
}

func TestContributionWindows_ExactlyOneWindowAtLimit(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := since.Add(maxContributionsWindow)

	windows := contributionWindows(since, until)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}
	if !windows[0].to.Equal(until) {
		t.Errorf("window.to = %s, want %s", windows[0].to, until)
	}
}

func TestContributionWindows_DisjointAndContiguousAcrossChunks(t *testing.T) {
	since := time.Date(2020, 3, 15, 8, 30, 0, 0, time.UTC)
	until := since.Add(2*maxContributionsWindow + 100*24*time.Hour)

	windows := contributionWindows(since, until)

	if len(windows) != 3 {
		t.Fatalf("len(windows) = %d, want 3", len(windows))
	}
	if !windows[0].from.Equal(since) {
		t.Errorf("first from = %s, want %s", windows[0].from, since)
	}
	if !windows[len(windows)-1].to.Equal(until) {
		t.Errorf("last to = %s, want %s", windows[len(windows)-1].to, until)
	}
	for i := 0; i+1 < len(windows); i++ {
		// Non-final windows must not overlap the next one, and the next window must
		// begin exactly one second after this window ends (no gap, no shared instant).
		if !windows[i].to.Before(windows[i+1].from) {
			t.Errorf("window %d to (%s) not before window %d from (%s)", i, windows[i].to, i+1, windows[i+1].from)
		}
		if !windows[i].to.Equal(windows[i+1].from.Add(-time.Second)) {
			t.Errorf("window %d to = %s, want one second before %s", i, windows[i].to, windows[i+1].from)
		}
		if windows[i].to.Sub(windows[i].from) >= maxContributionsWindow {
			t.Errorf("window %d spans %s, want < %s", i, windows[i].to.Sub(windows[i].from), maxContributionsWindow)
		}
	}
}

func TestContributionWindows_NormalizesSubSecondEndpoints(t *testing.T) {
	// Sub-second endpoints must be normalized to whole seconds so the one-second
	// handoff stays exact and no contribution can fall between two windows.
	since := time.Date(2020, 3, 15, 8, 30, 0, 500_000_000, time.UTC)
	until := since.Add(2 * maxContributionsWindow).Add(400_000_000 * time.Nanosecond)

	windows := contributionWindows(since, until)

	if len(windows) < 2 {
		t.Fatalf("len(windows) = %d, want >= 2", len(windows))
	}
	for i, w := range windows {
		if w.from.Nanosecond() != 0 || w.to.Nanosecond() != 0 {
			t.Errorf("window %d has sub-second component: from=%s to=%s", i, w.from, w.to)
		}
	}
	for i := 0; i+1 < len(windows); i++ {
		if !windows[i].to.Equal(windows[i+1].from.Add(-time.Second)) {
			t.Errorf("window %d to = %s, want one second before %s", i, windows[i].to, windows[i+1].from)
		}
	}
	if !windows[0].from.Equal(since.Truncate(time.Second)) {
		t.Errorf("first from = %s, want %s", windows[0].from, since.Truncate(time.Second))
	}
	if !windows[len(windows)-1].to.Equal(until.Truncate(time.Second)) {
		t.Errorf("last to = %s, want %s", windows[len(windows)-1].to, until.Truncate(time.Second))
	}
}

func TestContributionWindows_EmptyWhenUntilNotAfterSince(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if w := contributionWindows(ts, ts); len(w) != 0 {
		t.Errorf("equal endpoints: len(windows) = %d, want 0", len(w))
	}
	if w := contributionWindows(ts.Add(time.Hour), ts); len(w) != 0 {
		t.Errorf("since after until: len(windows) = %d, want 0", len(w))
	}
}

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
				Days:               []ContributionDay{{Date: "2024-01-01", Count: 5}, {Date: "2024-01-02", Count: 6}},
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
				Days:               []ContributionDay{{Date: "2025-01-01", Count: 9}, {Date: "2025-01-02", Count: 13}},
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
	wantDays := []ContributionDay{
		{Date: "2024-01-01", Count: 5},
		{Date: "2024-01-02", Count: 6},
		{Date: "2025-01-01", Count: 9},
		{Date: "2025-01-02", Count: 13},
	}
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

func TestMergeContributionChunks_DeduplicatesBoundaryCalendarDay(t *testing.T) {
	// The contribution calendar is day-granular in the user's timezone, so a day that
	// straddles a chunk boundary can be reported in full by both adjacent chunks. The
	// merge must keep a single entry per date (not append duplicates) and recompute the
	// calendar total from the unique days.
	chunks := []*ContributionsCollection{
		{
			ContributionCalendar: ContributionCalendar{
				TotalContributions: 11,
				Days:               []ContributionDay{{Date: "2024-06-03", Count: 4}, {Date: "2024-06-04", Count: 7}},
			},
		},
		{
			ContributionCalendar: ContributionCalendar{
				TotalContributions: 12,
				Days:               []ContributionDay{{Date: "2024-06-04", Count: 7}, {Date: "2024-06-05", Count: 5}},
			},
		},
	}

	got := mergeContributionChunks(chunks)

	wantDays := []ContributionDay{
		{Date: "2024-06-03", Count: 4},
		{Date: "2024-06-04", Count: 7},
		{Date: "2024-06-05", Count: 5},
	}
	if !reflect.DeepEqual(got.ContributionCalendar.Days, wantDays) {
		t.Errorf("ContributionCalendar.Days = %+v, want %+v", got.ContributionCalendar.Days, wantDays)
	}
	// The shared 2024-06-04 (7) is counted once: 4 + 7 + 5 = 16, not 11 + 12 = 23.
	if got.ContributionCalendar.TotalContributions != 16 {
		t.Errorf("ContributionCalendar.TotalContributions = %d, want 16", got.ContributionCalendar.TotalContributions)
	}
}
