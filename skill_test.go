package main

import "testing"

func TestSkillFatalFileBoundariesAndIndex(t *testing.T) {
	cases := []struct {
		name       string
		files      map[string]string
		unreadable bool
		checkWant  string
		indexWant  string
	}{
		{
			name:      "missing SKILL.md",
			files:     map[string]string{"d/s/": ""},
			checkWant: "s/SKILL.md:1: E01 SKILL.md is missing\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:       "unreadable SKILL.md",
			files:      map[string]string{"d/s/SKILL.md": "---\nname: s\ndescription: d\n---\nx"},
			unreadable: true,
			checkWant:  "s/SKILL.md:1: E01 SKILL.md cannot be read\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "empty file",
			files:     map[string]string{"d/s/SKILL.md": ""},
			checkWant: "s/SKILL.md:1: E03 front matter is missing or not closed\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "NUL byte",
			files:     map[string]string{"d/s/SKILL.md": string([]byte{0x00})},
			checkWant: "s/SKILL.md:1: E02 file is not valid UTF-8\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "surrogate encoding",
			files:     map[string]string{"d/s/SKILL.md": string([]byte{0xed, 0xa0, 0x80})},
			checkWant: "s/SKILL.md:1: E02 file is not valid UTF-8\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "overlong encoding",
			files:     map[string]string{"d/s/SKILL.md": string([]byte{0xc0, 0xaf})},
			checkWant: "s/SKILL.md:1: E02 file is not valid UTF-8\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "code point above Unicode maximum",
			files:     map[string]string{"d/s/SKILL.md": string([]byte{0xf4, 0x90, 0x80, 0x80})},
			checkWant: "s/SKILL.md:1: E02 file is not valid UTF-8\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "lone CR on opening delimiter",
			files:     map[string]string{"d/s/SKILL.md": "---\rname: s\ndescription: d\n---\nx"},
			checkWant: "s/SKILL.md:1: E03 front matter is missing or not closed\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:      "lone CR on closing delimiter",
			files:     map[string]string{"d/s/SKILL.md": "---\nname: s\ndescription: d\n---\rx"},
			checkWant: "s/SKILL.md:1: E03 front matter is missing or not closed\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			checkIO := NewMemIo(test.files)
			if test.unreadable {
				checkIO.Unreadable["d/s/SKILL.md"] = true
			}
			if status := Run([]string{"check", "d"}, checkIO); status != 1 {
				t.Errorf("check status = %d, want 1", status)
			}
			if got := checkIO.Stderr; got != "" {
				t.Errorf("check stderr = %q, want empty", got)
			}
			if got := checkIO.Stdout; got != test.checkWant {
				t.Errorf("check stdout = %q, want %q", got, test.checkWant)
			}

			indexIO := NewMemIo(test.files)
			if test.unreadable {
				indexIO.Unreadable["d/s/SKILL.md"] = true
			}
			if status := Run([]string{"index", "d"}, indexIO); status != 0 {
				t.Errorf("index status = %d, want 0", status)
			}
			if got := indexIO.Stderr; got != "" {
				t.Errorf("index stderr = %q, want empty", got)
			}
			if got := indexIO.Stdout; got != test.indexWant {
				t.Errorf("index stdout = %q, want %q", got, test.indexWant)
			}
		})
	}
}

func TestSkillBodyAccounting(t *testing.T) {
	cases := []struct {
		name      string
		file      string
		checkWant string
		indexWant string
	}{
		{
			name:      "zero body lines",
			file:      "---\nname: s\ndescription: d\n---",
			checkWant: "s/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      0      0      1\n",
		},
		{
			name:      "space and tab body lines",
			file:      "---\nname: s\ndescription: d\n---\n \n\t",
			checkWant: "s/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      2      0      1\n",
		},
		{
			name:      "three body lines at limit",
			file:      "---\nname: s\ndescription: d\n---\na\nb\nc\n",
			checkWant: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      3      0      1\n",
		},
		{
			name:      "four body lines above limit",
			file:      "---\nname: s\ndescription: d\n---\na\nb\nc\nd",
			checkWant: "s/SKILL.md:5: W02 body has 4 lines, more than 3\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      4      0      1\n",
		},
		{
			name:      "nonblank body with trailing blank lines",
			file:      "---\nname: s\ndescription: d\n---\na\n\n\n\n",
			checkWant: "s/SKILL.md:5: W02 body has 4 lines, more than 3\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      4      0      1\n",
		},
		{
			name:      "three blank body lines take empty precedence",
			file:      "---\nname: s\ndescription: d\n---\n \n\t\n \n\t",
			checkWant: "s/SKILL.md:4: W03 body is empty\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
			indexWant: "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      4      0      1\n",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			checkIO := NewMemIo(map[string]string{"d/s/SKILL.md": test.file})
			if status := Run([]string{"check", "d", "--max-lines", "3"}, checkIO); status != 0 {
				t.Errorf("check status = %d, want 0", status)
			}
			if got := checkIO.Stderr; got != "" {
				t.Errorf("check stderr = %q, want empty", got)
			}
			if got := checkIO.Stdout; got != test.checkWant {
				t.Errorf("check stdout = %q, want %q", got, test.checkWant)
			}

			indexIO := NewMemIo(map[string]string{"d/s/SKILL.md": test.file})
			if status := Run([]string{"index", "d"}, indexIO); status != 0 {
				t.Errorf("index status = %d, want 0", status)
			}
			if got := indexIO.Stderr; got != "" {
				t.Errorf("index stderr = %q, want empty", got)
			}
			if got := indexIO.Stdout; got != test.indexWant {
				t.Errorf("index stdout = %q, want %q", got, test.indexWant)
			}
		})
	}
}

func TestSkillDuplicateSelectionAndFindingOrdering(t *testing.T) {
	cases := []struct {
		name   string
		files  map[string]string
		args   []string
		status int
		stdout string
	}{
		{
			name: "earliest of three with ASCII case and blank folding",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: HELLO\t world\n---\nx",
				"d/b/SKILL.md": "---\nname: b\ndescription: hello  WORLD\n---\nx",
				"d/c/SKILL.md": "---\nname: c\ndescription: Hello world\n---\nx",
			},
			args:   []string{"check", "d"},
			status: 0,
			stdout: "b/SKILL.md:3: W04 description is the same as in skill a\nc/SKILL.md:3: W04 description is the same as in skill a\nskillplumb: 3 skills, 0 errors, 2 warnings\n",
		},
		{
			name: "ASCII case-only duplicate",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: Release NOTE\n---\nx",
				"d/b/SKILL.md": "---\nname: b\ndescription: release note\n---\nx",
			},
			args:   []string{"check", "d"},
			status: 0,
			stdout: "b/SKILL.md:3: W04 description is the same as in skill a\nskillplumb: 2 skills, 0 errors, 1 warnings\n",
		},
		{
			name: "non-ASCII case remains distinct",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: É\n---\nx",
				"d/b/SKILL.md": "---\nname: b\ndescription: é\n---\nx",
			},
			args:   []string{"check", "d"},
			status: 0,
			stdout: "skillplumb: 2 skills, 0 errors, 0 warnings\n",
		},
		{
			name: "empty descriptions are not duplicates",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: \n---\nx",
				"d/b/SKILL.md": "---\nname: b\ndescription: \n---\nx",
			},
			args:   []string{"check", "d"},
			status: 1,
			stdout: "a/SKILL.md:3: E07 description is empty\nb/SKILL.md:3: E07 description is empty\nskillplumb: 2 skills, 2 errors, 0 warnings\n",
		},
		{
			name: "duplicate emitted with other findings",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: shared\nversion: 2\n---\nx",
				"d/b/SKILL.md": "---\nname: wrong\ndescription: shared\n---\nx",
			},
			args:   []string{"check", "d"},
			status: 1,
			stdout: "a/SKILL.md:4: W01 unknown key version\nb/SKILL.md:2: E06 name wrong does not match folder b\nb/SKILL.md:3: W04 description is the same as in skill a\nskillplumb: 2 skills, 1 errors, 2 warnings\n",
		},
		{
			name: "ignore duplicate finding and its count",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: shared\n---\nx",
				"d/b/SKILL.md": "---\nname: b\ndescription: shared\n---\nx",
			},
			args:   []string{"check", "d", "--ignore", "W04"},
			status: 0,
			stdout: "skillplumb: 2 skills, 0 errors, 0 warnings\n",
		},
		{
			name: "line then code then stable tie ordering",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: d\n---\n[a](n1) [b](n2)\nx",
			},
			args:   []string{"check", "d", "--max-lines", "1"},
			status: 1,
			stdout: "a/SKILL.md:5: E09 link n1 does not exist\na/SKILL.md:5: E09 link n2 does not exist\na/SKILL.md:5: W02 body has 2 lines, more than 1\nskillplumb: 1 skills, 2 errors, 1 warnings\n",
		},
		{
			name: "ignored unrelated warning does not change earliest source",
			files: map[string]string{
				"d/a/SKILL.md": "---\nname: a\ndescription: shared\n---\nx",
				"d/b/SKILL.md": "---\nname: b\ndescription: shared\nversion: 2\n---\nx",
				"d/c/SKILL.md": "---\nname: c\ndescription: shared\n---\nx",
			},
			args:   []string{"check", "d", "--ignore", "W01"},
			status: 0,
			stdout: "b/SKILL.md:3: W04 description is the same as in skill a\nc/SKILL.md:3: W04 description is the same as in skill a\nskillplumb: 3 skills, 0 errors, 2 warnings\n",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			io := NewMemIo(test.files)
			if status := Run(test.args, io); status != test.status {
				t.Errorf("Run status = %d, want %d", status, test.status)
			}
			if got := io.Stderr; got != "" {
				t.Errorf("Run stderr = %q, want empty", got)
			}
			if got := io.Stdout; got != test.stdout {
				t.Errorf("Run stdout = %q, want %q", got, test.stdout)
			}
		})
	}
}
