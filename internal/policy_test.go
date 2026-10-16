package internal_test

import (
	"errors"
	"minecraft-item-helper/internal"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStopItemsConsumeInventoryWithoutExpandingRecipe(t *testing.T) {
	got, err := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem: "leather", RequestedAmount: 3,
		Recipes: map[string]internal.Recipe{
			"leather": {Yield: 1, Inputs: inputs(ingredient("rabbit_hide", 4))},
		},
		Inventory: internal.Inventory{"leather": 1},
		StopItems: map[string]bool{"leather": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := internal.Response{FlatMap: map[string]uint{"leather": 2}, Inventory: internal.Inventory{}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestRecipeCycleReturnsDependencyPath(t *testing.T) {
	_, err := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem: "ingot", RequestedAmount: 1,
		Recipes: map[string]internal.Recipe{
			"ingot":  {Yield: 1, Inputs: inputs(ingredient("nugget", 9))},
			"nugget": {Yield: 9, Inputs: inputs(ingredient("ingot", 1))},
		},
		Inventory: internal.Inventory{},
	})
	if !errors.Is(err, internal.ErrRecipeCycle) || !strings.Contains(err.Error(), "ingot -> nugget -> ingot") {
		t.Fatalf("expected cycle with dependency path, got %v", err)
	}
}
