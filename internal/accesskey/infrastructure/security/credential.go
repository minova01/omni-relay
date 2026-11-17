package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"omni-relay/internal/accesskey/application"
	"omni-relay/internal/accesskey/domain"
)

type CredentialService struct{}

func NewCredentialService() *CredentialService {
	return &CredentialService{}
}

func (s *CredentialService) Generate() (application.Credential, error) {
	prefixBytes := make([]byte, 8)
	if _, err := rand.Read(prefixBytes); err != nil {
		return application.Credential{}, fmt.Errorf("generate api key prefix: %w", err)
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return application.Credential{}, fmt.Errorf("generate api key secret: %w", err)
	}

	prefix := hex.EncodeToString(prefixBytes)
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)
	hash := sha256.Sum256([]byte(secret))
	return application.Credential{
		Plaintext: "or_" + prefix + "_" + secret,
		Prefix:    prefix,
		Hash:      hex.EncodeToString(hash[:]),
		Hint:      secret[len(secret)-4:],
	}, nil
}

func (s *CredentialService) Parse(plaintext string) (string, string, error) {
	parts := strings.SplitN(plaintext, "_", 3)
	if len(parts) != 3 || parts[0] != "or" || len(parts[1]) != 16 || parts[2] == "" {
		return "", "", domain.ErrInvalidAPIKey
	}
	if _, err := hex.DecodeString(parts[1]); err != nil {
		return "", "", domain.ErrInvalidAPIKey
	}
	return parts[1], parts[2], nil
}

func (s *CredentialService) Verify(storedHash, secret string) bool {
	hash := sha256.Sum256([]byte(secret))
	calculated := hex.EncodeToString(hash[:])
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(calculated)) == 1
}
