package commands

import (
	"backend/global"
	"backend/structures"
	"backend/utils"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
)

type MkDIR struct {
	Path string
	P    bool
}

func ParserMkDIR(tokens []string) (string, error) {
	cmd := &MkDIR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=\S+|-p`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		var key, value string
		var err error

		if match == "-p" {
			key = "-p"
			value = ""
		} else {
			key, value, err = utils.ParseToken(match)
			if err != nil {
				return "", err
			}

			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
		}

		switch key {
		case "-path":
			if value == "" {
				return "", fmt.Errorf("invalid path: %s", value)
			}
			cmd.Path = value
		case "-p":
			cmd.P = true
		}
	}

	if cmd.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	if err := cmd.commandMkDIR(); err != nil {
		return "", err
	}

	return "", nil
}

func (cmd *MkDIR) commandMkDIR() error {
	_, _, err := global.GetLoggedUser()
	if err != nil {
		return fmt.Errorf("you must be logged in")
	}

	mountedPartition, partitionPath, err := global.GetMountedPartition(global.LoggedPartition)
	if err != nil {
		return err
	}

	sb := &structures.SuperBlock{}
	if err := sb.ReadSuperBlock(partitionPath, int64(mountedPartition.PartStart)); err != nil {
		return err
	}

	array := strings.Split(cmd.Path, "/")
	var result []string
	for _, part := range array {
		if part != "" {
			result = append(result, part)
		}
	}

	if err := sb.CreateInode(partitionPath, 0, result, cmd.P, false); err != nil {
		return err
	}

	if err := sb.WriteSuperBlock(partitionPath, int64(mountedPartition.PartStart), int64(mountedPartition.PartStart+int32(binary.Size(sb)))); err != nil {
		return err
	}

	return nil
}
