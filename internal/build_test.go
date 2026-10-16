package internal_test

import (
	"minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildSharesCraftingLeftoversWithoutInitialInventory(t *testing.T) {
	got, err := internal.PlanBuild(internal.BuildRequest{
		Items:   map[string]uint{"planks": 1, "stick": 4},
		Recipes: buildRecipes(),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := internal.Response{
		FlatMap: map[string]uint{"log": 1}, Inventory: internal.Inventory{"planks": 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestBuildReservesOwnedFinishedItems(t *testing.T) {
	stock := internal.Inventory{"planks": 1}
	got, err := internal.PlanBuild(internal.BuildRequest{
		Items:   map[string]uint{"planks": 1, "stick": 4},
		Recipes: buildRecipes(), Inventory: stock,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := internal.Response{FlatMap: map[string]uint{"log": 1}, Inventory: internal.Inventory{"planks": 2}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
	if diff := cmp.Diff(internal.Inventory{"planks": 1}, stock); diff != "" {
		t.Fatalf("input inventory changed: %s", diff)
	}
}

func buildRecipes() map[string]internal.Recipe {
	return map[string]internal.Recipe{
		"planks": {Yield: 4, Inputs: inputs(ingredient("log", 1))},
		"stick":  {Yield: 4, Inputs: inputs(ingredient("planks", 2))},
	}
}
