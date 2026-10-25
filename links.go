package main

import "strings"

// LinkResult separates all counted link targets from local-link diagnostics.
type LinkResult struct {
	Count    int
	Findings []Finding
}

// joinPath joins virtual paths using the specified slash/backslash suffix
// rule. It intentionally does not clean or normalize either argument.
func joinPath(parent, child string) string {
	if strings.HasSuffix(parent, "/") || strings.HasSuffix(parent, "\\") {
		return parent + child
	}
	return parent + "/" + child
}

func scanLinks(body []string, firstLine int, skillFolder string, io Io) LinkResult {
	result := LinkResult{}
	inFence := false
	for lineIndex, line := range body {
		lineNumber := firstLine + lineIndex
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		position := 0
		for position < len(line) {
			marker := strings.Index(line[position:], "](")
			if marker < 0 {
				break
			}
			targetStart := position + marker + 2
			closeRelative := strings.IndexByte(line[targetStart:], ')')
			if closeRelative < 0 {
				break
			}
			closeIndex := targetStart + closeRelative
			target := line[targetStart:closeIndex]
			position = closeIndex + 1
			if target == "" || strings.ContainsAny(target, " \t") {
				continue
			}

			result.Count++
			code, message, checked := checkLocalLink(target, skillFolder, io)
			if checked {
				result.Findings = append(result.Findings, Finding{
					Line: lineNumber, Code: code, Message: message,
				})
			}
		}
	}
	return result
}

func checkLocalLink(target, skillFolder string, io Io) (string, string, bool) {
	originalTarget := target
	if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "/") ||
		strings.Contains(target, ":") {
		return "", "", false
	}
	if fragment := strings.IndexByte(target, '#'); fragment >= 0 {
		target = target[:fragment]
	}

	parts := make([]string, 0)
	for _, part := range strings.Split(target, "/") {
		switch part {
		case "", ".":
			continue
		case "..":
			if len(parts) == 0 {
				return "E10", "link " + originalTarget + " leaves the skill folder", true
			}
			parts = parts[:len(parts)-1]
		default:
			parts = append(parts, part)
		}
	}

	current := skillFolder
	for index, part := range parts {
		entries, ok := io.List(current)
		if !ok {
			return "E09", "link " + originalTarget + " does not exist", true
		}
		found := false
		kind := ""
		for _, entry := range entries {
			if entry.Name == part {
				found = true
				kind = entry.Kind
				break
			}
		}
		if !found || (index+1 < len(parts) && kind != "dir") {
			return "E09", "link " + originalTarget + " does not exist", true
		}
		if index+1 < len(parts) {
			current = joinPath(current, part)
		}
	}
	return "", "", false
}
