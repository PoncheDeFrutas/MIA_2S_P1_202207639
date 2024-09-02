package commands

import (
	"backend/utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MkFile struct {
	Path string
	R    bool
	Size int
	Cont string
}

func ParserMkFile(tokens []string) (string, error) {
	cmd := &MkFile{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=\S+|-r|-size=\d+|-cont="[^"]+"|-cont=\S+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		key, value, err := utils.ParseToken(match)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-path":
			if value == "" {
				return "", fmt.Errorf("invalid path: %s", value)
			}
			cmd.Path = value
		case "-r":
			cmd.R = true
		case "-size":
			num, err := strconv.Atoi(value)
			if err != nil || num < 0 {
				return "", fmt.Errorf("invalid size: %d", num)
			}
			cmd.Size = num
		case "-cont":
			if value == "" {
				return "", fmt.Errorf("invalid content: %s", value)
			}
			cmd.Cont = value
		}
	}

	if cmd.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	if err := cmd.commandMkFile(); err != nil {
		return "", err
	}

	return "", nil
}

func (cmd *MkFile) commandMkFile() error {
	return nil
}
