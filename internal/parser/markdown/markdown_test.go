package markdown

import "testing"

func TestExtractOutline(t *testing.T) {
	md := []byte("# Title\n\n## Section 1\n### Sub 1.1\n## Section 2\n")
	items := ExtractOutline(md)
	if len(items) != 4 {
		t.Fatalf("expected 4 headings, got %d", len(items))
	}
	if items[0].Level != 1 || items[0].Text != "Title" {
		t.Errorf("first heading: %+v", items[0])
	}
	if items[2].Level != 3 || items[2].Text != "Sub 1.1" {
		t.Errorf("third heading: %+v", items[2])
	}
}

func TestExtractOutlineEmpty(t *testing.T) {
	items := ExtractOutline([]byte("No headings here\n"))
	if len(items) != 0 {
		t.Error("expected no headings")
	}
}

func TestExtractLinks(t *testing.T) {
	md := []byte("See [part](demo-prt-0001-v1.0) and [asm](demo-asm-0100-v1.0)")
	links := ExtractLinks(md)
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0] != "demo-prt-0001-v1.0" {
		t.Errorf("link[0] = %q", links[0])
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct{ input, want string }{
		{"Bracket Motor", "bracket-motor"},
		{"  Spaced  Out  ", "spaced-out"},
		{"Already-kebab", "already-kebab"},
		{"mix 123 numbers", "mix-123-numbers"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
