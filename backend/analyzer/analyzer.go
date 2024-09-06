package analyzer

import (
	"backend/commands"
	"fmt"
	"strings"
)

func Analyzer(input string) string {
	var (
		results []string
		errs    []string
	)

	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			errs = append(errs, "empty line")
			continue
		}

		var (
			result string
			err    error
		)

		switch strings.ToLower(tokens[0]) {
		case "mkdisk":
			result, err = commands.ParserMkDisk(tokens[1:])
		case "rmdisk":
			result, err = commands.ParserRmDisk(tokens[1:])
		case "fdisk":
			result, err = commands.ParserFDisk(tokens[1:])
		case "mount":
			result, err = commands.ParserMount(tokens[1:])
		case "mkfs":
			result, err = commands.ParserMkFs(tokens[1:])
		case "rep":
			result, err = commands.ParserREP(tokens[1:])
		case "login":
			result, err = commands.ParserLogin(tokens[1:])
		case "logout":
			result, err = commands.ParserLogout(tokens[1:])
		case "mkgrp":
			result, err = commands.ParserMkGRP(tokens[1:])
		case "rmgrp":
			result, err = commands.ParserRmGRP(tokens[1:])
		case "mkusr":
			result, err = commands.ParserMkUSR(tokens[1:])
		case "rmusr":
			result, err = commands.ParserRmUSR(tokens[1:])
		case "chgrp":
			result, err = commands.ParserChGRP(tokens[1:])
		case "mkdir":
			result, err = commands.ParserMkDIR(tokens[1:])
		case "mkfile":
			result, err = commands.ParserMkFile(tokens[1:])
		case "cat":
			result, err = commands.ParserCat(tokens[1:])
		default:
			err = fmt.Errorf("command not found: %s", tokens[0])
		}

		if err != nil {
			errs = append(errs, err.Error())
		} else {
			results = append(results, result)
		}
	}

	var output strings.Builder
	if len(results) > 0 {
		output.WriteString("Results:\n")
		output.WriteString(strings.Join(results, "\n"))
		output.WriteString("\n")
	}

	if len(errs) > 0 {
		output.WriteString("Errors:\n")
		output.WriteString(strings.Join(errs, "\n"))
		output.WriteString("\n")
	}

	return output.String()
}
