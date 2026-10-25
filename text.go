package main

import (
	"strings"
	"unicode/utf8"
)

type Failure string

func (f Failure) Error() string {
	return string(f)
}

func shown(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			b.WriteByte('?')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func points(s string) int {
	return utf8.RuneCountInString(s)
}

func isBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return true
}

func trimBlanks(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func normalize(data []byte) (string, bool) {
	if !utf8.Valid(data) {
		return "", false
	}
	for _, b := range data {
		if b == 0 {
			return "", false
		}
	}

	text := string(data)
	if strings.HasPrefix(text, "\ufeff") {
		text = strings.TrimPrefix(text, "\ufeff")
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return text, true
}

func splitLines(text string) []string {
	if text == "" {
		return []string{}
	}
	lines := strings.Split(text, "\n")
	if strings.HasSuffix(text, "\n") {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func foldKey(s string) string {
	var b strings.Builder
	inBlanks := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' {
			if !inBlanks {
				b.WriteByte(' ')
				inBlanks = true
			}
			continue
		}
		inBlanks = false
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b.WriteByte(c)
	}
	return b.String()
}

func table(header []string, rows [][]string) string {
	columns := len(header)
	for _, row := range rows {
		if len(row) > columns {
			columns = len(row)
		}
	}

	widths := make([]int, columns)
	displayHeader := make([]string, columns)
	for column := 0; column < columns; column++ {
		if column < len(header) {
			displayHeader[column] = shown(header[column])
			widths[column] = points(displayHeader[column])
		}
	}
	displayRows := make([][]string, len(rows))
	for rowIndex, row := range rows {
		displayRows[rowIndex] = make([]string, columns)
		for column := 0; column < columns; column++ {
			if column < len(row) {
				displayRows[rowIndex][column] = shown(row[column])
				if cellWidth := points(displayRows[rowIndex][column]); cellWidth > widths[column] {
					widths[column] = cellWidth
				}
			}
		}
	}

	var output strings.Builder
	writeRow := func(cells []string) {
		var line strings.Builder
		for column := 0; column < columns; column++ {
			if column > 0 {
				line.WriteString("  ")
			}
			cell := cells[column]
			line.WriteString(cell)
			if column+1 < columns {
				for padding := points(cell); padding < widths[column]; padding++ {
					line.WriteByte(' ')
				}
			}
		}
		rendered := line.String()
		for strings.HasSuffix(rendered, " ") {
			rendered = strings.TrimSuffix(rendered, " ")
		}
		output.WriteString(rendered)
		output.WriteByte('\n')
	}
	writeRow(displayHeader)
	for _, row := range displayRows {
		writeRow(row)
	}
	return output.String()
}
