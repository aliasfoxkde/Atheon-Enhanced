package core

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ignoreRule struct {
	re      *regexp.Regexp
	negated bool
}

type ignoreMatcher struct {
	rules []ignoreRule
}

func compileIgnoreFile(path string) (*ignoreMatcher, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rules []ignoreRule
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = line[1:]
		}
		re, err := ignorePatternToRegexp(line)
		if err != nil {
			continue
		}
		rules = append(rules, ignoreRule{re: re, negated: negated})
	}
	return &ignoreMatcher{rules: rules}, sc.Err()
}

func (m *ignoreMatcher) matchesPath(path string) bool {
	path = filepath.ToSlash(path)
	ignored := false
	for _, r := range m.rules {
		if r.re.MatchString(path) {
			ignored = !r.negated
		}
	}
	return ignored
}

func ignorePatternToRegexp(pattern string) (*regexp.Regexp, error) {
	pattern = strings.TrimSuffix(pattern, "/")
	if pattern == "" {
		return nil, errors.New("empty pattern")
	}

	// anchored to root if pattern contains a slash
	anchored := strings.Contains(pattern, "/")
	pattern = strings.TrimPrefix(pattern, "/")

	var b strings.Builder
	switch {
	case strings.HasPrefix(pattern, "**/"):
		// **/foo matches foo at any depth — treat as unanchored
		b.WriteString("(?:^|.*/)")
		pattern = pattern[3:]
	case anchored:
		b.WriteString("^")
	default:
		b.WriteString("(?:^|.*/)")
	}

	// /**/  in the middle means zero or more intermediate directories
	pattern = strings.ReplaceAll(pattern, "/**/", "\x00")

	trailAll := strings.HasSuffix(pattern, "/**")
	if trailAll {
		pattern = pattern[:len(pattern)-3]
	}

	parts := strings.Split(pattern, "\x00")
	for i, part := range parts {
		if i > 0 {
			b.WriteString("/(?:.*/)?")
		}
		writeIgnoreSegment(part, &b)
	}

	if trailAll {
		b.WriteString("/.*")
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}

func writeIgnoreSegment(seg string, b *strings.Builder) {
	for i := 0; i < len(seg); i++ {
		switch {
		case seg[i] == '*' && i+1 < len(seg) && seg[i+1] == '*':
			b.WriteString(".*")
			i++
		case seg[i] == '*':
			b.WriteString("[^/]*")
		case seg[i] == '?':
			b.WriteString("[^/]")
		case seg[i] == '[':
			j := i + 1
			if j < len(seg) && seg[j] == '!' {
				j++
			}
			if j < len(seg) && seg[j] == ']' {
				j++
			}
			for j < len(seg) && seg[j] != ']' {
				j++
			}
			if j < len(seg) {
				b.WriteString(seg[i : j+1])
				i = j
			} else {
				b.WriteString(`\[`)
			}
		default:
			if strings.ContainsRune(`\.+^${}()|`, rune(seg[i])) {
				b.WriteByte('\\')
			}
			b.WriteByte(seg[i])
		}
	}
}
