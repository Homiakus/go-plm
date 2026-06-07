package project

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("demo", "Demo Project")
	if cfg.Project.Code != "demo" {
		t.Errorf("Code = %q", cfg.Project.Code)
	}
	if cfg.Project.Title != "Demo Project" {
		t.Errorf("Title = %q", cfg.Project.Title)
	}
	if cfg.Naming.Standard != "v10" {
		t.Errorf("Naming standard = %q", cfg.Naming.Standard)
	}
	if cfg.Storage.Index != ".plm/index.db" {
		t.Errorf("Index = %q", cfg.Storage.Index)
	}
	if !cfg.Git.Enabled {
		t.Error("Git must be enabled by default")
	}
	if cfg.Modules.ShopFloor {
		t.Error("ShopFloor must be disabled by default")
	}
}

func TestDefaultConfigSequences(t *testing.T) {
	cfg := DefaultConfig("test", "Test")
	// asm should start at 99 (one below 0100, because Next() increments)
	if cfg.Sequences["asm"] != 99 {
		t.Errorf("asm sequence = %d, want 99", cfg.Sequences["asm"])
	}
	// prt should start at 0
	if cfg.Sequences["prt"] != 0 {
		t.Errorf("prt sequence = %d, want 0", cfg.Sequences["prt"])
	}
}

func TestModuleFlagsDefaults(t *testing.T) {
	cfg := DefaultConfig("x", "y")
	if !cfg.Modules.Objects {
		t.Error("Objects must be enabled")
	}
	if !cfg.Modules.BOM {
		t.Error("BOM must be enabled")
	}
	if !cfg.Modules.Release {
		t.Error("Release must be enabled")
	}
	if cfg.Modules.Manufacturing {
		t.Error("Manufacturing must be disabled")
	}
	if cfg.Modules.Procurement {
		t.Error("Procurement must be disabled")
	}
}
