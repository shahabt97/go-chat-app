package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) interface{} {

	passBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil
	}

	return string(passBytes)
	// bcrypt.CompareHashAndPassword()
}

func CompareHash(hashedPass, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))
	return err
}
