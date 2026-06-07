// Package frontmatter splits and joins YAML frontmatter and Markdown body.
package frontmatter

import (
	"errors"
	"strings"
)

// Document holds separated frontmatter and body.
type Document struct {
	Frontmatter []byte
	Body        []byte
}

// Split separates a Markdown document into YAML frontmatter and body.
func Split(src []byte) (Document, error) {
	text := string(src)
	if !strings.HasPrefix(text, "---") {
		return Document{}, errors.New("frontmatter: missing opening delimiter")
	}

	rest := text[3:]
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}

	// Empty frontmatter
	if strings.HasPrefix(rest, "---") {
		bd := rest[3:]
		if strings.HasPrefix(bd, "\n") {
			bd = bd[1:]
		}
		return Document{Frontmatter: []byte(""), Body: []byte(bd)}, nil
	}

	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return Document{}, errors.New("frontmatter: missing closing delimiter")
	}

	fm := rest[:idx]
	bd := rest[idx+4:]
	return Document{Frontmatter: []byte(fm), Body: []byte(bd)}, nil
}

// Join assembles frontmatter and body into a complete Markdown document.
func Join(fm, body []byte) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("\n---")
	if len(body) > 0 {
		b.WriteString("\n")
		b.Write(body)
	}
	b.WriteString("\n")
	return []byte(b.String())
}
