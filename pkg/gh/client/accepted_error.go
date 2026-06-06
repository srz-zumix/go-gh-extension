package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-github/v84/github"
)

// handleAcceptedError normalizes github.AcceptedError (HTTP 202) as a non-fatal success.
// If out is non-nil and the AcceptedError carries a response body, the body is
// unmarshalled into out. Returns nil on success, or an error if unmarshalling fails
// or if err is not an AcceptedError.
func handleAcceptedError(err error, out any) error {
	if err == nil {
		return nil
	}

	var aerr *github.AcceptedError
	if !errors.As(err, &aerr) {
		return err
	}

	if out != nil && len(aerr.Raw) > 0 {
		if jsonErr := json.Unmarshal(aerr.Raw, out); jsonErr != nil {
			return fmt.Errorf("accepted (http 202) but failed to parse response body: %w", jsonErr)
		}
	}
	}

	return nil
}
