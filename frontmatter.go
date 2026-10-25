package main

import (
	"strconv"
	"strings"
)

// Finding is the common, source-located diagnostic contract used by skill
// analysis and report rendering.
type Finding struct {
	Line    int
	Code    string
	Message string
}

// FrontMatterResult carries parsed first values and validated metadata. Fatal
// marks a missing or unclosed front matter block, which ends skill analysis.
type FrontMatterResult struct {
	Values            map[string]string
	KeyLines          map[string]int
	ClosingLine       int
	Findings          []Finding
	Fatal             bool
	Name              string
	NameLine          int
	Description       string
	DescriptionLine   int
	DescriptionPoints int
}

func analyzeFrontMatter(lines []string, folder string, maxDesc int) FrontMatterResult {
	result := FrontMatterResult{
		Values:   make(map[string]string),
		KeyLines: make(map[string]int),
	}
	if len(lines) == 0 || lines[0] != "---" {
		result.Fatal = true
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E03", Message: "front matter is missing or not closed",
		})
		return result
	}

	closeIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closeIndex = i
			break
		}
	}
	if closeIndex < 0 {
		result.Fatal = true
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E03", Message: "front matter is missing or not closed",
		})
		return result
	}
	result.ClosingLine = closeIndex + 1

	for i := 1; i < closeIndex; i++ {
		line := lines[i]
		if isBlank(line) || strings.HasPrefix(line, "#") {
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon < 0 || !validFrontMatterKey(line[:colon]) {
			result.Findings = append(result.Findings, Finding{
				Line: i + 1, Code: "E04", Message: "line is not key: value",
			})
			continue
		}

		key := line[:colon]
		if _, exists := result.Values[key]; exists {
			result.Findings = append(result.Findings, Finding{
				Line: i + 1, Code: "E05", Message: "key " + key + " given twice",
			})
			continue
		}
		result.Values[key] = trimBlanks(line[colon+1:])
		result.KeyLines[key] = i + 1
		if key != "name" && key != "description" && key != "license" {
			result.Findings = append(result.Findings, Finding{
				Line: i + 1, Code: "W01", Message: "unknown key " + key,
			})
		}
	}

	if line, ok := result.KeyLines["name"]; ok {
		result.Name = result.Values["name"]
		result.NameLine = line
	}
	if line, ok := result.KeyLines["description"]; ok {
		result.Description = result.Values["description"]
		result.DescriptionLine = line
	}

	if result.NameLine == 0 {
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E06", Message: "name is missing",
		})
	} else if result.Name == "" {
		result.Findings = append(result.Findings, Finding{
			Line: result.NameLine, Code: "E06", Message: "name is empty",
		})
	} else if !validSkillName(result.Name) {
		result.Findings = append(result.Findings, Finding{
			Line: result.NameLine, Code: "E06", Message: "name " + result.Name + " is not valid",
		})
	} else if result.Name != folder {
		result.Findings = append(result.Findings, Finding{
			Line:    result.NameLine,
			Code:    "E06",
			Message: "name " + result.Name + " does not match folder " + folder,
		})
	}

	if result.DescriptionLine == 0 {
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E07", Message: "description is missing",
		})
	} else if result.Description == "" {
		result.Findings = append(result.Findings, Finding{
			Line: result.DescriptionLine, Code: "E07", Message: "description is empty",
		})
	} else {
		result.DescriptionPoints = points(result.Description)
		if result.DescriptionPoints > maxDesc {
			result.Findings = append(result.Findings, Finding{
				Line: result.DescriptionLine,
				Code: "E08",
				Message: "description has " + strconv.Itoa(result.DescriptionPoints) +
					" code points, more than " + strconv.Itoa(maxDesc),
			})
		}
	}
	return result
}

func validFrontMatterKey(key string) bool {
	if len(key) == 0 || key[0] < 'a' || key[0] > 'z' {
		return false
	}
	for i := 1; i < len(key); i++ {
		c := key[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

func validSkillName(name string) bool {
	if len(name) < 1 || len(name) > 64 || name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
		if i > 0 && c == '-' && name[i-1] == '-' {
			return false
		}
	}
	return true
}
