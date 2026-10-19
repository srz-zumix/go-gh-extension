package ioutil

import "testing"

func TestJSONStringsReportsLines(t *testing.T) {
	got, err := JSONStrings("{\n  \"a\": \"x\",\n  \"b\": [\"y\", \"\"]\n}\n", 10)
	if err != nil {
		t.Fatalf("JSONStrings() error = %v", err)
	}
	want := []JSONString{{"a", 11}, {"x", 11}, {"b", 12}, {"y", 12}}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestJSONStringsDecodesEscapes(t *testing.T) {
	got, err := JSONStrings(`{"k":"line1\nline2"}`, 1)
	if err != nil {
		t.Fatalf("JSONStrings() error = %v", err)
	}
	if len(got) != 2 || got[1].Value != "line1\nline2" {
		t.Fatalf("got %v, want the decoded value", got)
	}
}

func TestJSONStringsRejectsInvalidJSON(t *testing.T) {
	if _, err := JSONStrings(`{"a": `, 1); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestJSONStringsIgnoredKeys(t *testing.T) {
	text := "{\n\"omit\":{\"nested\":[\"hidden\"]},\n\"array\":[{\"omit\":\"hidden\",\"keep\":\"visible\"},\"omit\"],\n\"omit\":null,\"last\":\"end\"\n}\n\"tail\""
	got, err := JSONStrings(text, 10, "omit")
	if err != nil {
		t.Fatal(err)
	}
	want := []JSONString{{"omit", 11}, {"array", 12}, {"omit", 12}, {"keep", 12}, {"visible", 12}, {"omit", 12}, {"omit", 13}, {"last", 13}, {"end", 13}, {"tail", 15}}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("got[%d] = %v, want %v", index, got[index], want[index])
		}
	}
}
