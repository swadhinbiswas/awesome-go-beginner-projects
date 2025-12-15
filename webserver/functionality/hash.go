package functionality

import (
	"golang.org/x/crypto/bcrypt"
	"fmt"

)

func PassHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func PassVerify(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	return nil
}


func VerifyUserPassword(u *User, password string) error {
    return PassVerify(u.Password, password)
}
