// Package markdown provides backend Markdown parsing: outline extraction, link detection.
package markdown

import (
	"regexp"
	"strings"
)

// OutlineItem represents a heading in a Markdown document.
type OutlineItem struct {
	Level int
	Text  string
	Slug  string
}

var headingRE = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// ExtractOutline parses headings from Markdown content.
func ExtractOutline(md []byte) []OutlineItem {
	var items []OutlineItem
	for _, line := range strings.Split(string(md), "\n") {
		m := headingRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		level := len(m[1])
		text := strings.TrimSpace(m[2])
		slug := slugify(text)
		items = append(items, OutlineItem{Level: level, Text: text, Slug: slug})
	}
	return items
}

// ExtractLinks finds all [text](url) links in Markdown.
func ExtractLinks(md []byte) []string {
	re := regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
	var links []string
	for _, m := range re.FindAllSubmatch(md, -1) {
		links = append(links, string(m[2]))
	}
	return links
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ' || r == '\t':
			if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
				b.WriteRune('-')
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
