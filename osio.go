package main

import (
	"os"
	"unicode/utf8"
)

type OsIo struct{}

func (OsIo) List(dir string) ([]Entry, bool) {
	children, err := os.ReadDir(dir)
	if err != nil {
		return nil, false
	}

	entries := make([]Entry, 0, len(children))
	for _, child := range children {
		name := child.Name()
		if !utf8.ValidString(name) {
			continue
		}

		kind := "other"
		if child.IsDir() {
			kind = "dir"
		} else if child.Type() == 0 {
			info, infoErr := child.Info()
			if infoErr == nil {
				if info.IsDir() {
					kind = "dir"
				} else if info.Mode().IsRegular() {
					kind = "file"
				}
			}
		}
		entries = append(entries, Entry{Name: name, Kind: kind})
	}
	return entries, true
}

func (OsIo) Read(path string) ([]byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func (OsIo) Out(text string) {
	_, _ = os.Stdout.WriteString(text)
}

func (OsIo) Err(text string) {
	_, _ = os.Stderr.WriteString(text)
}
