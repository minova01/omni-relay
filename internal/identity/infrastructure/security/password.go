package security

import "golang.org/x/crypto/bcrypt"

type PasswordService struct {
	dummyHash []byte
}

func NewPasswordService() *PasswordService {
	dummyHash, err := bcrypt.GenerateFromPassword([]byte("omni-relay-dummy-password"), bcrypt.DefaultCost)
	if err != nil {
		panic("create bcrypt dummy hash: " + err.Error())
	}
	return &PasswordService{dummyHash: dummyHash}
}

func (s *PasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *PasswordService) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *PasswordService) DummyVerify(password string) {
	_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
}
