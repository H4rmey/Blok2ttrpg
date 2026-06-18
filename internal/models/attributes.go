package models

// Attributes represents the character's descriptive attributes.
type Attributes struct {
	Name       string `yaml:"name"`
	Age        string `yaml:"age"`
	Size       string `yaml:"size"`
	Alignment  string `yaml:"alignment"`
	Backstory  string `yaml:"backstory"`
	Personality string `yaml:"personality"`
	Traits     string `yaml:"traits"`
	Appearance string `yaml:"appearance"`
	Hobbies    string `yaml:"hobbies"`
	Occupation string `yaml:"occupation"`
	Inventory  string `yaml:"inventory"`
	Quirks     string `yaml:"quirks"`
	Custom     []CustomField `yaml:"custom,omitempty"`
}

// CustomField is a user-defined attribute field.
type CustomField struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// TemporaryAttribute represents a short-term effect or condition.
type TemporaryAttribute struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Duration    string `yaml:"duration,omitempty"`
}
