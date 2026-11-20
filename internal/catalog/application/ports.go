package application

type SecretCipher interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}
