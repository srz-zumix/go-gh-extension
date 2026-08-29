package gh

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/google/go-github/v90/github"
	"golang.org/x/crypto/nacl/box"
)

func TestEncryptSecret(t *testing.T) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	encodedKey := base64.StdEncoding.EncodeToString(publicKey[:])
	keyID := "1234567890"
	eSecret, err := EncryptSecret(&github.PublicKey{Key: &encodedKey, KeyID: &keyID}, "MY_SECRET", "s3cret-value")
	if err != nil {
		t.Fatalf("EncryptSecret returned error: %v", err)
	}
	if eSecret.Name != "MY_SECRET" {
		t.Errorf("expected name MY_SECRET, got %q", eSecret.Name)
	}
	if eSecret.KeyID != keyID {
		t.Errorf("expected key ID %q, got %q", keyID, eSecret.KeyID)
	}

	sealed, err := base64.StdEncoding.DecodeString(eSecret.EncryptedValue)
	if err != nil {
		t.Fatalf("failed to decode encrypted value: %v", err)
	}
	decrypted, ok := box.OpenAnonymous(nil, sealed, publicKey, privateKey)
	if !ok {
		t.Fatal("failed to decrypt the encrypted value")
	}
	if string(decrypted) != "s3cret-value" {
		t.Errorf("expected s3cret-value, got %q", string(decrypted))
	}
}

func TestEncryptSecretInvalidPublicKey(t *testing.T) {
	shortKey := base64.StdEncoding.EncodeToString([]byte("too-short"))
	if _, err := EncryptSecret(&github.PublicKey{Key: &shortKey}, "MY_SECRET", "value"); err == nil {
		t.Error("expected an error for a public key of the wrong length")
	}

	notBase64 := "!!!not-base64!!!"
	if _, err := EncryptSecret(&github.PublicKey{Key: &notBase64}, "MY_SECRET", "value"); err == nil {
		t.Error("expected an error for a public key that is not base64")
	}
}

func TestEncryptSecretNilPublicKey(t *testing.T) {
	if _, err := EncryptSecret(nil, "MY_SECRET", "value"); err == nil {
		t.Error("expected an error for a nil public key")
	}
}

func TestEncryptSecretMissingKeyID(t *testing.T) {
	publicKey, _, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(publicKey[:])
	if _, err := EncryptSecret(&github.PublicKey{Key: &encodedKey}, "MY_SECRET", "value"); err == nil {
		t.Error("expected an error for a public key with an empty key ID")
	}
}
