// Package project defines the Project domain type — the top-level container for PLM data.
// Pure domain model — no filesystem, index, or Git dependencies.
package project

// Project represents the top-level PLM workspace configuration.
type Project struct {
	Code        string `yaml:"code" json:"code"`
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Version     int    `yaml:"version" json:"version"`
}

// NamingConfig holds the naming standard configuration.
type NamingConfig struct {
	Standard     string `yaml:"standard" json:"standard"`
	Pattern      string `yaml:"pattern" json:"pattern"`
	ProjectCode  string `yaml:"project_code" json:"project_code"`
	Case         string `yaml:"case" json:"case"` // kebab, snake, camel
	MaxLength    int    `yaml:"max_length_without_ext,omitempty" json:"max_length_without_ext,omitempty"`
}

// ModuleFlags enables/disables PLM modules.
type ModuleFlags struct {
	Objects          bool `yaml:"objects" json:"objects"`
	BOM              bool `yaml:"bom" json:"bom"`
	Release          bool `yaml:"release" json:"release"`
	Manufacturing    bool `yaml:"manufacturing" json:"manufacturing"`
	Procurement      bool `yaml:"procurement" json:"procurement"`
	StandardParts    bool `yaml:"standardparts" json:"standardparts"`
	ShopFloor        bool `yaml:"shopfloor" json:"shopfloor"`
	ChangeManagement bool `yaml:"change_management" json:"change_management"`
}

// StorageConfig defines where the source of truth and index live.
type StorageConfig struct {
	SourceOfTruth string `yaml:"source_of_truth" json:"source_of_truth"` // "markdown_yaml"
	Index         string `yaml:"index" json:"index"`
}

// GitConfig controls Git integration behaviour.
type GitConfig struct {
	Enabled        bool `yaml:"enabled" json:"enabled"`
	AutoCheckpoint bool `yaml:"auto_checkpoint" json:"auto_checkpoint"`
	ReleaseTags    bool `yaml:"release_tags" json:"release_tags"`
}

// Config is the full project configuration as stored in project.md (YAML frontmatter).
type Config struct {
	Project   Project      `yaml:"project" json:"project"`
	Naming    NamingConfig `yaml:"naming" json:"naming"`
	Sequences map[string]int `yaml:"sequences" json:"sequences"`
	Modules   ModuleFlags  `yaml:"modules" json:"modules"`
	Storage   StorageConfig `yaml:"storage" json:"storage"`
	Git       GitConfig    `yaml:"git" json:"git"`
}

// DefaultConfig returns a sensible default project configuration.
func DefaultConfig(code, title string) Config {
	return Config{
		Project: Project{
			Code:    code,
			Title:   title,
			Version: 1,
		},
		Naming: NamingConfig{
			Standard:    "v10",
			Pattern:     "[project]-[class]-[sequence]-v[major].[minor]",
			ProjectCode: code,
			Case:        "kebab",
			MaxLength:   64,
		},
		Sequences: map[string]int{
			"prt": 0, "asm": 99, "drw": 0, "doc": 0, "cut": 0, "bnd": 0,
			"nc": 0, "ins": 0, "tpc": 0, "wi": 0, "bom": 0, "rel": 0, "cr": 0,
		},
		Modules: ModuleFlags{
			Objects:       true,
			BOM:           true,
			Release:       true,
			StandardParts: true,
		},
		Storage: StorageConfig{
			SourceOfTruth: "markdown_yaml",
			Index:         ".plm/index.db",
		},
		Git: GitConfig{
			Enabled:     true,
			ReleaseTags: true,
		},
	}
}
