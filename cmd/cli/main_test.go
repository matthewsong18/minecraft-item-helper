package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIResolvesSingleRequestUsingTags(t *testing.T) {
	root := t.TempDir()
	recipesDir := filepath.Join(root, "recipes")
	tagsDir := filepath.Join(root, "tags")
	if err := os.MkdirAll(recipesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tagsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(recipesDir, "oak_planks.json"), `{
		"type": "minecraft:crafting_shapeless",
		"ingredients": ["#minecraft:logs"],
		"result": {"count": 4, "id": "minecraft:oak_planks"}
	}`)
	writeFile(t, filepath.Join(tagsDir, "logs.json"), `{
		"values": ["minecraft:oak_log"]
	}`)

	var output strings.Builder
	err := run([]string{
		"--item", "minecraft:oak_planks",
		"--amount", "8",
		"--recipes", recipesDir,
		"--tags", tagsDir,
	}, &output)
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	const want = "minecraft:oak_log: 2\n"
	if output.String() != want {
		t.Errorf("run() output = %q, want %q", output.String(), want)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
