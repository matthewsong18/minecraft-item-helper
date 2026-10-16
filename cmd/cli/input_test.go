package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRejectsInvalidInputFiles(t *testing.T) {
	recipes, tags := cliAssets(t)
	for _, tt := range []struct{ name, flag, data, want string }{
		{"empty build", "--requests", `{}`, "build list is empty"},
		{"zero request", "--requests", `{"minecraft:oak_planks":0}`, "positive quantity"},
		{"negative stock", "--inventory", `{"minecraft:oak_log":-1}`, "parse"},
		{"fractional request", "--requests", `{"minecraft:oak_planks":1.5}`, "parse"},
		{"null requests", "--requests", `null`, "must contain an object"},
		{"null config", "--config", `null`, "settings object"},
		{"misspelled setting", "--config", `{"preferances":{}}`, "unknown field"},
		{"invalid preference", "--config", `{"preferences":{"planks":["minecraft:oak_planks"]}}`, "must be a tag"},
		{"malformed JSON", "--requests", `{`, "parse"},
		{"multiple JSON values", "--requests", `{} {}`, "parse"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.json")
			writeFile(t, path, tt.data)
			args := []string{"--recipes", recipes, "--tags", tags, tt.flag, path}
			if tt.flag != "--requests" {
				args = append(args, "--item", "minecraft:oak_planks")
			}
			var output strings.Builder
			err := run(args, &output)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
			if output.Len() != 0 {
				t.Fatalf("error wrote a partial plan: %q", output.String())
			}
		})
	}
}
