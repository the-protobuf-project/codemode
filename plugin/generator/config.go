// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"os"
	"path"
	"time"

	"go.yaml.in/yaml/v3"
)

// Config is codemode.yaml: how one deployment exposes its API. Every key is
// optional. Generation settings shape the output; runtime settings are written
// to codemode/config.ts for createCodemode to read.
type Config struct {
	Version         string         `yaml:"version"`
	Include         []string       `yaml:"include"`
	Exclude         []string       `yaml:"exclude"`
	Naming          string         `yaml:"naming"`
	Approval        ApprovalConfig `yaml:"approval"`
	Discovery       string         `yaml:"discovery"`
	ImportExtension *string        `yaml:"import_extension"`
	RuntimePrefix   string         `yaml:"runtime_prefix"`
	Executor        ExecutorConfig `yaml:"executor"`
	Runtime         RuntimeConfig  `yaml:"runtime"`
	MCP             MCPConfig      `yaml:"mcp"`
}

// ApprovalConfig decides which methods pause for approval when the proto
// does not say so with codemode.v1.method.approval.
type ApprovalConfig struct {
	Destructive  string `yaml:"destructive"`    // require | allow
	Elicitation  string `yaml:"elicitation"`    // require | allow
	InferFromAIP *bool  `yaml:"infer_from_aip"` // Delete* counts as destructive
}

// ExecutorConfig limits each sandbox pass.
type ExecutorConfig struct {
	Timeout  string   `yaml:"timeout"`
	MemoryMB int      `yaml:"memory_mb"`
	Outbound string   `yaml:"outbound"` // none | allowlist
	Allow    []string `yaml:"allow"`
}

// RuntimeConfig controls the durable execution runtime.
type RuntimeConfig struct {
	MaxExecutions      int    `yaml:"max_executions"`
	PausedTTL          string `yaml:"paused_ttl"`
	MaxValueBytes      int    `yaml:"max_value_bytes"`
	ResultBudgetTokens int    `yaml:"result_budget_tokens"`
}

// MCPConfig controls the MCP surface.
type MCPConfig struct {
	Expose string `yaml:"expose"` // code | search_execute
}

// LoadConfig reads codemode.yaml from file, or returns defaults when file is
// empty.
func LoadConfig(file string) (*Config, error) {
	c := &Config{}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, c); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
	}
	return c, c.defaults()
}

func (c *Config) defaults() error {
	def := func(v *string, d string, allowed ...string) error {
		if *v == "" {
			*v = d
		}
		for _, a := range allowed {
			if *v == a {
				return nil
			}
		}
		return fmt.Errorf("codemode.yaml: %q is not one of %v", *v, allowed)
	}
	if c.RuntimePrefix == "" {
		c.RuntimePrefix = "@the-protobuf-project/"
	}
	for _, d := range []string{c.Executor.Timeout, c.Runtime.PausedTTL} {
		if _, err := Millis(d); err != nil {
			return err
		}
	}
	for _, err := range []error{
		def(&c.Naming, "lowerCamel", "lowerCamel", "snake"),
		def(&c.Approval.Destructive, "require", "require", "allow"),
		def(&c.Approval.Elicitation, "require", "require", "allow"),
		def(&c.Discovery, "progressive", "progressive", "inline"),
		def(&c.Executor.Outbound, "none", "none", "allowlist"),
		def(&c.MCP.Expose, "code", "code", "search_execute"),
	} {
		if err != nil {
			return err
		}
	}
	return nil
}

// InferFromAIP reports whether unannotated Delete methods count as destructive.
func (c *Config) InferFromAIP() bool {
	return c.Approval.InferFromAIP == nil || *c.Approval.InferFromAIP
}

// Extension is the suffix written on relative imports between generated files.
func (c *Config) Extension() string {
	if c.ImportExtension == nil {
		return ".ts"
	}
	return *c.ImportExtension
}

// Selects reports whether a service's full name passes include and exclude.
func (c *Config) Selects(fullName string) bool {
	match := func(patterns []string) bool {
		for _, p := range patterns {
			if ok, _ := path.Match(p, fullName); ok {
				return true
			}
		}
		return false
	}
	if len(c.Include) > 0 && !match(c.Include) {
		return false
	}
	return !match(c.Exclude)
}

// Millis parses a Go duration such as "30s" into milliseconds; "" is 0.
func Millis(d string) (int64, error) {
	if d == "" {
		return 0, nil
	}
	v, err := time.ParseDuration(d)
	if err != nil {
		return 0, fmt.Errorf("codemode.yaml: bad duration %q: %w", d, err)
	}
	return v.Milliseconds(), nil
}
