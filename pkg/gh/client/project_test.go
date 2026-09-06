package client

import (
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

func newMultiSelectNode(fieldName string, options ...[2]string) multiSelectItemFieldValueNode {
	var n multiSelectItemFieldValueNode
	n.AsMultiSelect.Field.OnMultiSelect.Name = githubv4.String(fieldName)
	for _, o := range options {
		n.AsMultiSelect.Options = append(n.AsMultiSelect.Options, struct {
			ID   githubv4.String
			Name githubv4.String
		}{ID: githubv4.String(o[0]), Name: githubv4.String(o[1])})
	}
	return n
}

func TestMultiSelectItemFieldValueNode_ToFieldValue_EmptyOptions(t *testing.T) {
	n := newMultiSelectNode("Labels")
	fv, ok := n.toFieldValue()
	if !ok {
		t.Fatal("expected a usable value for a multi-select field with zero options")
	}
	if fv.ValueType != "MULTI_SELECT" {
		t.Errorf("expected ValueType MULTI_SELECT, got %q", fv.ValueType)
	}
	if fv.FieldName != "Labels" {
		t.Errorf("expected FieldName Labels, got %q", fv.FieldName)
	}
	if len(fv.SelectNames) != 0 || len(fv.SelectOptionIDs) != 0 {
		t.Errorf("expected empty selection, got names=%v ids=%v", fv.SelectNames, fv.SelectOptionIDs)
	}
}

func TestMultiSelectItemFieldValueNode_ToFieldValue_Populated(t *testing.T) {
	n := newMultiSelectNode("Labels", [2]string{"id1", "bug"}, [2]string{"id2", "urgent"})
	fv, ok := n.toFieldValue()
	if !ok {
		t.Fatal("expected a usable value for a populated multi-select field")
	}
	if fv.ValueType != "MULTI_SELECT" || fv.FieldName != "Labels" {
		t.Fatalf("unexpected value: %+v", fv)
	}
	if len(fv.SelectNames) != 2 || fv.SelectNames[0] != "bug" || fv.SelectNames[1] != "urgent" {
		t.Errorf("unexpected SelectNames: %v", fv.SelectNames)
	}
	if len(fv.SelectOptionIDs) != 2 || fv.SelectOptionIDs[0] != "id1" || fv.SelectOptionIDs[1] != "id2" {
		t.Errorf("unexpected SelectOptionIDs: %v", fv.SelectOptionIDs)
	}
}

func TestMultiSelectItemFieldValueNode_ToFieldValue_FallsThroughToText(t *testing.T) {
	var n multiSelectItemFieldValueNode
	// AsMultiSelect is zero-valued (not a multi-select node); a TEXT value is present instead.
	text := githubv4.String("hello")
	n.AsText.Text = &text
	n.AsText.Field.OnField.Name = githubv4.String("Notes")

	fv, ok := n.toFieldValue()
	if !ok {
		t.Fatal("expected the TEXT value to be returned")
	}
	if fv.ValueType != "TEXT" {
		t.Errorf("expected ValueType TEXT, got %q", fv.ValueType)
	}
	if fv.FieldName != "Notes" || fv.Text != "hello" {
		t.Errorf("unexpected value: %+v", fv)
	}
}

func TestProjectV2ItemNode_ToProjectV2Item_Issue(t *testing.T) {
	created := githubv4.DateTime{Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	updated := githubv4.DateTime{Time: time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)}
	closed := githubv4.DateTime{Time: time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)}

	var n projectV2ItemNode[multiSelectItemFieldValueNode]
	n.ID = "PVTI_1"
	n.Type = githubv4.String(ProjectV2ItemTypeIssue)
	issue := &n.Content.AsIssue
	issue.ID = "I_1"
	issue.Number = 42
	issue.Title = "Broken build"
	issue.URL = "https://github.com/octo/hello/issues/42"
	issue.State = "CLOSED"
	issue.CreatedAt = created
	issue.UpdatedAt = updated
	issue.ClosedAt = &closed
	issue.Author.Login = "octocat"
	issue.Repository.Name = "hello"
	issue.Repository.Owner.Login = "octo"
	issue.Assignees.Nodes = []struct{ Login githubv4.String }{{Login: "alice"}, {Login: "bob"}}
	issue.Labels.Nodes = []struct{ Name githubv4.String }{{Name: "bug"}}
	issue.Milestone = &milestoneTitleRef{Title: "v1.0"}

	c := n.toProjectV2Item().Content
	if c.Type != ProjectV2ItemTypeIssue || c.Number != 42 {
		t.Fatalf("unexpected content: %+v", c)
	}
	if c.State != "CLOSED" || c.IsOpen() {
		t.Errorf("expected a closed issue, got State=%q IsOpen=%v", c.State, c.IsOpen())
	}
	if c.CreatedAt != "2026-01-02T03:04:05Z" || c.UpdatedAt != "2026-02-03T04:05:06Z" || c.ClosedAt != "2026-03-04T05:06:07Z" {
		t.Errorf("unexpected timestamps: created=%q updated=%q closed=%q", c.CreatedAt, c.UpdatedAt, c.ClosedAt)
	}
	if len(c.Assignees) != 2 || c.Assignees[0] != "alice" || c.Assignees[1] != "bob" {
		t.Errorf("unexpected assignees: %v", c.Assignees)
	}
	if len(c.Labels) != 1 || c.Labels[0] != "bug" {
		t.Errorf("unexpected labels: %v", c.Labels)
	}
	if c.Milestone != "v1.0" {
		t.Errorf("unexpected milestone: %q", c.Milestone)
	}
}

func TestProjectV2ItemNode_ToProjectV2Item_OpenIssueWithoutOptionalFields(t *testing.T) {
	var n projectV2ItemNode[multiSelectItemFieldValueNode]
	n.Type = githubv4.String(ProjectV2ItemTypeIssue)
	n.Content.AsIssue.State = "OPEN"

	c := n.toProjectV2Item().Content
	if !c.IsOpen() {
		t.Error("expected the issue to be reported as open")
	}
	// A null closedAt, milestone, and empty connections must not produce placeholder values.
	if c.ClosedAt != "" || c.Milestone != "" {
		t.Errorf("expected empty optional fields, got closedAt=%q milestone=%q", c.ClosedAt, c.Milestone)
	}
	if c.Assignees != nil || c.Labels != nil {
		t.Errorf("expected nil slices, got assignees=%v labels=%v", c.Assignees, c.Labels)
	}
	// A zero-valued createdAt means the field was not returned, not the zero time.
	if c.CreatedAt != "" {
		t.Errorf("expected an empty createdAt, got %q", c.CreatedAt)
	}
}

func TestProjectV2ItemNode_ToProjectV2Item_MergedPullRequest(t *testing.T) {
	var n projectV2ItemNode[multiSelectItemFieldValueNode]
	n.Type = githubv4.String(ProjectV2ItemTypePullRequest)
	n.Content.AsPullRequest.State = "MERGED"
	n.Content.AsPullRequest.Number = 7

	c := n.toProjectV2Item().Content
	if c.Type != ProjectV2ItemTypePullRequest || c.Number != 7 {
		t.Fatalf("unexpected content: %+v", c)
	}
	if c.State != "MERGED" || c.IsOpen() {
		t.Errorf("expected a merged pull request, got State=%q IsOpen=%v", c.State, c.IsOpen())
	}
}

func TestProjectV2ItemNode_ToProjectV2Item_DraftIssue(t *testing.T) {
	var n projectV2ItemNode[multiSelectItemFieldValueNode]
	n.Type = githubv4.String(ProjectV2ItemTypeDraftIssue)
	n.Content.AsDraftIssue.Title = "Spike"
	n.Content.AsDraftIssue.CreatedAt = githubv4.DateTime{Time: time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)}

	c := n.toProjectV2Item().Content
	if c.Type != ProjectV2ItemTypeDraftIssue || c.Title != "Spike" {
		t.Fatalf("unexpected content: %+v", c)
	}
	if c.CreatedAt != "2026-05-06T07:08:09Z" {
		t.Errorf("unexpected createdAt: %q", c.CreatedAt)
	}
	// Draft issues have no state, so they are never reported as open.
	if c.State != "" || c.IsOpen() {
		t.Errorf("expected no state for a draft issue, got State=%q IsOpen=%v", c.State, c.IsOpen())
	}
}
