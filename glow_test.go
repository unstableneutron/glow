package main

import (
	"strings"
	"testing"
)

func TestRenderMermaidFlag(t *testing.T) {
	tt := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "default value",
			args:     []string{},
			expected: "plain",
		},
		{
			name:     "explicit plain",
			args:     []string{"--render-mermaid", "plain"},
			expected: "plain",
		},
		{
			name:     "ascii mode",
			args:     []string{"--render-mermaid", "ascii"},
			expected: "ascii",
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			// Reset flag to default before each test
			if err := rootCmd.Flags().Set("render-mermaid", "plain"); err != nil {
				t.Fatalf("failed to reset flag: %v", err)
			}

			err := rootCmd.ParseFlags(v.args)
			if err != nil {
				t.Fatal(err)
			}
			if renderMermaid != v.expected {
				t.Errorf("Parsing --render-mermaid failed: got %s, want %s", renderMermaid, v.expected)
			}
		})
	}
}

func TestRenderMermaidValidation(t *testing.T) {
	// Reset flag to default after test
	t.Cleanup(func() {
		_ = rootCmd.Flags().Set("render-mermaid", "plain")
	})

	// Set an invalid value
	if err := rootCmd.Flags().Set("render-mermaid", "invalid"); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}

	err := validateOptions(rootCmd)
	if err == nil {
		t.Error("expected error for invalid render-mermaid value, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid --render-mermaid value") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGlowFlags(t *testing.T) {
	tt := []struct {
		args  []string
		check func() bool
	}{
		{
			args: []string{"-p"},
			check: func() bool {
				return pager
			},
		},
		{
			args: []string{"-s", "light"},
			check: func() bool {
				return style == "light"
			},
		},
		{
			args: []string{"-w", "40"},
			check: func() bool {
				return width == 40
			},
		},
	}

	for _, v := range tt {
		err := rootCmd.ParseFlags(v.args)
		if err != nil {
			t.Fatal(err)
		}
		if !v.check() {
			t.Errorf("Parsing flag failed: %s", v.args)
		}
	}
}
