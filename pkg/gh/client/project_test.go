package client

import (
	"testing"

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
