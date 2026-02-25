package cmd

import (
	"testing"
)

func TestWalkCommands_NonEmpty(t *testing.T) {
	var commands []commandInfo
	walkCommands(rootCmd, "", &commands)
	if len(commands) == 0 {
		t.Error("walkCommands should return at least one command")
	}
}

func TestFindCommand_Known(t *testing.T) {
	tests := []string{"markets list", "markets get", "version", "commands list"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			cmd := findCommand(rootCmd, path)
			if cmd == nil {
				t.Errorf("findCommand(%q) returned nil", path)
			}
		})
	}
}

func TestFindCommand_Unknown(t *testing.T) {
	cmd := findCommand(rootCmd, "nonexistent foo")
	if cmd != nil {
		t.Error("findCommand should return nil for unknown path")
	}
}

func TestCommandRequiresAuth(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"clob create-order", true},
		{"approve check", true},
		{"ctf split", true},
		{"markets list", false},
		{"version", false},
		{"commands list", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := commandRequiresAuth(tt.path)
			if got != tt.expected {
				t.Errorf("commandRequiresAuth(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestDescribeCommand_RequiredFlag(t *testing.T) {
	cmd := findCommand(rootCmd, "clob price")
	if cmd == nil {
		t.Fatal("could not find 'clob price'")
	}
	info := describeCommand(cmd, "clob price")
	found := false
	for _, f := range info.Flags {
		if f.Name == "token" {
			found = true
			if !f.Required {
				t.Error("expected --token flag to be marked required")
			}
			break
		}
	}
	if !found {
		t.Error("expected --token flag in clob price")
	}
}

func TestDescribeCommand_OptionalFlagNotRequired(t *testing.T) {
	cmd := findCommand(rootCmd, "markets list")
	if cmd == nil {
		t.Fatal("could not find 'markets list'")
	}
	info := describeCommand(cmd, "markets list")
	for _, f := range info.Flags {
		if f.Name == "limit" {
			if f.Required {
				t.Error("expected --limit flag to NOT be marked required")
			}
			return
		}
	}
	t.Error("expected --limit flag in markets list")
}

func TestDescribeCommand_IncludesFlags(t *testing.T) {
	cmd := findCommand(rootCmd, "markets list")
	if cmd == nil {
		t.Fatal("could not find 'markets list'")
	}
	info := describeCommand(cmd, "markets list")
	if info.Name != "markets list" {
		t.Errorf("expected name 'markets list', got %q", info.Name)
	}
	if len(info.Flags) == 0 {
		t.Error("markets list should have flags")
	}
	// Check that --limit flag is present
	found := false
	for _, f := range info.Flags {
		if f.Name == "limit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected --limit flag in markets list")
	}
}
