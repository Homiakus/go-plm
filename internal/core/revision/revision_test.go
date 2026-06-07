package revision

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"1.0", Version{1, 0}, false},
		{"v2.3", Version{2, 3}, false},
		{"0.0", Version{0, 0}, false},
		{"10.99", Version{10, 99}, false},
		{"", Version{}, true},
		{"1", Version{}, true},
		{"a.b", Version{}, true},
		{"-1.0", Version{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	v := Version{1, 0}
	if v.String() != "1.0" {
		t.Errorf("String() = %q", v.String())
	}
	if v.WithVPrefix() != "v1.0" {
		t.Errorf("WithVPrefix() = %q", v.WithVPrefix())
	}
}

func TestVersionBump(t *testing.T) {
	v := Version{1, 5}
	minor := v.BumpMinor()
	if minor.String() != "1.6" {
		t.Errorf("BumpMinor = %v", minor)
	}
	major := v.BumpMajor()
	if major.String() != "2.0" {
		t.Errorf("BumpMajor = %v", major)
	}
}

func TestVersionIsNewer(t *testing.T) {
	v10 := Version{1, 0}
	v11 := Version{1, 1}
	v20 := Version{2, 0}

	if !v11.IsNewer(v10) {
		t.Error("1.1 not newer than 1.0")
	}
	if !v20.IsNewer(v11) {
		t.Error("2.0 not newer than 1.1")
	}
	if v10.IsNewer(v11) {
		t.Error("1.0 newer than 1.1")
	}
}
