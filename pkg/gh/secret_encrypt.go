package gh

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
	"golang.org/x/crypto/nacl/box"
)

// publicKeySize is the byte length of an Actions secret public key.
const publicKeySize = 32

// EncryptSecret seals value with publicKey so the result can be passed to the
// CreateOrUpdate*Secret functions.
func EncryptSecret(publicKey *github.PublicKey, name, value string) (*github.EncryptedSecret, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("public key is nil")
	}
	if publicKey.GetKeyID() == "" {
		return nil, fmt.Errorf("public key ID is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	if len(decoded) != publicKeySize {
		return nil, fmt.Errorf("unexpected public key length: got %d bytes, want %d", len(decoded), publicKeySize)
	}
	var peersPublicKey [publicKeySize]byte
	copy(peersPublicKey[:], decoded)

	encrypted, err := box.SealAnonymous(nil, []byte(value), &peersPublicKey, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
	}

	return &github.EncryptedSecret{
		Name:           name,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: base64.StdEncoding.EncodeToString(encrypted),
	}, nil
}

// SetRepoSecret encrypts a plaintext value with the repository public key and
// stores it as a repository secret.
func SetRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name, value string) error {
	publicKey, err := GetRepoPublicKey(ctx, g, repo)
	if err != nil {
		return err
	}
	eSecret, err := EncryptSecret(publicKey, name, value)
	if err != nil {
		return err
	}
	return CreateOrUpdateRepoSecret(ctx, g, repo, eSecret)
}
