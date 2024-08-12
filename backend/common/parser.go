package common

import (
	"fmt"
	"strings"
)

func ParseToken(token string) (string, string, error) {
	parts := strings.SplitN(token, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid token format: %s", token)
	}
	return strings.ToLower(parts[0]), parts[1], nil
}
