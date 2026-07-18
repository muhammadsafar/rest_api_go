package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password string, reqPassword string) error {
	parts := strings.Split(password, ",")
	if len(parts) != 2 {
		return ErrorHandler(errors.New("Failed to decode the salt"), "Failed to decode the salt")
	}

	saltBase64 := parts[0]
	hashPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {

		// http.Error(w, "Failed to decode the salt", http.StatusInternalServerError)
		return ErrorHandler(err, "Failed to decode the salt")
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashPasswordBase64)
	if err != nil {
		// http.Error(w, "Failed to decode the hash password", http.StatusInternalServerError)
		return ErrorHandler(err, "Failed to decode the hash password")
	}

	hash := argon2.IDKey([]byte(reqPassword), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(hashedPassword) {
		// http.Error(w, "incorrect password", http.StatusInternalServerError)
		return ErrorHandler(errors.New("Incorrent password"), "incorrect password")
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		//ok
		return nil
	}
	return ErrorHandler(errors.New("Incorrent password"), "incorrect password")
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrorHandler(errors.New("Please fill the blank"), "Please enter password")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("failed to generate salt"), "internal error")
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	encodeHash := fmt.Sprintf("%s,%s", saltBase64, hashBase64)
	password = encodeHash
	return encodeHash, nil
}
