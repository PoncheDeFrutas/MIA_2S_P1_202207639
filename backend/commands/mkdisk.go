package commands

import (
	"backend/common"
	"backend/structures"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MkDisk struct {
	Size int
	Fit  string
	Unit string
	Path string
}

func ParserMkDisk(tokens []string) (string, error) {
	cmd := &MkDisk{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfFwW]{2}|-path="[^"]+"|-path=\S+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		key, value, err := common.ParseToken(match)
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size < 1 {
				return "", fmt.Errorf("invalid size: %s", value)
			}
			cmd.Size = size
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", fmt.Errorf("invalid fit: %s", value)
			}
			cmd.Fit = value
		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" {
				return "", fmt.Errorf("invalid unit: %s", value)
			}
			cmd.Unit = value
		case "-path":
			if value == "" {
				return "", fmt.Errorf("invalid path: %s", value)
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

	if err := commandMkDisk(cmd); err != nil {
		return "", err
	}

	return "", nil
}

func commandMkDisk(cmd *MkDisk) error {
	sizeInBytes, err := common.ConvertToBytes(cmd.Size, cmd.Unit)
	if err != nil {
		return err
	}

	if err := createDisk(cmd, sizeInBytes); err != nil {
		return err
	}

	mbr := &structures.MBR{}

	if err := mbr.CreateMBR(sizeInBytes, cmd.Fit, cmd.Path); err != nil {
		return err
	}

	mbr.Print()
	return nil
}

func createDisk(cmd *MkDisk, sizeInBytes int) error {
	if err := os.MkdirAll(filepath.Dir(cmd.Path), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(cmd.Path)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	if err := writeToFile(file, sizeInBytes); err != nil {
		return err
	}

	return nil
}

func writeToFile(file *os.File, sizeInBytes int) error {
	buffer := make([]byte, 1024*1024)

	for sizeInBytes > 0 {
		writeSize := len(buffer)

		if sizeInBytes < len(buffer) {
			writeSize = sizeInBytes
		}

		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err
		}

		sizeInBytes -= writeSize
	}

	return nil
}
