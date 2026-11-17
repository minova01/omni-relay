package application

import "context"

type Credential struct {
	Plaintext string
	Prefix    string
	Hash      string
	Hint      string
}

type CredentialService interface {
	Generate() (Credential, error)
	Parse(string) (prefix string, secret string, err error)
	Verify(storedHash, secret string) bool
}

type BalanceStore interface {
	GetOrInitialize(context.Context, int64, int64) (int64, error)
}
