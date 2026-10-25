package main

import "strconv"

type SkillResult struct {
	Folder            string
	Findings          []Finding
	BodyLines         int
	LinkCount         int
	Description       string
	DescriptionLine   int
	DescriptionPoints int
	Fatal             bool
}

func analyzeSkill(io Io, folderPath, folder string, maxDesc, maxLines int) SkillResult {
	result := SkillResult{Folder: folder}
	entries, listed := io.List(folderPath)
	if !listed || !hasSkillFile(entries) {
		result.Fatal = true
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E01", Message: "SKILL.md is missing",
		})
		return result
	}

	data, read := io.Read(joinPath(folderPath, "SKILL.md"))
	if !read {
		result.Fatal = true
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E01", Message: "SKILL.md cannot be read",
		})
		return result
	}
	text, valid := normalize(data)
	if !valid {
		result.Fatal = true
		result.Findings = append(result.Findings, Finding{
			Line: 1, Code: "E02", Message: "file is not valid UTF-8",
		})
		return result
	}

	lines := splitLines(text)
	front := analyzeFrontMatter(lines, folder, maxDesc)
	result.Findings = append(result.Findings, front.Findings...)
	if front.Fatal {
		result.Fatal = true
		return result
	}
	result.Description = front.Description
	result.DescriptionLine = front.DescriptionLine
	result.DescriptionPoints = front.DescriptionPoints

	body := lines[front.ClosingLine:]
	result.BodyLines = len(body)
	bodyEmpty := true
	for _, line := range body {
		if !isBlank(line) {
			bodyEmpty = false
			break
		}
	}
	if bodyEmpty {
		result.Findings = append(result.Findings, Finding{
			Line: front.ClosingLine, Code: "W03", Message: "body is empty",
		})
	} else if result.BodyLines > maxLines {
		result.Findings = append(result.Findings, Finding{
			Line: front.ClosingLine + 1,
			Code: "W02",
			Message: "body has " + strconv.Itoa(result.BodyLines) +
				" lines, more than " + strconv.Itoa(maxLines),
		})
	}

	links := scanLinks(body, front.ClosingLine+1, folderPath, io)
	result.LinkCount = links.Count
	result.Findings = append(result.Findings, links.Findings...)
	return result
}

func hasSkillFile(entries []Entry) bool {
	for _, entry := range entries {
		if entry.Name == "SKILL.md" && entry.Kind == "file" {
			return true
		}
	}
	return false
}
