package main

import (
	"encoding/json"
	"fmt"
	"minecraft-item-helper/internal/planner"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type capturedPlan struct {
	deferred int
	messages []string
	token    string
}

func (c *capturedPlan) DeferPrivate() error { c.deferred++; return nil }
func (c *capturedPlan) SendPrivate(s string, _, _ int) error {
	c.messages = append(c.messages, s)
	return nil
}
func (c *capturedPlan) SetToken(s string) { c.token = s }
func TestPlanUsesPrivateSenderAndRejectsBadInputs(t *testing.T) {
	svc := planService{recipes: map[string]planner.Recipe{}, tags: planner.Tags{}}
	for _, tc := range []struct {
		item     string
		quantity int64
	}{{"#minecraft:logs", 2}, {"minecraft:oak_log", 0}} {
		c := &capturedPlan{}
		if err := svc.handle(tc.item, tc.quantity, "user", c); err != nil {
			t.Fatal(err)
		}
		if c.deferred != 1 || len(c.messages) != 1 || !strings.Contains(c.messages[0], "concrete item") {
			t.Fatalf("unexpected private response: %#v", c)
		}
	}
}
func TestPlanReturnsMaterialsAndUsefulPrivatePlanningErrors(t *testing.T) {
	svc := planService{recipes: map[string]planner.Recipe{}, tags: planner.Tags{}}
	c := &capturedPlan{}
	if err := svc.handle("minecraft:oak_log", 3, "user", c); err != nil {
		t.Fatal(err)
	}
	if c.deferred != 1 || len(c.messages) != 1 || !strings.Contains(c.messages[0], "minecraft:oak_log: 3") {
		t.Fatalf("unexpected response: %#v", c)
	}
	svc.recipes = map[string]planner.Recipe{"minecraft:oak_log": {Yield: 1, Inputs: []planner.Ingredient{{Options: []string{"minecraft:oak_log"}, Count: 1}}}}
	c = &capturedPlan{}
	if err := svc.handle("minecraft:oak_log", 1, "user", c); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(c.messages[0], "recipe cycle") {
		t.Fatalf("expected a useful private planning error, got %q", c.messages[0])
	}
}
func TestLongPlanCreatesOnePrivatePageWithControlsToken(t *testing.T) {
	inputs := make([]planner.Ingredient, 0, 100)
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("minecraft:item_%03d", i)
		inputs = append(inputs, planner.Ingredient{Options: []string{id}, Count: 1})
	}
	recipes := map[string]planner.Recipe{"minecraft:root": {Yield: 1, Inputs: inputs}}
	svc := planService{recipes: recipes, tags: planner.Tags{}, newPager: func(_ string, pages []string) (string, int, error) {
		if len(pages) < 2 {
			t.Fatalf("expected pages, got %d", len(pages))
		}
		return "opaque", 0, nil
	}}
	c := &capturedPlan{}
	if err := svc.handle("minecraft:root", 1, "owner", c); err != nil {
		t.Fatal(err)
	}
	if c.token != "opaque" || len(c.messages) != 1 {
		t.Fatalf("expected initial page and token: %#v", c)
	}
}

func TestLoadedRuntimeAssetsDrivePlanCommand(t *testing.T) {
	root := t.TempDir()
	recipesDir := filepath.Join(root, "recipes")
	tagsDir := filepath.Join(root, "tags")
	if err := os.Mkdir(recipesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(tagsDir, 0755); err != nil {
		t.Fatal(err)
	}
	recipe := map[string]any{"type": "minecraft:crafting_shapeless", "ingredients": []string{"minecraft:stone", "minecraft:stone"}, "result": map[string]any{"id": "minecraft:lamp", "count": 1}}
	encoded, err := json.Marshal(recipe)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(recipesDir, "lamp.json"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MINECRAFT_ITEM_HELPER_RECIPES", recipesDir)
	t.Setenv("MINECRAFT_ITEM_HELPER_TAGS", tagsDir)
	t.Setenv("MINECRAFT_ITEM_HELPER_CONFIG", "")
	rt, err := loadRuntime()
	if err != nil {
		t.Fatal(err)
	}
	sender := &capturedPlan{}
	if err := rt.service.handle("minecraft:lamp", 1, "requester", sender); err != nil {
		t.Fatal(err)
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0], "minecraft:stone: 2") {
		t.Fatalf("shared loader/planner result missing: %#v", sender.messages)
	}
}

func TestMaterialPagesPreserveEverySortedLineWithinDiscordLimit(t *testing.T) {
	materials := map[string]uint{}
	for i := 0; i < 100; i++ {
		materials[fmt.Sprintf("minecraft:item_%03d", i)] = uint(i + 1)
	}
	pages := materialPages(materials)
	joined := strings.Join(pages, "\n")
	if len(pages) < 2 {
		t.Fatalf("expected multiple pages, got %d", len(pages))
	}
	for i := 0; i < 100; i++ {
		line := fmt.Sprintf("minecraft:item_%03d: %d", i, i+1)
		if strings.Count(joined, line) != 1 {
			t.Fatalf("expected exactly one %q in pages", line)
		}
	}
	for _, page := range pages {
		if len(page) > maxPageChars {
			t.Fatalf("page exceeds safe limit: %d", len(page))
		}
	}
}
