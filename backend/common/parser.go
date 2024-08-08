package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

func ParseToken(token string) (string, string, error) {
	parts := strings.SplitN(token, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid token format: %s", token)
	}
	return strings.ToLower(parts[0]), parts[1], nil
}

func HandleQuote(value string, tokens []string, i int) (string, int, error) {
	if strings.HasPrefix(value, "\"") {
		for !strings.HasSuffix(value, "\"") && i+1 < len(tokens) {
			i++
			value += " " + tokens[i]
		}
		if !strings.HasSuffix(value, "\"") {
			return "", i, errors.New("missing closing quote")
		} else {
			value = strings.Trim(value, "\"")
		}
	}
	return value, i, nil
}

func WriteToFile(path string, offset int64, data interface{}) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	if _, err = file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek file: %v", err)
	}

	if err = binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}

func ReadFromFile(path string, offset int64, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	if _, err = file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek file: %v", err)
	}

	if err = binary.Read(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("failed to read from file: %v", err)
	}

	return nil
}

//TODO configurar metodos para encontrar espacios libres
