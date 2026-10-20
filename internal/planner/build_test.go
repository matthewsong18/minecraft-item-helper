package planner_test

import (
	"minecraft-item-helper/internal/planner"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildSharesCraftingLeftoversWithoutInitialInventory(t *testing.T) {
	got, err := planner.PlanBuild(planner.BuildRequest{
		Items:   map[string]uint{"planks": 1, "stick": 4},
		Recipes: buildRecipes(),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := planner.Response{
		FlatMap: map[string]uint{"log": 1}, Inventory: planner.Inventory{"planks": 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestBuildReservesOwnedFinishedItems(t *testing.T) {
	stock := planner.Inventory{"planks": 1}
	got, err := planner.PlanBuild(planner.BuildRequest{
		Items:   map[string]uint{"planks": 1, "stick": 4},
		Recipes: buildRecipes(), Inventory: stock,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := planner.Response{FlatMap: map[string]uint{"log": 1}, Inventory: planner.Inventory{"planks": 2}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
	if diff := cmp.Diff(planner.Inventory{"planks": 1}, stock); diff != "" {
		t.Fatalf("input inventory changed: %s", diff)
	}
}

func buildRecipes() map[string]planner.Recipe {
	return map[string]planner.Recipe{
		"planks": {Yield: 4, Inputs: inputs(ingredient("log", 1))},
		"stick":  {Yield: 4, Inputs: inputs(ingredient("planks", 2))},
	}
}
