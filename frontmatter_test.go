package main

import "testing"

func TestFrontMatterRuleTable(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		stdout string
		status int
	}{
		{
			name:   "valid metadata",
			file:   "---\nname: s\ndescription: Does a thing.\n---\nx\n",
			stdout: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			status: 0,
		},
		{
			name:   "unknown key",
			file:   "---\nname: s\ndescription: d\nversion: 2\n---\nx\n",
			stdout: "s/SKILL.md:4: W01 unknown key version\nskillplumb: 1 skills, 0 errors, 1 warnings\n",
			status: 0,
		},
		{
			name:   "duplicate key keeps first value",
			file:   "---\nname: s\nname: t\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:3: E05 key name given twice\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "missing name",
			file:   "---\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:1: E06 name is missing\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "empty name",
			file:   "---\nname:\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:2: E06 name is empty\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "empty description after trimming blanks",
			file:   "---\nname: s\ndescription: \n---\nx\n",
			stdout: "s/SKILL.md:3: E07 description is empty\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "uppercase name",
			file:   "---\nname: S\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:2: E06 name S is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "2 E06 name a--b is not valid",
			file:   "---\nname: a--b\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:2: E06 name a--b is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "name does not match folder",
			file:   "---\nname: t\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:2: E06 name t does not match folder s\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "uppercase key is malformed",
			file:   "---\nname: s\ndescription: d\nName: x\n---\nx\n",
			stdout: "s/SKILL.md:4: E04 line is not key: value\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "spaced key is malformed and name is missing",
			file:   "---\nname : s\ndescription: d\n---\nx\n",
			stdout: "s/SKILL.md:1: E06 name is missing\ns/SKILL.md:2: E04 line is not key: value\nskillplumb: 1 skills, 2 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "blank and comment lines are skipped",
			file:   "---\nname: s\n# note\n\t\ndescription: d\n---\nx\n",
			stdout: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			status: 0,
		},
		{
			name:   "colon in value is retained",
			file:   "---\nname: s\ndescription: a: b\n---\nx\n",
			stdout: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			status: 0,
		},
		{
			name:   "accepted front matter key grammar",
			file:   "---\nname: s\ndescription: d\na: one\na0: two\na-z: three\naz9: four\nlicense: MIT\n---\nx\n",
			stdout: "s/SKILL.md:4: W01 unknown key a\ns/SKILL.md:5: W01 unknown key a0\ns/SKILL.md:6: W01 unknown key a-z\ns/SKILL.md:7: W01 unknown key az9\nskillplumb: 1 skills, 0 errors, 4 warnings\n",
			status: 0,
		},
		{
			name:   "invalid front matter key characters",
			file:   "---\nname: s\ndescription: d\n9a: one\na_b: two\na.: three\n---\nx\n",
			stdout: "s/SKILL.md:4: E04 line is not key: value\ns/SKILL.md:5: E04 line is not key: value\ns/SKILL.md:6: E04 line is not key: value\nskillplumb: 1 skills, 3 errors, 0 warnings\n",
			status: 1,
		},
		{
			name:   "invalid line unknown key and duplicate unknown key",
			file:   "---\nname: s\ndescription: d\nx_y: 1\nversion: 1\nversion: 2\n---\nx\n",
			stdout: "s/SKILL.md:4: E04 line is not key: value\ns/SKILL.md:5: W01 unknown key version\ns/SKILL.md:6: E05 key version given twice\nskillplumb: 1 skills, 2 errors, 1 warnings\n",
			status: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mem := NewMemIo(map[string]string{"s/SKILL.md": test.file})
			status := Run([]string{"check", "."}, mem)
			if status != test.status {
				t.Errorf("Run status = %d, want %d", status, test.status)
			}
			if mem.Stdout != test.stdout {
				t.Errorf("stdout = %q, want %q", mem.Stdout, test.stdout)
			}
			if mem.Stderr != "" {
				t.Errorf("stderr = %q, want empty", mem.Stderr)
			}
		})
	}
}

func TestFrontMatterTerminationAndDescriptionLimits(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		args   []string
		stdout string
		status int
		index  string
	}{
		{
			name:   "inexact opening delimiter",
			file:   "--- \nname: s\ndescription: d\n---\nx\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:1: E03 front matter is missing or not closed\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:   "unclosed front matter",
			file:   "---\nname: s\ndescription: d\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:1: E03 front matter is missing or not closed\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      -      -      -\n",
		},
		{
			name:   "empty front matter",
			file:   "---\n---\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:1: E06 name is missing\ns/SKILL.md:1: E07 description is missing\ns/SKILL.md:2: W03 body is empty\nskillplumb: 1 skills, 2 errors, 1 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      0      0      -\n",
		},
		{
			name:   "leading hyphen name",
			file:   "---\nname: -s\ndescription: d\n---\nx\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:2: E06 name -s is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      1\n",
		},
		{
			name:   "trailing hyphen name",
			file:   "---\nname: s-\ndescription: d\n---\nx\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:2: E06 name s- is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      1\n",
		},
		{
			name:   "underscore name",
			file:   "---\nname: s_s\ndescription: d\n---\nx\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:2: E06 name s_s is not valid\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      1\n",
		},
		{
			name:   "missing description",
			file:   "---\nname: s\n---\nx\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:1: E07 description is missing\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      -\n",
		},
		{
			name:   "empty description",
			file:   "---\nname: s\ndescription: \t \n---\nx\n",
			args:   []string{"check", "."},
			stdout: "s/SKILL.md:3: E07 description is empty\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      -\n",
		},
		{
			name:   "ASCII description over limit",
			file:   "---\nname: s\ndescription: abcdef\n---\nx\n",
			args:   []string{"check", ".", "--max-desc", "5"},
			stdout: "s/SKILL.md:3: E08 description has 6 code points, more than 5\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      6\n",
		},
		{
			name:   "six emoji description over limit",
			file:   "---\nname: s\ndescription: 😀😀😀😀😀😀\n---\nx\n",
			args:   []string{"check", ".", "--max-desc", "5"},
			stdout: "s/SKILL.md:3: E08 description has 6 code points, more than 5\nskillplumb: 1 skills, 1 errors, 0 warnings\n",
			status: 1,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      6\n",
		},
		{
			name:   "five code points is inclusive",
			file:   "---\nname: s\ndescription: abcde\n---\nx\n",
			args:   []string{"check", ".", "--max-desc", "5"},
			stdout: "skillplumb: 1 skills, 0 errors, 0 warnings\n",
			status: 0,
			index:  "skillplumb: 1 skills\nSKILL  LINES  LINKS  DESC\ns      1      0      5\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			check := NewMemIo(map[string]string{"s/SKILL.md": test.file})
			if status := Run(test.args, check); status != test.status {
				t.Errorf("Run status = %d, want %d", status, test.status)
			}
			if check.Stdout != test.stdout {
				t.Errorf("stdout = %q, want %q", check.Stdout, test.stdout)
			}
			if check.Stderr != "" {
				t.Errorf("stderr = %q, want empty", check.Stderr)
			}

			index := NewMemIo(map[string]string{"s/SKILL.md": test.file})
			if status := Run([]string{"index", "."}, index); status != 0 {
				t.Errorf("index status = %d, want 0", status)
			}
			if index.Stdout != test.index {
				t.Errorf("index stdout = %q, want %q", index.Stdout, test.index)
			}
			if index.Stderr != "" {
				t.Errorf("index stderr = %q, want empty", index.Stderr)
			}
		})
	}
}
