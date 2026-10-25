package main

import (
	"strconv"
	"strings"
)

// CommandOptions is the validated command contract consumed by Run.
type CommandOptions struct {
	Command  string
	Dir      string
	MaxDesc  int
	MaxLines int
	Strict   bool
	Ignore   map[string]bool
}

func parseOptions(args []string) (CommandOptions, string) {
	if len(args) == 0 {
		return CommandOptions{}, "missing command (check or index)"
	}

	command := args[0]
	if command == "--help" {
		if len(args) == 1 {
			return CommandOptions{Command: "help"}, ""
		}
		return CommandOptions{}, "unknown command " + shownArgument(command)
	}
	if command != "check" && command != "index" {
		return CommandOptions{}, "unknown command " + shownArgument(command)
	}

	options := CommandOptions{
		Command:  command,
		MaxDesc:  200,
		MaxLines: 80,
		Ignore:   make(map[string]bool),
	}
	var maxDescValue string
	var maxLinesValue string
	var ignoreValues []string
	maxDescSeen := false
	maxLinesSeen := false
	strictSeen := false
	dirSeen := false

	for index := 1; index < len(args); index++ {
		arg := args[index]
		if strings.HasPrefix(arg, "--") {
			if command == "index" {
				return CommandOptions{}, "unknown option " + shownArgument(arg)
			}
			switch arg {
			case "--max-desc":
				if maxDescSeen {
					return CommandOptions{}, "--max-desc given twice"
				}
				maxDescSeen = true
				if index+1 >= len(args) {
					return CommandOptions{}, "--max-desc needs a value"
				}
				index++
				maxDescValue = args[index]
			case "--max-lines":
				if maxLinesSeen {
					return CommandOptions{}, "--max-lines given twice"
				}
				maxLinesSeen = true
				if index+1 >= len(args) {
					return CommandOptions{}, "--max-lines needs a value"
				}
				index++
				maxLinesValue = args[index]
			case "--ignore":
				if index+1 >= len(args) {
					return CommandOptions{}, "--ignore needs a value"
				}
				index++
				ignoreValues = append(ignoreValues, args[index])
			case "--strict":
				if strictSeen {
					return CommandOptions{}, "--strict given twice"
				}
				strictSeen = true
				options.Strict = true
			default:
				return CommandOptions{}, "unknown option " + shownArgument(arg)
			}
			continue
		}

		if dirSeen {
			return CommandOptions{}, "unexpected argument " + shownArgument(arg)
		}
		dirSeen = true
		options.Dir = arg
	}
	if !dirSeen {
		return CommandOptions{}, "missing DIR"
	}

	if maxDescSeen {
		value, ok := parseLimit(maxDescValue)
		if !ok {
			return CommandOptions{}, "--max-desc needs a whole number of 1 or more, got " +
				shownArgument(maxDescValue)
		}
		options.MaxDesc = value
	}
	if maxLinesSeen {
		value, ok := parseLimit(maxLinesValue)
		if !ok {
			return CommandOptions{}, "--max-lines needs a whole number of 1 or more, got " +
				shownArgument(maxLinesValue)
		}
		options.MaxLines = value
	}
	for _, code := range ignoreValues {
		if !validCode(code) {
			return CommandOptions{}, "unknown code " + shownArgument(code)
		}
		options.Ignore[code] = true
	}
	return options, ""
}

func parseLimit(value string) (int, bool) {
	if len(value) > 6 || !isDigits(value) {
		return 0, false
	}
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 {
		return 0, false
	}
	return number, true
}

func validCode(code string) bool {
	if len(code) != 3 {
		return false
	}
	switch code {
	case "E01", "E02", "E03", "E04", "E05", "E06", "E07", "E08", "E09", "E10",
		"W01", "W02", "W03", "W04":
		return true
	default:
		return false
	}
}

func shownArgument(value string) string {
	if value == "" {
		return "(empty)"
	}
	return shown(value)
}
