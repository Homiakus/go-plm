package yamlx

import "testing"

type testObj struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

func TestDecode(t *testing.T) {
	data := []byte("id: prt-001\ntitle: Bracket\n")
	obj, err := Decode[testObj](data)
	if err != nil {
		t.Fatal(err)
	}
	if obj.ID != "prt-001" {
		t.Errorf("ID = %q", obj.ID)
	}
}

func TestEncode(t *testing.T) {
	obj := testObj{ID: "prt-001", Title: "Bracket"}
	data, err := Encode(obj)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("empty output")
	}
}

func TestRoundTrip(t *testing.T) {
	obj1 := testObj{ID: "asm-0100-v1.0", Title: "Motor Assembly"}
	data, _ := Encode(obj1)
	obj2, err := Decode[testObj](data)
	if err != nil {
		t.Fatal(err)
	}
	if obj1 != obj2 {
		t.Errorf("round-trip mismatch: %+v vs %+v", obj1, obj2)
	}
}

func TestStrictDecode(t *testing.T) {
	_, err := StrictDecode[testObj]([]byte("id: x\ntitle: y\nunknown: z\n"))
	if err == nil {
		t.Error("expected error for unknown field")
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := Decode[testObj]([]byte("{"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
