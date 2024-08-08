package common

import (
	"errors"
	"fmt"
)

// ConvertToBytes converts a size to bytes
func ConvertToBytes(size int, unit string) (int, error) {
	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	case "G":
		return size * 1024 * 1024 * 1024, nil
	default:
		return 0, errors.New(fmt.Sprintf("Unknown unit: %s", unit))
	}
}
