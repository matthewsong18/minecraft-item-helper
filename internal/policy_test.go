package internal_test

import (
	"minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStopItemsConsumeInventoryWithoutExpandingRecipe(t *testing.T) {
	got := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem: "leather", RequestedAmount: 3,
		Recipes: map[string]internal.Recipe{
			"leather": {Yield: 1, Inputs: inputs(ingredient("rabbit_hide", 4))},
		},
		Inventory: internal.Inventory{"leather": 1},
		StopItems: map[string]bool{"leather": true},
	})
	want := internal.Response{FlatMap: map[string]uint{"leather": 2}, Inventory: internal.Inventory{}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}
