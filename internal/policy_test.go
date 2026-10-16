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

func TestIngredientChoiceUsesBuildThenStockThenPreferences(t *testing.T) {
	for _, tt := range []struct {
		name        string
		build       map[string]uint
		stock       internal.Inventory
		preferences map[string][]string
		want        string
	}{
		{"build wins over stock", map[string]uint{"oak": 1}, internal.Inventory{"birch": 5}, nil, "oak"},
		{"stock wins over preference", nil, internal.Inventory{"birch": 5}, map[string][]string{"#planks": {"cherry"}}, "birch"},
		{"preference breaks stock tie", nil, internal.Inventory{"oak": 1, "cherry": 1}, map[string][]string{"#planks": {"cherry"}}, "cherry"},
		{"saved preference", nil, nil, map[string][]string{"#planks": {"cherry"}}, "cherry"},
		{"first option", nil, nil, nil, "oak"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.ResolveBaseMaterials(internal.Request{
				RequestedItem: "#planks", RequestedAmount: 10,
				Tags:      internal.Tags{"#planks": {"oak", "birch", "cherry"}},
				Inventory: tt.stock, BuildItems: tt.build, Preferences: tt.preferences,
			})
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]uint{tt.want: 10 - tt.stock[tt.want]}
			if diff := cmp.Diff(want, got.FlatMap); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestExplicitItemIgnoresMaterialPreferences(t *testing.T) {
	got, err := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem: "oak", RequestedAmount: 1,
		Tags:        internal.Tags{"#planks": {"oak", "cherry"}},
		Preferences: map[string][]string{"#planks": {"cherry"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]uint{"oak": 1}, got.FlatMap); diff != "" {
		t.Fatal(diff)
	}
}

func TestNestedTagsResolveAndCyclesFail(t *testing.T) {
	req := internal.Request{
		RequestedItem: "#logs", RequestedAmount: 1,
		Tags: internal.Tags{"#logs": {"#oak_logs"}, "#oak_logs": {"oak_log"}},
	}
	got, err := internal.ResolveBaseMaterials(req)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]uint{"oak_log": 1}, got.FlatMap); diff != "" {
		t.Fatal(diff)
	}
	req.Tags["#oak_logs"] = []string{"#logs"}
	if _, err := internal.ResolveBaseMaterials(req); !errors.Is(err, internal.ErrTagCycle) {
		t.Fatalf("expected tag cycle, got %v", err)
	}
}

func TestAlternativeIngredientUsesInventory(t *testing.T) {
	got, err := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem: "torch", RequestedAmount: 1,
		Recipes: map[string]internal.Recipe{
			"torch": {Yield: 1, Inputs: []internal.Ingredient{{Options: []string{"coal", "charcoal"}, Count: 1}}},
		},
		Inventory: internal.Inventory{"charcoal": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.FlatMap) != 0 {
		t.Fatalf("unexpected shopping list: %v", got.FlatMap)
	}
}
