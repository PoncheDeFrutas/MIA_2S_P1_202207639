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
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			errs = append(errs, "Line is empty")
			continue
		}

		var (
			result string
			err    error
		)
		switch strings.ToLower(tokens[0]) {
		case "mkdisk":
			result, err = commands.ParseMkDisk(tokens[1:])
		case "rmdisk":
			result, err = commands.ParseRmDisk(tokens[1:])
		default:
			err = fmt.Errorf("unknown command: %s", tokens[0])
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
