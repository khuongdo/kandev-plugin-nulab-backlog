package connection

import (
	"errors"
	"strings"
)

// ErrInvalidAPIKey is returned when the key input fails BR3.5. Its text never
// contains the input.
var ErrInvalidAPIKey = errors.New("API key must be 1-256 printable ASCII characters without spaces")

const maxAPIKeyLen = 256

// ValidateAPIKey trims the key and checks BR3.5. It returns the trimmed key.
func ValidateAPIKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if key == "" || len(key) > maxAPIKeyLen {
		return "", ErrInvalidAPIKey
	}
	for i := 0; i < len(key); i++ {
		if key[i] < 0x21 || key[i] > 0x7e {
			return "", ErrInvalidAPIKey
		}
	}
	return key, nil
}
