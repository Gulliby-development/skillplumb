package main

import "testing"

func TestIndexSort(t *testing.T) {
	header := "skillplumb: 5 skills\nSKILL  LINES  LINKS  DESC\n"
	cases := []struct {
		name   string
		args   []string
		status int
		stdout string
		stderr string
	}{
		{"name is the default order", []string{"index", "skills", "--sort", "name"}, 0, header +
			"alpha  2      2      23\nbeta   1      3      24\ndelta  0      0      -\ngamma  -      -      -\nzeta   -      -      -\n", ""},
		{"lines puts the longest body first", []string{"index", "skills", "--sort", "lines"}, 0, header +
			"alpha  2      2      23\nbeta   1      3      24\ndelta  0      0      -\ngamma  -      -      -\nzeta   -      -      -\n", ""},
		{"links puts the most links first", []string{"index", "skills", "--sort", "links"}, 0, header +
			"beta   1      3      24\nalpha  2      2      23\ndelta  0      0      -\ngamma  -      -      -\nzeta   -      -      -\n", ""},
		{"desc puts the longest description first", []string{"index", "skills", "--sort", "desc"}, 0, header +
			"beta   1      3      24\nalpha  2      2      23\ndelta  0      0      -\ngamma  -      -      -\nzeta   -      -      -\n", ""},
		{"the option may come before DIR", []string{"index", "--sort", "links", "skills"}, 0, header +
			"beta   1      3      24\nalpha  2      2      23\ndelta  0      0      -\ngamma  -      -      -\nzeta   -      -      -\n", ""},
		{"the option is optional", []string{"index", "skills"}, 0, header +
			"alpha  2      2      23\nbeta   1      3      24\ndelta  0      0      -\ngamma  -      -      -\nzeta   -      -      -\n", ""},
		{"a missing value", []string{"index", "skills", "--sort"}, 2, "", "skillplumb: --sort needs a value\n"},
		{"an unknown key", []string{"index", "skills", "--sort", "size"}, 2, "", "skillplumb: --sort needs one of name, lines, links, desc, got size\n"},
		{"keys are case-sensitive", []string{"index", "skills", "--sort", "Lines"}, 2, "", "skillplumb: --sort needs one of name, lines, links, desc, got Lines\n"},
		{"an empty key", []string{"index", "skills", "--sort", ""}, 2, "", "skillplumb: --sort needs one of name, lines, links, desc, got (empty)\n"},
		{"an option-looking key is consumed", []string{"index", "skills", "--sort", "--strict"}, 2, "", "skillplumb: --sort needs one of name, lines, links, desc, got --strict\n"},
		{"given twice", []string{"index", "skills", "--sort", "lines", "--sort", "links"}, 2, "", "skillplumb: --sort given twice\n"},
		{"the equals form is an unknown option", []string{"index", "skills", "--sort=lines"}, 2, "", "skillplumb: unknown option --sort=lines\n"},
		{"check does not take the option", []string{"check", "skills", "--sort", "lines"}, 2, "", "skillplumb: unknown option --sort\n"},
		{"index still rejects other options", []string{"index", "skills", "--strict"}, 2, "", "skillplumb: unknown option --strict\n"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			io := workedExampleIo()
			status := Run(test.args, io)
			if status != test.status {
				t.Errorf("Run(%q) status = %d, want %d", test.args, status, test.status)
			}
			if io.Stdout != test.stdout {
				t.Errorf("Run(%q) stdout = %q, want %q", test.args, io.Stdout, test.stdout)
			}
			if io.Stderr != test.stderr {
				t.Errorf("Run(%q) stderr = %q, want %q", test.args, io.Stderr, test.stderr)
			}
		})
	}
}

func TestIndexSortKeepsFolderOrderForEqualValues(t *testing.T) {
	files := map[string]string{
		"d/b/SKILL.md": "---\nname: b\ndescription: same\n---\nbody\n",
		"d/a/SKILL.md": "---\nname: a\ndescription: same\n---\nbody\n",
		"d/c/SKILL.md": "---\nname: c\ndescription: same\n---\nbody\n",
	}
	for _, key := range []string{"lines", "links", "desc"} {
		io := NewMemIo(files)
		if status := Run([]string{"index", "d", "--sort", key}, io); status != 0 {
			t.Fatalf("Run(index d --sort %s) status = %d, want 0", key, status)
		}
		want := "skillplumb: 3 skills\nSKILL  LINES  LINKS  DESC\na      1      0      4\nb      1      0      4\nc      1      0      4\n"
		if io.Stdout != want {
			t.Errorf("--sort %s stdout = %q, want %q", key, io.Stdout, want)
		}
	}
}
