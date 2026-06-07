package naming

import (
	"testing"
)

func TestParseValid(t *testing.T) {
	tests := []struct {
		id    string
		want  ParsedID
	}{
		{"a320-prt-0001-v1.0", ParsedID{"a320", "prt", "0001", 1, 0}},
		{"demo-asm-0100-v2.3", ParsedID{"demo", "asm", "0100", 2, 3}},
		{"x-drw-9999-v10.99", ParsedID{"x", "drw", "9999", 10, 99}},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got, err := Parse(tt.id)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.id, got, tt.want)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	invalid := []string{
		"",
		"prt-0001-v1.0",          // no project
		"a320-0001-v1.0",         // no class
		"a320-prt-v1.0",          // no sequence
		"a320-PRT-0001-v1.0",     // uppercase
		"a320_prt_0001_v1_0",     // underscores
		"a320-prt-0001",           // no version
	}

	for _, id := range invalid {
		t.Run(id, func(t *testing.T) {
			if err := Validate(id); err == nil {
				t.Errorf("expected error for %q", id)
			}
		})
	}
}

func TestGeneratorNextID(t *testing.T) {
	store := NewMemorySequenceStore(DefaultStarts())
	g := NewGenerator("demo", store)

	id1 := g.NextID("prt")
	if id1 != "demo-prt-0001-v1.0" {
		t.Errorf("first prt = %q", id1)
	}

	id2 := g.NextID("prt")
	if id2 != "demo-prt-0002-v1.0" {
		t.Errorf("second prt = %q", id2)
	}

	id3 := g.NextID("asm")
	if id3 != "demo-asm-0100-v1.0" {
		t.Errorf("first asm = %q", id3)
	}
}

func TestParsedIDString(t *testing.T) {
	p := ParsedID{"demo", "prt", "0001", 1, 0}
	if p.String() != "demo-prt-0001-v1.0" {
		t.Errorf("String() = %q", p.String())
	}
}

func TestMemorySequenceStoreConcurrency(t *testing.T) {
	store := NewMemorySequenceStore(DefaultStarts())
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			store.Next("prt")
			done <- true
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	if store.Current("prt") != 100 {
		t.Errorf("expected 100, got %d", store.Current("prt"))
	}
}
