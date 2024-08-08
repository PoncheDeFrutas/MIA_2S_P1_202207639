package commands

import (
	"backend/common"
	"backend/structures"
	"fmt"
)

type RmDisk struct {
	Path string
}

func ParseRmDisk(tokens []string) (string, error) {
	cmd := &RmDisk{}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		key, value, err := common.ParseToken(token)
		if err != nil {
			return "", err
		}

		switch key {
		case "-path":
			value, i, err = common.HandleQuote(value, tokens, i)
			if err != nil {
				return "", err
			}
			cmd.Path = value
		default:
			return "", fmt.Errorf("unknown parameter: %s", key)
		}
	}

	if cmd.Path == "" {
		return "", fmt.Errorf("missing path")
	}

	err := structures.DeleteBinaryFile(cmd.Path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("RmDisk: %s", cmd.Path), nil
}
