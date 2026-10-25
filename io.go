package main

import (
	"sort"
	"strings"
)

type Entry struct {
	Name string
	Kind string
}

type Io interface {
	List(dir string) ([]Entry, bool)
	Read(path string) ([]byte, bool)
	Out(text string)
	Err(text string)
}

type MemIo struct {
	Stdout     string
	Stderr     string
	Unreadable map[string]bool

	files  map[string][]byte
	dirs   map[string]bool
	others map[string]bool
}

func NewMemIo(files map[string]string, others ...string) *MemIo {
	mem := &MemIo{
		Unreadable: make(map[string]bool),
		files:      make(map[string][]byte),
		dirs:       map[string]bool{".": true},
		others:     make(map[string]bool),
	}

	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := memPath(key)
		if path == "" {
			continue
		}
		if strings.HasSuffix(key, "/") {
			mem.addDirectory(path)
			continue
		}
		mem.files[path] = []byte(files[key])
		mem.addParents(path)
	}

	for _, name := range others {
		path := memPath(name)
		if path == "" {
			continue
		}
		mem.others[path] = true
		mem.addParents(path)
	}
	return mem
}

func memPath(path string) string {
	nonempty := path != ""
	for strings.HasPrefix(path, "./") {
		path = path[2:]
	}
	for strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	if path == "" && nonempty {
		return "."
	}
	return path
}

func (m *MemIo) addDirectory(path string) {
	m.dirs[path] = true
	m.addParents(path)
}

func (m *MemIo) addParents(path string) {
	prefix := ""
	for {
		separator := strings.Index(path, "/")
		if separator < 0 {
			return
		}
		part := path[:separator]
		if prefix == "" {
			prefix = part
		} else {
			prefix += "/" + part
		}
		if prefix == "" {
			prefix = "/"
		}
		m.dirs[prefix] = true
		path = path[separator+1:]
	}
}

func memParent(path string) string {
	separator := strings.LastIndex(path, "/")
	if separator < 0 {
		return "."
	}
	if separator == 0 {
		return "/"
	}
	return path[:separator]
}

func (m *MemIo) List(dir string) ([]Entry, bool) {
	dir = memPath(dir)
	if dir == "" || !m.dirs[dir] {
		return nil, false
	}

	entriesByName := make(map[string]Entry)
	add := func(path, kind string) {
		if path == "." || memParent(path) != dir {
			return
		}
		name := path
		if separator := strings.LastIndex(path, "/"); separator >= 0 {
			name = path[separator+1:]
		}
		existing, ok := entriesByName[name]
		if !ok || entryPriority(kind) > entryPriority(existing.Kind) {
			entriesByName[name] = Entry{Name: name, Kind: kind}
		}
	}

	for path := range m.files {
		add(path, "file")
	}
	for path := range m.others {
		add(path, "other")
	}
	for path := range m.dirs {
		add(path, "dir")
	}

	names := make([]string, 0, len(entriesByName))
	for name := range entriesByName {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]Entry, 0, len(names))
	for _, name := range names {
		entries = append(entries, entriesByName[name])
	}
	return entries, true
}

func entryPriority(kind string) int {
	switch kind {
	case "dir":
		return 3
	case "other":
		return 2
	default:
		return 1
	}
}

func (m *MemIo) Read(path string) ([]byte, bool) {
	path = memPath(path)
	if m.Unreadable != nil {
		for unreadable := range m.Unreadable {
			if memPath(unreadable) == path {
				return nil, false
			}
		}
	}
	data, ok := m.files[path]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), data...), true
}

func (m *MemIo) Out(text string) {
	m.Stdout += text
}

func (m *MemIo) Err(text string) {
	m.Stderr += text
}
