package main

import "testing"

func TestShownAndPoints(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"shown.txt":   "a\t\n\x00\x7fb😀",
		"points.txt":  "a😀",
		"unicode.txt": "😀é",
	})
	shownInput, ok := mem.Read("shown.txt")
	if !ok {
		t.Fatal("shown fixture is missing")
	}
	if got, want := shown(string(shownInput)), "a????b😀"; got != want {
		t.Errorf("shown(%q) = %q, want %q", string(shownInput), got, want)
	}

	pointsInput, ok := mem.Read("points.txt")
	if !ok {
		t.Fatal("points fixture is missing")
	}
	if got, want := points(string(pointsInput)), 2; got != want {
		t.Errorf("points(%q) = %d, want %d", string(pointsInput), got, want)
	}

	unicodeInput, ok := mem.Read("unicode.txt")
	if !ok {
		t.Fatal("unicode fixture is missing")
	}
	if got, want := points(string(unicodeInput)), 2; got != want {
		t.Errorf("points(%q) = %d, want %d", string(unicodeInput), got, want)
	}
}

func TestBlankHelpers(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"spaces-tabs.txt": " \t",
		"empty.txt":       "",
		"nbsp.txt":        "\u00a0",
		"trim.txt":        " \tx\u00a0\t ",
	})
	spacesTabs, ok := mem.Read("spaces-tabs.txt")
	if !ok {
		t.Fatal("spaces and tabs fixture is missing")
	}
	if got := isBlank(string(spacesTabs)); !got {
		t.Errorf("isBlank(%q) = false, want true", string(spacesTabs))
	}

	empty, ok := mem.Read("empty.txt")
	if !ok {
		t.Fatal("empty fixture is missing")
	}
	if got := isBlank(string(empty)); !got {
		t.Errorf("isBlank(%q) = false, want true", string(empty))
	}

	nbsp, ok := mem.Read("nbsp.txt")
	if !ok {
		t.Fatal("NBSP fixture is missing")
	}
	if got := isBlank(string(nbsp)); got {
		t.Errorf("isBlank(%q) = true, want false", string(nbsp))
	}

	trimInput, ok := mem.Read("trim.txt")
	if !ok {
		t.Fatal("trim fixture is missing")
	}
	if got, want := trimBlanks(string(trimInput)), "x\u00a0"; got != want {
		t.Errorf("trimBlanks(%q) = %q, want %q", string(trimInput), got, want)
	}
}

func TestDigitGrammar(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"ascii.txt":     "012",
		"empty.txt":     "",
		"fullwidth.txt": "１２",
		"mixed.txt":     "12x",
	})
	ascii, ok := mem.Read("ascii.txt")
	if !ok {
		t.Fatal("ASCII digits fixture is missing")
	}
	if got := isDigits(string(ascii)); !got {
		t.Errorf("isDigits(%q) = false, want true", string(ascii))
	}

	empty, ok := mem.Read("empty.txt")
	if !ok {
		t.Fatal("empty fixture is missing")
	}
	if got := isDigits(string(empty)); got {
		t.Errorf("isDigits(%q) = true, want false", string(empty))
	}

	fullwidth, ok := mem.Read("fullwidth.txt")
	if !ok {
		t.Fatal("fullwidth digits fixture is missing")
	}
	if got := isDigits(string(fullwidth)); got {
		t.Errorf("isDigits(%q) = true, want false", string(fullwidth))
	}

	mixed, ok := mem.Read("mixed.txt")
	if !ok {
		t.Fatal("mixed digits fixture is missing")
	}
	if got := isDigits(string(mixed)); got {
		t.Errorf("isDigits(%q) = true, want false", string(mixed))
	}
}

func TestFoldKey(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"fold.txt": "A\t  B É",
	})
	input, ok := mem.Read("fold.txt")
	if !ok {
		t.Fatal("fold fixture is missing")
	}
	if got, want := foldKey(string(input)), "a b É"; got != want {
		t.Errorf("foldKey(%q) = %q, want %q", string(input), got, want)
	}
}

func TestFailureError(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"failure.txt": "failure",
	})
	input, ok := mem.Read("failure.txt")
	if !ok {
		t.Fatal("failure fixture is missing")
	}
	if got, want := Failure(string(input)).Error(), "failure"; got != want {
		t.Errorf("Failure(%q).Error() = %q, want %q", string(input), got, want)
	}
}

func TestMemIoHierarchyAndAliases(t *testing.T) {
	mem := NewMemIo(
		map[string]string{
			"d/a/SKILL.md": "x",
			"d/f.txt":      "file",
			"empty/":       "",
		},
		"d/link",
	)

	root, ok := mem.List(".")
	if !ok {
		t.Fatal("root folder is not listable")
	}
	if want := []Entry{{Name: "d", Kind: "dir"}, {Name: "empty", Kind: "dir"}}; !entriesEqual(root, want) {
		t.Errorf("List(.) = %#v, want %#v", root, want)
	}

	children, ok := mem.List("d")
	if !ok {
		t.Fatal("d folder is not listable")
	}
	if want := []Entry{{Name: "a", Kind: "dir"}, {Name: "f.txt", Kind: "file"}, {Name: "link", Kind: "other"}}; !entriesEqual(children, want) {
		t.Errorf("List(d) = %#v, want %#v", children, want)
	}

	alias, ok := mem.List("./d/")
	if !ok {
		t.Fatal("./d/ alias is not listable")
	}
	if !entriesEqual(alias, children) {
		t.Errorf("List(./d/) = %#v, want %#v", alias, children)
	}

	empty, ok := mem.List("empty")
	if !ok {
		t.Fatal("explicit empty folder is not listable")
	}
	if len(empty) != 0 {
		t.Errorf("List(empty) = %#v, want no entries", empty)
	}

	data, ok := mem.Read("d/a/SKILL.md")
	if !ok {
		t.Fatal("file is not readable")
	}
	if got, want := string(data), "x"; got != want {
		t.Errorf("Read(d/a/SKILL.md) = %q, want %q", got, want)
	}
}

func TestMemIoRootSlashAlias(t *testing.T) {
	mem := NewMemIo(map[string]string{})
	entries, ok := mem.List("./")
	if !ok {
		t.Fatal(`List("./") did not treat the root alias as a directory`)
	}
	if entries == nil || len(entries) != 0 {
		t.Errorf(`List("./") = %#v, want a non-nil empty slice`, entries)
	}
	if _, ok := mem.List(""); ok {
		t.Error(`List("") treated an empty path as the root alias`)
	}
}

func TestMemIoUnreadableAndNonDirectories(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"d/a/SKILL.md": "content",
	})
	mem.Unreadable["d/a/SKILL.md"] = true

	if _, ok := mem.Read("d/a/SKILL.md"); ok {
		t.Error("Read returned true for an unreadable file")
	}
	if _, ok := mem.Read("d/a/missing.txt"); ok {
		t.Error("Read returned true for a missing file")
	}
	if _, ok := mem.List("absent"); ok {
		t.Error("List returned true for an absent folder")
	}
	if _, ok := mem.List("d/a/SKILL.md"); ok {
		t.Error("List returned true for a file")
	}
	if _, ok := mem.List(""); ok {
		t.Error("List returned true for the empty path")
	}
}

func TestMemIoOutputAccumulation(t *testing.T) {
	mem := NewMemIo(map[string]string{})
	mem.Out("a")
	mem.Out("b")
	mem.Err("c")

	if got, want := mem.Stdout, "ab"; got != want {
		t.Errorf("Stdout = %q, want %q", got, want)
	}
	if got, want := mem.Stderr, "c"; got != want {
		t.Errorf("Stderr = %q, want %q", got, want)
	}
}

func entriesEqual(got, want []Entry) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestNormalizeByteBoundaries(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"nul.bin":          string([]byte{0x00}),
		"continuation.bin": string([]byte{0x80}),
		"overlong.bin":     string([]byte{0xc0, 0xaf}),
		"surrogate.bin":    string([]byte{0xed, 0xa0, 0x80}),
		"above-max.bin":    string([]byte{0xf4, 0x90, 0x80, 0x80}),
		"truncated.bin":    string([]byte{0xe2, 0x82}),
		"one-bom.txt":      "\ufeffx",
		"two-bom.txt":      "\ufeff\ufeffx",
		"line-endings.txt": "a\r\nb\rc\n",
	})
	cases := []struct {
		path  string
		want  string
		valid bool
	}{
		{path: "nul.bin", want: "", valid: false},
		{path: "continuation.bin", want: "", valid: false},
		{path: "overlong.bin", want: "", valid: false},
		{path: "surrogate.bin", want: "", valid: false},
		{path: "above-max.bin", want: "", valid: false},
		{path: "truncated.bin", want: "", valid: false},
		{path: "one-bom.txt", want: "x", valid: true},
		{path: "two-bom.txt", want: "\ufeffx", valid: true},
		{path: "line-endings.txt", want: "a\nb\rc\n", valid: true},
	}
	for _, test := range cases {
		data, ok := mem.Read(test.path)
		if !ok {
			t.Fatalf("%s fixture is missing", test.path)
		}
		got, valid := normalize(data)
		if got != test.want || valid != test.valid {
			t.Errorf("normalize(%s) = (%q, %t), want (%q, %t)", test.path, got, valid, test.want, test.valid)
		}
	}
}

func TestSplitLinesBoundaries(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"empty.txt":      "",
		"final-lf.txt":   "a\n",
		"blank-line.txt": "a\n\n",
		"lone-cr.txt":    "a\rb",
	})
	cases := []struct {
		path string
		want []string
	}{
		{path: "empty.txt", want: []string{}},
		{path: "final-lf.txt", want: []string{"a"}},
		{path: "blank-line.txt", want: []string{"a", ""}},
		{path: "lone-cr.txt", want: []string{"a\rb"}},
	}
	for _, test := range cases {
		data, ok := mem.Read(test.path)
		if !ok {
			t.Fatalf("%s fixture is missing", test.path)
		}
		got := splitLines(string(data))
		if len(got) != len(test.want) {
			t.Errorf("splitLines(%s) has %d lines, want %d", test.path, len(got), len(test.want))
			continue
		}
		for i := range test.want {
			if got[i] != test.want[i] {
				t.Errorf("splitLines(%s)[%d] = %q, want %q", test.path, i, got[i], test.want[i])
			}
		}
	}
}

func TestTableUnicodeWidthsAndControls(t *testing.T) {
	mem := NewMemIo(map[string]string{
		"header-a.txt":     "A",
		"header-b.txt":     "B",
		"emoji.txt":        "😀",
		"x.txt":            "x",
		"aa.txt":           "aa",
		"empty.txt":        "",
		"control-head.txt": "X",
		"control-cell.txt": "a\t\x7f",
	})
	headerA, ok := mem.Read("header-a.txt")
	if !ok {
		t.Fatal("first header fixture is missing")
	}
	headerB, ok := mem.Read("header-b.txt")
	if !ok {
		t.Fatal("second header fixture is missing")
	}
	emoji, ok := mem.Read("emoji.txt")
	if !ok {
		t.Fatal("emoji fixture is missing")
	}
	x, ok := mem.Read("x.txt")
	if !ok {
		t.Fatal("x fixture is missing")
	}
	aa, ok := mem.Read("aa.txt")
	if !ok {
		t.Fatal("double-a fixture is missing")
	}
	empty, ok := mem.Read("empty.txt")
	if !ok {
		t.Fatal("empty fixture is missing")
	}
	if got, want := table([]string{string(headerA), string(headerB)}, [][]string{{string(emoji), string(x)}, {string(aa), string(empty)}}), "A   B\n😀   x\naa\n"; got != want {
		t.Errorf("table with Unicode cells = %q, want %q", got, want)
	}

	controlHeader, ok := mem.Read("control-head.txt")
	if !ok {
		t.Fatal("control header fixture is missing")
	}
	controlCell, ok := mem.Read("control-cell.txt")
	if !ok {
		t.Fatal("control cell fixture is missing")
	}
	if got, want := table([]string{string(controlHeader)}, [][]string{{string(controlCell)}}), "X\na??\n"; got != want {
		t.Errorf("table with controls = %q, want %q", got, want)
	}
}
