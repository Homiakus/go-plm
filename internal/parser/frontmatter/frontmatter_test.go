package frontmatter

import "testing"

func TestSplitBasic(t *testing.T) {
	input := `---
id: test
---

# Body
`
	doc, err := Split([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if string(doc.Frontmatter) != "id: test" {
		t.Errorf("frontmatter = %q", doc.Frontmatter)
	}
}

func TestSplitEmpty(t *testing.T) {
	input := `---
---
Body
`
	doc, err := Split([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Frontmatter) != 0 {
		t.Error("expected empty frontmatter")
	}
}

func TestSplitMissing(t *testing.T) {
	_, err := Split([]byte("no frontmatter"))
	if err == nil {
		t.Error("expected error")
	}
}

func TestJoinRoundTrip(t *testing.T) {
	fm := []byte("id: test")
	body := []byte("# Body\n")
	result := Join(fm, body)
	doc, err := Split(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(doc.Frontmatter) != "id: test" {
		t.Errorf("round-trip fm = %q", doc.Frontmatter)
	}
}
