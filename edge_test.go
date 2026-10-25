package main

import "testing"

func TestDiscoveryFiltersAndOrdersSkillFolders(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"a-b/SKILL.md":    "---\nname: a-b\ndescription: x\n---\nx\n",
		"a0/SKILL.md":     "---\nname: a0\ndescription: x\n---\nx\n",
		"～/SKILL.md":      "---\nname: ～\ndescription: x\n---\nx\n",
		"😀/SKILL.md":      "---\nname: 😀\ndescription: x\n---\nx\n",
		".cache/SKILL.md": "---\nname: hidden\ndescription: x\n---\nx\n",
		"README.md":       "not a skill",
		"\xff/SKILL.md":   "---\nname: invalid\ndescription: x\n---\nx\n",
	}, "external")

	if status := Run([]string{"index", "."}, mem); status != 0 {
		t.Errorf("Run(index .) status = %d, want 0", status)
	}
	want := "skillplumb: 4 skills\n" +
		"SKILL  LINES  LINKS  DESC\n" +
		"a-b    1      0      1\n" +
		"a0     1      0      1\n" +
		"～      1      0      1\n" +
		"😀      1      0      1\n"
	if mem.Stdout != want {
		t.Errorf("index stdout = %q, want %q", mem.Stdout, want)
	}
	if mem.Stderr != "" {
		t.Errorf("index stderr = %q, want empty", mem.Stderr)
	}
	hidden, listed := mem.List(".cache")
	if !listed || len(hidden) != 1 || hidden[0] != (Entry{Name: "SKILL.md", Kind: "file"}) {
		t.Errorf(`List(".cache") = %#v, %t, want the SKILL.md file`, hidden, listed)
	}
}

func TestDiscoveryPathJoining(t *testing.T) {
	cases := []struct {
		name  string
		dir   string
		files map[string]string
		want  string
	}{
		{
			name: "dot root",
			dir:  ".",
			files: map[string]string{
				"a/SKILL.md": "---\nname: a\ndescription: x\n---\nx\n",
			},
			want: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\na      1      0      1\n",
		},
		{
			name: "leading dot slash",
			dir:  "./d",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: x\n---\nx\n",
			},
			want: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\na      1      0      1\n",
		},
		{
			name: "trailing slash",
			dir:  "d/",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: x\n---\nx\n",
			},
			want: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\na      1      0      1\n",
		},
		{
			name: "trailing backslash",
			dir:  "d\\",
			files: map[string]string{
				"d\\/a/placeholder.txt": "placeholder",
				"d\\a/SKILL.md":         "---\nname: a\ndescription: x\n---\nx\n",
			},
			want: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\na      1      0      1\n",
		},
	}

	for _, test := range cases {
		mem := NewMemIo(test.files)
		if status := Run([]string{"index", test.dir}, mem); status != 0 {
			t.Errorf("%s: Run(index %q) status = %d, want 0", test.name, test.dir, status)
		}
		if mem.Stdout != test.want {
			t.Errorf("%s: index stdout = %q, want %q", test.name, mem.Stdout, test.want)
		}
		if mem.Stderr != "" {
			t.Errorf("%s: index stderr = %q, want empty", test.name, mem.Stderr)
		}
	}
}

func TestEmptyDirectoryArgument(t *testing.T) {
	mem := NewMemIo(map[string]string{})
	if status := Run([]string{"check", ""}, mem); status != 2 {
		t.Errorf("Run(check empty DIR) status = %d, want 2", status)
	}
	if mem.Stdout != "" {
		t.Errorf("stdout = %q, want empty", mem.Stdout)
	}
	if got, want := mem.Stderr, "skillplumb: (empty): not a directory\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestSkillFileKindsAndUnreadableFile(t *testing.T) {
	cases := []struct {
		name       string
		files      map[string]string
		others     []string
		unreadable string
		want       string
	}{
		{
			name:  "missing file",
			files: map[string]string{"s/notes.txt": "notes"},
			want:  "s/SKILL.md:1: E01 SKILL.md is missing\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
		},
		{
			name:  "directory named SKILL.md",
			files: map[string]string{"s/SKILL.md/child.txt": "child"},
			want:  "s/SKILL.md:1: E01 SKILL.md is missing\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
		},
		{
			name:   "other entry named SKILL.md",
			others: []string{"s/SKILL.md"},
			want:   "s/SKILL.md:1: E01 SKILL.md is missing\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
		},
		{
			name:       "Unreadable regular file",
			files:      map[string]string{"s/SKILL.md": "---\nname: s\ndescription: x\n---\nx\n"},
			unreadable: "s/SKILL.md",
			want:       "s/SKILL.md:1: E01 SKILL.md cannot be read\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
		},
	}

	for _, test := range cases {
		mem := NewMemIo(test.files, test.others...)
		if test.unreadable != "" {
			mem.Unreadable[test.unreadable] = true
		}
		if status := Run([]string{"check", "."}, mem); status != 1 {
			t.Errorf("%s: Run(check .) status = %d, want 1", test.name, status)
		}
		if mem.Stdout != test.want {
			t.Errorf("%s: stdout = %q, want %q", test.name, mem.Stdout, test.want)
		}
		if mem.Stderr != "" {
			t.Errorf("%s: stderr = %q, want empty", test.name, mem.Stderr)
		}
	}
}

func TestNormalizeByteAndLineBoundaries(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"crlf/SKILL.md":       "\uFEFFa\r\nb\rc\n",
		"second-bom/SKILL.md": "\uFEFF\uFEFFx",
		"empty/SKILL.md":      "",
		"final-lf/SKILL.md":   "a\n\n",
		"nul/SKILL.md":        "\x00",
		"surrogate/SKILL.md":  "\xed\xa0\x80",
		"overlong/SKILL.md":   "\xc0\xaf",
		"above-max/SKILL.md":  "\xf4\x90\x80\x80",
	})

	cases := []struct {
		path      string
		wantText  string
		wantLines []string
		wantValid bool
	}{
		{"crlf/SKILL.md", "a\nb\rc\n", []string{"a", "b\rc"}, true},
		{"second-bom/SKILL.md", "\uFEFFx", []string{"\uFEFFx"}, true},
		{"empty/SKILL.md", "", []string{}, true},
		{"final-lf/SKILL.md", "a\n\n", []string{"a", ""}, true},
		{"nul/SKILL.md", "", nil, false},
		{"surrogate/SKILL.md", "", nil, false},
		{"overlong/SKILL.md", "", nil, false},
		{"above-max/SKILL.md", "", nil, false},
	}
	for _, test := range cases {
		data, read := mem.Read(test.path)
		if !read {
			t.Fatalf("Read(%q) failed", test.path)
		}
		got, valid := normalize(data)
		if valid != test.wantValid {
			t.Errorf("%s: normalize validity = %t, want %t", test.path, valid, test.wantValid)
		}
		if !valid {
			continue
		}
		if got != test.wantText {
			t.Errorf("%s: normalized text = %q, want %q", test.path, got, test.wantText)
		}
		lines := splitLines(got)
		if len(lines) != len(test.wantLines) {
			t.Errorf("%s: splitLines length = %d, want %d", test.path, len(lines), len(test.wantLines))
			continue
		}
		for i := range lines {
			if lines[i] != test.wantLines[i] {
				t.Errorf("%s: line %d = %q, want %q", test.path, i+1, lines[i], test.wantLines[i])
			}
		}
	}
}

func TestFileEncodingAndFrontMatterBoundaries(t *testing.T) {
	invalid := []struct {
		path string
		data string
	}{
		{"nul", "\x00"},
		{"surrogate", "\xed\xa0\x80"},
		{"overlong", "\xc0\xaf"},
		{"above-max", "\xf4\x90\x80\x80"},
	}
	for _, test := range invalid {
		mem := NewMemIo(map[string]string{"s/SKILL.md": test.data})
		if status := Run([]string{"check", "."}, mem); status != 1 {
			t.Errorf("%s: Run(check .) status = %d, want 1", test.path, status)
		}
		want := "s/SKILL.md:1: E02 file is not valid UTF-8\nskillplumb: 1 skills, 1 errors, 0 warnings\n"
		if mem.Stdout != want {
			t.Errorf("%s: stdout = %q, want %q", test.path, mem.Stdout, want)
		}
		if mem.Stderr != "" {
			t.Errorf("%s: stderr = %q, want empty", test.path, mem.Stderr)
		}
	}

	fatalCases := []struct {
		name string
		file string
	}{
		{"inexact opener", "--- \nname: s\ndescription: d\n---\nx\n"},
		{"unclosed front matter", "---\nname: s\ndescription: d\n"},
	}
	for _, test := range fatalCases {
		mem := NewMemIo(map[string]string{"s/SKILL.md": test.file})
		if status := Run([]string{"check", "."}, mem); status != 1 {
			t.Errorf("%s: Run(check .) status = %d, want 1", test.name, status)
		}
		want := "s/SKILL.md:1: E03 front matter is missing or not closed\nskillplumb: 1 skills, 1 errors, 0 warnings\n"
		if mem.Stdout != want {
			t.Errorf("%s: stdout = %q, want %q", test.name, mem.Stdout, want)
		}
	}
	mem := NewMemIo(map[string]string{"s/SKILL.md": "---\n---\n"})
	if status := Run([]string{"check", "."}, mem); status != 1 {
		t.Errorf("Run(check empty front matter) status = %d, want 1", status)
	}
	want := "s/SKILL.md:1: E06 name is missing\ns/SKILL.md:1: E07 description is missing\ns/SKILL.md:2: W03 body is empty\nskillplumb: 1 skills, 2 errors, 1 warnings\n"
	if mem.Stdout != want {
		t.Errorf("empty front matter stdout = %q, want %q", mem.Stdout, want)
	}

	name64 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mem = NewMemIo(map[string]string{
		name64 + "/SKILL.md": "---\nname: " + name64 + "\ndescription: d\n---\nx\n",
	})
	if status := Run([]string{"check", "."}, mem); status != 0 {
		t.Errorf("Run(check 64-character name) status = %d, want 0", status)
	}
	if mem.Stdout != "skillplumb: 1 skills, 0 errors, 0 warnings\n" {
		t.Errorf("64-character name stdout = %q, want no findings", mem.Stdout)
	}

	name65 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mem = NewMemIo(map[string]string{
		"s/SKILL.md": "---\nname: " + name65 + "\ndescription: d\n---\nx\n",
	})
	if status := Run([]string{"check", "."}, mem); status != 1 {
		t.Errorf("Run(check 65-character name) status = %d, want 1", status)
	}
	want = "s/SKILL.md:2: E06 name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n"
	if mem.Stdout != want {
		t.Errorf("65-character name stdout = %q, want %q", mem.Stdout, want)
	}
}

func TestDescriptionCodePointBoundaries(t *testing.T) {
	description200 := "😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀"
	description201 := "😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀😀😀😀😀😀😀😀😀😀" +
		"😀"
	cases := []struct {
		name        string
		description string
		status      int
		check       string
		index       string
	}{
		{
			name:        "Default limit with a literal 200-emoji description",
			description: description200,
			status:      0,
			check:       "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			index:       "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      200\n",
		},
		{
			name:        "Default limit with a literal 201-emoji description",
			description: description201,
			status:      1,
			check:       "s/SKILL.md:3: E08 description has 201 code points, more than 200\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			index:       "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      201\n",
		},
	}
	for _, test := range cases {
		file := "---\nname: s\ndescription: " + test.description + "\n---\nx\n"
		mem := NewMemIo(map[string]string{"s/SKILL.md": file})
		if status := Run([]string{"check", "."}, mem); status != test.status {
			t.Errorf("%s: check status = %d, want %d", test.name, status, test.status)
		}
		if mem.Stdout != test.check {
			t.Errorf("%s: check stdout = %q, want %q", test.name, mem.Stdout, test.check)
		}
		if mem.Stderr != "" {
			t.Errorf("%s: check stderr = %q, want empty", test.name, mem.Stderr)
		}

		mem = NewMemIo(map[string]string{"s/SKILL.md": file})
		if status := Run([]string{"index", "."}, mem); status != 0 {
			t.Errorf("%s: index status = %d, want 0", test.name, status)
		}
		if mem.Stdout != test.index {
			t.Errorf("%s: index stdout = %q, want %q", test.name, mem.Stdout, test.index)
		}
	}
}

func TestFatalIndexRowsUseDashes(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"a/notes.txt": "notes",
		"b/SKILL.md":  "\x00",
		"c/SKILL.md":  "not front matter",
	})
	if status := Run([]string{"index", "."}, mem); status != 0 {
		t.Errorf("Run(index fatal rows) status = %d, want 0", status)
	}
	want := "skillplumb: 3 skills\n" +
		"SKILL  LINES  LINKS  DESC\n" +
		"a      -      -      -\n" +
		"b      -      -      -\n" +
		"c      -      -      -\n"
	if mem.Stdout != want {
		t.Errorf("fatal index stdout = %q, want %q", mem.Stdout, want)
	}
	if mem.Stderr != "" {
		t.Errorf("fatal index stderr = %q, want empty", mem.Stderr)
	}
}

func TestControlDisplayAndCodePointTableAlignment(t *testing.T) {
	file := "---\nname: a\tb\ndescription: 😀\n---\n[x](bad\x7f.md)\n"
	mem := NewMemIo(map[string]string{"a\tb/SKILL.md": file})
	if status := Run([]string{"check", "."}, mem); status != 1 {
		t.Errorf("Run(check control display) status = %d, want 1", status)
	}
	wantCheck := "a?b/SKILL.md:2: E06 name a?b is not valid\n" +
		"a?b/SKILL.md:5: E09 link bad?.md does not exist\n" +
		"skillplumb: 1 skills, 2 errors, 0 warnings\n"
	if mem.Stdout != wantCheck {
		t.Errorf("control check stdout = %q, want %q", mem.Stdout, wantCheck)
	}
	if mem.Stderr != "" {
		t.Errorf("control check stderr = %q, want empty", mem.Stderr)
	}

	mem = NewMemIo(map[string]string{"a\tb/SKILL.md": file})
	if status := Run([]string{"index", "."}, mem); status != 0 {
		t.Errorf("Run(index control display) status = %d, want 0", status)
	}
	wantIndex := "skillplumb: 1 skills\n" +
		"SKILL  LINES  LINKS  DESC\n" +
		"a?b    1      1      1\n"
	if mem.Stdout != wantIndex {
		t.Errorf("control index stdout = %q, want %q", mem.Stdout, wantIndex)
	}
	if got, want := table([]string{"A", "B"}, [][]string{{"😀", "x"}, {"aa", ""}}), "A   B\n😀   x\naa\n"; got != want {
		t.Errorf("code-point table = %q, want %q", got, want)
	}
	if got, want := table([]string{"A"}, [][]string{{"x\x7f"}}), "A\nx?\n"; got != want {
		t.Errorf("sanitized table cell = %q, want %q", got, want)
	}
	if len(mem.Stdout) == 0 || mem.Stdout[0] == 0xef || mem.Stdout[len(mem.Stdout)-1] != 10 {
		t.Errorf("index output does not have the required UTF-8/LF boundaries: %q", mem.Stdout)
	}
	for _, line := range splitLines(mem.Stdout) {
		if line != trimBlanks(line) || (len(line) > 0 && line[len(line)-1] == ' ') {
			t.Errorf("index output line has trailing whitespace: %q", line)
		}
	}
}
