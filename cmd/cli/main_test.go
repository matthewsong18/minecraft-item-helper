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

func TestCLIPlansBuildListAndShowsProjectedInventory(t *testing.T) {
	recipes, tags := cliAssets(t)
	writeFile(t, filepath.Join(recipes, "stick.json"), `{
		"type":"minecraft:crafting_shapeless", "ingredients":["minecraft:oak_planks","minecraft:oak_planks"],
		"result":{"id":"minecraft:stick","count":4}
	}`)
	path := filepath.Join(t.TempDir(), "build.json")
	writeFile(t, path, `{"minecraft:oak_planks":1,"minecraft:stick":4}`)
	var output strings.Builder
	err := run([]string{"--requests", path, "--recipes", recipes, "--tags", tags, "--show-inventory"}, &output)
	if err != nil {
		t.Fatal(err)
	}
	const want = "minecraft:oak_log: 1\n\nProjected inventory after crafting (requested items set aside):\nminecraft:oak_planks: 1\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestCLIRejectsConflictingRequestModes(t *testing.T) {
	for _, args := range [][]string{
		{"--item", "minecraft:oak_planks", "--requests", "build.json"},
		{"--requests", "build.json", "--amount", "2"},
	} {
		var output strings.Builder
		err := run(args, &output)
		if err == nil || !strings.Contains(err.Error(), "cannot combine") {
			t.Fatalf("expected conflicting flags error for %v, got %v", args, err)
		}
	}
}

func TestCLIStopsAtLeatherByDefaultAndAllowsOverrides(t *testing.T) {
	recipes, tags := cliAssets(t)
	writeFile(t, filepath.Join(recipes, "leather.json"), `{
		"type":"minecraft:crafting_shapeless", "ingredients":["minecraft:rabbit_hide","minecraft:rabbit_hide","minecraft:rabbit_hide","minecraft:rabbit_hide"],
		"result":{"id":"minecraft:leather","count":1}
	}`)
	config := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, config, `{"stop_items":{"remove":["minecraft:leather"],"add":["minecraft:oak_planks"]}}`)
	for _, tt := range []struct{ item, config, want string }{
		{"minecraft:leather", "", "minecraft:leather: 1\n"},
		{"minecraft:leather", config, "minecraft:rabbit_hide: 4\n"},
		{"minecraft:oak_planks", config, "minecraft:oak_planks: 1\n"},
	} {
		var output strings.Builder
		args := []string{"--item", tt.item, "--recipes", recipes, "--tags", tags}
		if tt.config != "" {
			args = append(args, "--config", tt.config)
		}
		if err := run(args, &output); err != nil {
			t.Fatal(err)
		}
		if output.String() != tt.want {
			t.Fatalf("output = %q, want %q", output.String(), tt.want)
		}
	}
}

func TestCLIUsesSavedTagPreferences(t *testing.T) {
	recipes, tags := cliAssets(t)
	writeFile(t, filepath.Join(tags, "logs.json"), `{"values":["minecraft:oak_log","minecraft:cherry_log"]}`)
	config := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, config, `{"preferences":{"#minecraft:logs":["minecraft:cherry_log"]}}`)
	var output strings.Builder
	if err := run([]string{"--item", "minecraft:oak_planks", "--recipes", recipes, "--tags", tags, "--config", config}, &output); err != nil {
		t.Fatal(err)
	}
	if output.String() != "minecraft:cherry_log: 1\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestCLISelectsRecipeExplicitly(t *testing.T) {
	recipes, tags := cliAssets(t)
	writeFile(t, filepath.Join(recipes, "zz_bamboo_route.json"), `{
		"type":"minecraft:crafting_shapeless", "ingredients":["minecraft:bamboo"],
		"result":{"id":"minecraft:oak_planks","count":2}
	}`)
	config := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, config, `{"recipe_choices":{"minecraft:oak_planks":"zz_bamboo_route"}}`)
	for _, tt := range []struct{ config, want string }{
		{"", "minecraft:oak_log: 1\n"}, {config, "minecraft:bamboo: 1\n"},
	} {
		var output strings.Builder
		args := []string{"--item", "minecraft:oak_planks", "--recipes", recipes, "--tags", tags}
		if tt.config != "" {
			args = append(args, "--config", tt.config)
		}
		if err := run(args, &output); err != nil {
			t.Fatal(err)
		}
		if output.String() != tt.want {
			t.Fatalf("output = %q, want %q", output.String(), tt.want)
		}
	}
	writeFile(t, config, `{"recipe_choices":{"minecraft:oak_planks":"missing_recipe"}}`)
	var output strings.Builder
	err := run([]string{"--item", "minecraft:oak_planks", "--recipes", recipes, "--tags", tags, "--config", config}, &output)
	if err == nil || !strings.Contains(err.Error(), "missing_recipe") {
		t.Fatalf("expected invalid recipe choice, got %v", err)
	}
}

func TestCLIBundledRecipesUseGatherableStopItems(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	for _, item := range []string{"coal", "diamond", "emerald", "redstone", "lapis_lazuli", "raw_iron", "raw_gold", "raw_copper", "wheat", "slime_ball"} {
		t.Run(item, func(t *testing.T) {
			var output strings.Builder
			id := "minecraft:" + item
			err := run([]string{"--item", id, "--recipes", filepath.Join(root, "recipes"), "--tags", filepath.Join(root, "item-tag-groups")}, &output)
			if err != nil {
				t.Fatal(err)
			}
			if output.String() != id+": 1\n" {
				t.Fatalf("unexpected output: %q", output.String())
			}
		})
	}
}

func TestCLIBuildMaterialWinsOverStockAndSavedPreference(t *testing.T) {
	recipes, tags := cliAssets(t)
	writeFile(t, filepath.Join(recipes, "crafting_table.json"), `{
		"type":"minecraft:crafting_shaped","key":{"P":"#minecraft:planks"},"pattern":["PP","PP"],
		"result":{"id":"minecraft:crafting_table","count":1}
	}`)
	writeFile(t, filepath.Join(tags, "planks.json"), `{"values":["minecraft:cherry_planks","minecraft:oak_planks"]}`)
	root := t.TempDir()
	build := filepath.Join(root, "build.json")
	stock := filepath.Join(root, "inventory.json")
	config := filepath.Join(root, "config.json")
	writeFile(t, build, `{"minecraft:crafting_table":1,"minecraft:oak_planks":1}`)
	writeFile(t, stock, `{"minecraft:cherry_planks":64,"minecraft:oak_log":1}`)
	writeFile(t, config, `{"preferences":{"#minecraft:planks":["minecraft:cherry_planks"]}}`)
	var output strings.Builder
	err := run([]string{"--requests", build, "--inventory", stock, "--config", config,
		"--recipes", recipes, "--tags", tags, "--show-inventory"}, &output)
	if err != nil {
		t.Fatal(err)
	}
	const want = "minecraft:oak_log: 1\n\nProjected inventory after crafting (requested items set aside):\nminecraft:cherry_planks: 64\nminecraft:oak_planks: 3\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestCLIBundledLecternAndTorchShoppingLists(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	for _, tt := range []struct{ item, amount, want string }{
		{"minecraft:lectern", "1", "minecraft:leather: 3\nminecraft:oak_log: 3\nminecraft:sugar_cane: 9\n"},
		{"minecraft:torch", "16", "minecraft:coal: 4\nminecraft:oak_log: 1\n"},
	} {
		t.Run(tt.item, func(t *testing.T) {
			var output strings.Builder
			err := run([]string{"--item", tt.item, "--amount", tt.amount,
				"--recipes", filepath.Join(root, "recipes"), "--tags", filepath.Join(root, "item-tag-groups")}, &output)
			if err != nil {
				t.Fatal(err)
			}
			if output.String() != tt.want {
				t.Fatalf("output = %q, want %q", output.String(), tt.want)
			}
		})
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
