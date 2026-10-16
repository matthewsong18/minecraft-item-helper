package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIResolvesSingleRequestUsingTags(t *testing.T) {
	recipesDir, tagsDir := cliAssets(t)

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

func TestCLIUsesInventoryWithoutChangingFile(t *testing.T) {
	recipes, tags := cliAssets(t)
	path := filepath.Join(t.TempDir(), "inventory.json")
	const stock = `{"minecraft:oak_planks":4,"minecraft:oak_log":1}`
	writeFile(t, path, stock)
	var output strings.Builder
	err := run([]string{"--item", "minecraft:oak_planks", "--amount", "12",
		"--recipes", recipes, "--tags", tags, "--inventory", path}, &output)
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != "minecraft:oak_log: 1\n" {
		t.Fatalf("unexpected shopping list: %q", output.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != stock {
		t.Fatal("inventory file changed")
	}
}

func cliAssets(t *testing.T) (string, string) {
	t.Helper()
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

	return recipesDir, tagsDir
}
