package main

import (
	"sort"
	"unicode/utf8"
)

const helpText = "usage: skillplumb check DIR [--max-desc N] [--max-lines N] [--ignore CODE]... [--strict]\n" +
	"       skillplumb index DIR\n" +
	"       skillplumb --help\n"

func Run(args []string, io Io) int {
	if io == nil {
		return 2
	}
	options, message := parseOptions(args)
	if message != "" {
		io.Err("skillplumb: " + message + "\n")
		return 2
	}
	if options.Command == "help" {
		io.Out(helpText)
		return 0
	}

	entries, listed := io.List(options.Dir)
	if !listed {
		io.Err("skillplumb: " + shownArgument(options.Dir) + ": not a directory\n")
		return 2
	}
	folders := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Kind != "dir" || entry.Name == "" || entry.Name[0] == '.' ||
			!utf8.ValidString(entry.Name) {
			continue
		}
		folders = append(folders, entry.Name)
	}
	sort.Strings(folders)
	if len(folders) == 0 {
		io.Err("skillplumb: " + shownArgument(options.Dir) + ": no skills found\n")
		return 2
	}

	skills := make([]SkillResult, 0, len(folders))
	for _, folder := range folders {
		skillPath := joinPath(options.Dir, folder)
		skills = append(skills, analyzeSkill(io, skillPath, folder, options.MaxDesc, options.MaxLines))
	}
	if options.Command == "index" {
		io.Out(renderIndex(skills))
		return 0
	}
	addDescriptionDuplicates(skills)
	output, status := renderCheck(skills, options.Ignore, options.Strict)
	io.Out(output)
	return status
}
