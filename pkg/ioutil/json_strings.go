package ioutil

import (
	"encoding/json"
	"io"
	"slices"
	"strings"
)

// JSONString is a decoded JSON string token (object key or value) and the line
// it ends on.
type JSONString struct {
	Value string
	Line  int
}

// JSONStrings decodes every non-empty string token in text, which may hold
// several concatenated JSON values. firstLine is the line number of text's
// first line. Values under ignoredKeys are omitted, including nested values.
// It returns an error when text is not valid JSON.
func JSONStrings(text string, firstLine int, ignoredKeys ...string) ([]JSONString, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	var out []JSONString
	line := firstLine
	var prev int64
	type container struct {
		object    bool
		expectKey bool
		skip      bool
		skipValue bool
	}
	var stack []container
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			// Token reports a truncated document as a clean EOF.
			if len(stack) != 0 {
				return nil, io.ErrUnexpectedEOF
			}
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		off := dec.InputOffset()
		line += strings.Count(text[prev:off], "\n")
		prev = off
		skip := false
		if len(stack) > 0 {
			parent := &stack[len(stack)-1]
			skip = parent.skip || parent.skipValue
			if key, ok := tok.(string); ok && parent.object && parent.expectKey {
				parent.expectKey = false
				parent.skipValue = slices.Contains(ignoredKeys, key)
				if key != "" && !parent.skip {
					out = append(out, JSONString{Value: key, Line: line})
				}
				continue
			}
			if parent.object {
				parent.expectKey = true
				parent.skipValue = false
			}
		}
		switch v := tok.(type) {
		case json.Delim:
			if v == '{' || v == '[' {
				stack = append(stack, container{object: v == '{', expectKey: v == '{', skip: skip})
			} else {
				stack = stack[:len(stack)-1]
			}
		case string:
			if v != "" && !skip {
				out = append(out, JSONString{Value: v, Line: line})
			}
		}
	}
}
