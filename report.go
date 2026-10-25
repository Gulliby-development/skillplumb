package main

import (
	"sort"
	"strconv"
	"strings"
)

func addDescriptionDuplicates(skills []SkillResult) {
	earliest := make(map[string]string)
	for index := range skills {
		skill := &skills[index]
		if skill.Description == "" {
			continue
		}
		key := foldKey(skill.Description)
		if folder, exists := earliest[key]; exists {
			skill.Findings = append(skill.Findings, Finding{
				Line:    skill.DescriptionLine,
				Code:    "W04",
				Message: "description is the same as in skill " + folder,
			})
			continue
		}
		earliest[key] = skill.Folder
	}
}

func renderCheck(skills []SkillResult, ignored map[string]bool, strict bool) (string, int) {
	var output strings.Builder
	errors := 0
	warnings := 0
	for _, skill := range skills {
		findings := make([]Finding, 0, len(skill.Findings))
		for _, finding := range skill.Findings {
			if !ignored[finding.Code] {
				findings = append(findings, finding)
			}
		}
		sort.SliceStable(findings, func(i, j int) bool {
			if findings[i].Line != findings[j].Line {
				return findings[i].Line < findings[j].Line
			}
			return findings[i].Code < findings[j].Code
		})
		for _, finding := range findings {
			if strings.HasPrefix(finding.Code, "E") {
				errors++
			} else if strings.HasPrefix(finding.Code, "W") {
				warnings++
			}
			output.WriteString(shown(skill.Folder))
			output.WriteString("/SKILL.md:")
			output.WriteString(strconv.Itoa(finding.Line))
			output.WriteString(": ")
			output.WriteString(shown(finding.Code))
			output.WriteByte(' ')
			output.WriteString(shown(finding.Message))
			output.WriteByte('\n')
		}
	}
	output.WriteString("skillplumb: ")
	output.WriteString(strconv.Itoa(len(skills)))
	output.WriteString(" skills, ")
	output.WriteString(strconv.Itoa(errors))
	output.WriteString(" errors, ")
	output.WriteString(strconv.Itoa(warnings))
	output.WriteString(" warnings\n")
	if errors > 0 || (strict && warnings > 0) {
		return output.String(), 1
	}
	return output.String(), 0
}

func renderIndex(skills []SkillResult) string {
	rows := make([][]string, 0, len(skills))
	for _, skill := range skills {
		if skill.Fatal {
			rows = append(rows, []string{skill.Folder, "-", "-", "-"})
			continue
		}
		description := "-"
		if skill.Description != "" {
			description = strconv.Itoa(skill.DescriptionPoints)
		}
		rows = append(rows, []string{
			skill.Folder,
			strconv.Itoa(skill.BodyLines),
			strconv.Itoa(skill.LinkCount),
			description,
		})
	}
	var output strings.Builder
	output.WriteString("skillplumb: ")
	output.WriteString(strconv.Itoa(len(skills)))
	output.WriteString(" skills\n")
	output.WriteString(table([]string{"SKILL", "LINES", "LINKS", "DESC"}, rows))
	return output.String()
}
