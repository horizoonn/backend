package sharedrecap

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

const sharedRecapTokenBytes = 16

func GenerateToken() (entity.SharedRecapToken, error) {
	randomBytes := make([]byte, sharedRecapTokenBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return entity.SharedRecapToken(
		base64.RawURLEncoding.EncodeToString(randomBytes),
	), nil
}

func generateValidToken(generator TokenGenerator) (entity.SharedRecapToken, error) {
	token, err := generator()
	if err != nil {
		return "", err
	}
	if tokenErr := validateToken(token); tokenErr != nil {
		return "", tokenErr
	}

	return token, nil
}

func validateToken(token entity.SharedRecapToken) error {
	value := string(token)
	if len(value) != entity.SharedRecapTokenLength {
		return fmt.Errorf("token must contain %d characters", entity.SharedRecapTokenLength)
	}
	for _, char := range value {
		if isURLSafeBase64Character(char) {
			continue
		}

		return fmt.Errorf("token contains invalid character %q", char)
	}

	return nil
}

func isURLSafeBase64Character(char rune) bool {
	switch {
	case char >= 'a' && char <= 'z':
		return true
	case char >= 'A' && char <= 'Z':
		return true
	case char >= '0' && char <= '9':
		return true
	case char == '_', char == '-':
		return true
	default:
		return false
	}
}
