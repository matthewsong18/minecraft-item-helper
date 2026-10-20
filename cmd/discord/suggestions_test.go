package main

import (
	"fmt"
	"minecraft-item-helper/internal/planner"
	"reflect"
	"strings"
	"testing"
)

func TestSuggestionsIncludeRecipeInputsAndNestedTagsWithCanonicalValues(t *testing.T) {
	recipes := map[string]planner.Recipe{"minecraft:oak_planks": {Inputs: []planner.Ingredient{{Options: []string{"#minecraft:logs"}}}}}
	tags := planner.Tags{"#minecraft:logs": {"#minecraft:overworld_logs"}, "#minecraft:overworld_logs": {"minecraft:oak_log", "minecraft:birch_log"}}
	got := suggestions(recipes, tags, "oak")
	want := []itemSuggestion{{Name: "Oak Log (minecraft:oak_log)", Value: "minecraft:oak_log"}, {Name: "Oak Planks (minecraft:oak_planks)", Value: "minecraft:oak_planks"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
func TestSuggestionsAreDeterministicAndLimitedToDiscordMaximum(t *testing.T) {
	recipes := map[string]planner.Recipe{}
	for i := 0; i < 40; i++ {
		recipes[fmt.Sprintf("minecraft:item_%02d", i)] = planner.Recipe{}
	}
	first := suggestions(recipes, nil, "")
	second := suggestions(recipes, nil, "")
	if len(first) != 25 || !reflect.DeepEqual(first, second) {
		t.Fatalf("suggestions not capped and stable: %d", len(first))
	}
	if got := suggestions(recipes, nil, "not-found"); len(got) != 0 {
		t.Fatalf("expected no suggestions, got %#v", got)
	}
}

func TestSelectedSuggestionUsesCanonicalIdentifierInPlanning(t *testing.T) {
	recipes := map[string]planner.Recipe{"minecraft:oak_planks": {Yield: 1, Inputs: []planner.Ingredient{{Options: []string{"minecraft:oak_log"}, Count: 1}}}}
	match := suggestions(recipes, nil, "oak_log")[0]
	service := planService{recipes: recipes, tags: planner.Tags{}}
	sender := &capturedPlan{}
	if err := service.handle(match.Value, 2, "requester", sender); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sender.messages[0], "minecraft:oak_log: 2") {
		t.Fatalf("selected value was not used exactly: %#v", sender.messages)
	}
}

func TestDuplicateReadableNamesRetainDistinctCanonicalIDs(t *testing.T) {
	got := suggestions(map[string]planner.Recipe{"minecraft:oak_log": {}, "mod:oak_log": {}}, nil, "oak log")
	if len(got) != 2 || got[0].Value == got[1].Value || got[0].Name == got[1].Name {
		t.Fatalf("duplicate names should remain distinct: %#v", got)
	}
}
