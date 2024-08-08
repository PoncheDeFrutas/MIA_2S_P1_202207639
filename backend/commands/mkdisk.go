package commands

import (
	"backend/common"
	"backend/structures"
	"fmt"
	"strconv"
)

type MkDisk struct {
	Size int
	Fit  string
	Unit string
	Path string
}

func ParseMkDisk(tokens []string) (string, error) {
	cmd := &MkDisk{}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		key, value, err := common.ParseToken(token)
		if err != nil {
			return "", err
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size < 1 {
				return "", fmt.Errorf("invalid size: %s", value)
			}
			cmd.Size = size
		case "-fit":
			if value != "BF" && value != "FF" && value != "WF" {
				return "", fmt.Errorf("invalid fit: %s", value)
			}
			cmd.Fit = value
		case "-unit":
			if value != "K" && value != "M" {
				return "", fmt.Errorf("invalid unit: %s", value)
			}
			cmd.Unit = value
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

	if cmd.Size == 0 {
		return "", fmt.Errorf("missing size")
	}
	if cmd.Fit == "" {
		cmd.Fit = "FF"
	}
	if cmd.Unit == "" {
		cmd.Unit = "M"
	}
	if cmd.Path == "" {
		return "", fmt.Errorf("missing path")
	}

	sizeInBytes, err := common.ConvertToBytes(cmd.Size, cmd.Unit)

	err = structures.CreateBinaryFile(sizeInBytes, cmd.Path)
	if err != nil {
		return "", err
	}

	mbr := &structures.MBR{}
	err = mbr.CreateMBR(sizeInBytes, cmd.Fit, cmd.Path)
	if err != nil {
		return "", err
	}

	err = common.WriteToFile(cmd.Path, 0, mbr)
	if err != nil {
		return "", err
	}

	err = common.ReadFromFile(cmd.Path, 0, mbr)
	fmt.Printf(mbr.String())
	return cmd.String(), nil
}

func (cmd *MkDisk) String() string {
	return fmt.Sprintf(
		"Disk created with size: %d, fit: %s, unit: %s, path: %s",
		cmd.Size, cmd.Fit, cmd.Unit, cmd.Path,
	)
}
