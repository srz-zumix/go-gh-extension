package ioutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// DecodeJSONFile reads JSON from a file path or stdin (when input is "-") and decodes it into a value of type T.
func DecodeJSONFile[T any](input string) (T, error) {
	var result T
	var r io.Reader
	source := input
	if input == "-" {
		source = "stdin"
		r = os.Stdin
	} else {
		f, err := os.Open(input)
		if err != nil {
			return result, fmt.Errorf("error opening input %q: %w", input, err)
		}
		defer f.Close() // nolint
		r = f
	}
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return result, fmt.Errorf("error parsing JSON input from %q: %w", source, err)
	}
	return result, nil
}
