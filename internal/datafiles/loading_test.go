package datafiles_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"minecraft-item-helper/internal/datafiles"
)

func TestLoadRecipesHonorsConfiguredChoiceAndDefaultFilename(t *testing.T) {
	dir := t.TempDir()
	writeJSONFile(t, filepath.Join(dir, "oak_planks.json"), `{"type":"minecraft:crafting_shapeless","ingredients":["minecraft:oak_log"],"result":{"id":"minecraft:oak_planks","count":4}}`)
	writeJSONFile(t, filepath.Join(dir, "alternate.json"), `{"type":"minecraft:crafting_shapeless","ingredients":["minecraft:bamboo"],"result":{"id":"minecraft:oak_planks","count":2}}`)
	writeJSONFile(t, filepath.Join(dir, "unsupported.json"), `{"type":"minecraft:stonecutting","ingredient":"minecraft:stone","result":{"id":"minecraft:oak_planks"}}`)

	defaultRecipes, err := datafiles.LoadRecipes(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := defaultRecipes["minecraft:oak_planks"].Inputs[0].Options[0]; got != "minecraft:oak_log" {
		t.Fatalf("default recipe input = %q, want oak log", got)
	}

	chosenRecipes, err := datafiles.LoadRecipes(dir, map[string]string{"minecraft:oak_planks": "alternate"})
	if err != nil {
		t.Fatal(err)
	}
	if got := chosenRecipes["minecraft:oak_planks"].Inputs[0].Options[0]; got != "minecraft:bamboo" {
		t.Fatalf("chosen recipe input = %q, want bamboo", got)
	}

	_, err = datafiles.LoadRecipes(dir, map[string]string{"minecraft:oak_planks": "unsupported"})
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unsupported recipe choice error = %v", err)
	}
}

func TestLoadTagsUsesNestedFilePathAsTagID(t *testing.T) {
	dir := t.TempDir()
	writeJSONFile(t, filepath.Join(dir, "logs", "oak.json"), `{"values":["minecraft:oak_log"]}`)

	tags, err := datafiles.LoadTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(tags["#minecraft:logs/oak"], ","); got != "minecraft:oak_log" {
		t.Fatalf("nested tag members = %q, want minecraft:oak_log", got)
	}
}

func TestLoadQuantitiesRejectsNullAndTagIdentifiers(t *testing.T) {
	for _, input := range []struct {
		name string
		json string
		want string
	}{
		{name: "null", json: "null", want: "object"},
		{name: "tag ID", json: `{"#minecraft:logs":1}`, want: "concrete item ID"},
		{name: "negative quantity", json: `{"minecraft:oak_log":-1}`, want: "parse"},
	} {
		t.Run(input.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "quantities.json")
			writeJSONFile(t, path, input.json)
			_, err := datafiles.LoadQuantities(path)
			if err == nil || !strings.Contains(err.Error(), input.want) {
				t.Fatalf("LoadQuantities() error = %v, want %q", err, input.want)
			}
		})
	}
}

func TestLoadSettingsAppliesStopItemOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	writeJSONFile(t, path, `{"preferences":{"#minecraft:logs":["minecraft:oak_log"]},"stop_items":{"add":["minecraft:oak_planks"],"remove":["minecraft:leather"]}}`)

	settings, err := datafiles.LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	stops := settings.StopItemSet()
	if !stops["minecraft:oak_planks"] || stops["minecraft:leather"] {
		t.Fatalf("resolved stop items = %v, want oak planks added and leather removed", stops)
	}
	if got := settings.Preferences["#minecraft:logs"][0]; got != "minecraft:oak_log" {
		t.Fatalf("loaded preference = %q, want minecraft:oak_log", got)
	}
}

func writeJSONFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
