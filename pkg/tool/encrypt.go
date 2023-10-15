package tool

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type IEncryptor interface {
	Hash(token string) (string, error)
	Validate(token, hashedToken string) bool
}

type Encryptor struct{}

func (gen *Encryptor) Hash(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing token: %w", err)
	}
	return string(hash), nil
}

func (gen *Encryptor) Validate(token, hashedToken string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(token))
	return err == nil
}
