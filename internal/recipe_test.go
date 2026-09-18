package recipe_test

import (
	recipe "minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveBaseMaterials(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		inputs recipe.Request
		want   recipe.Response
	}{
		{
			name: "test raw materials for a direct recipe",
			inputs: recipe.Request{
				RequestedItem:   "planks",
				RequestedAmount: 4,
				Recipes: map[string]recipe.Recipe{
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
				},
			},
			want: recipe.Response{
				FlatMap: map[string]uint{"log": 1},
			},
		},
		{
			name: "test nested two-step recipe",
			inputs: recipe.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]recipe.Recipe{
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
				},
			},
			want: recipe.Response{
				FlatMap: map[string]uint{"log": 1},
			},
		},
		{
			name: "test recipe with varying levels of base materials",
			inputs: recipe.Request{
				RequestedItem:   "iron-pickaxe",
				RequestedAmount: 1,
				Recipes: map[string]recipe.Recipe{
					"iron-pickaxe": {
						Yield: 1,
						Inputs: map[string]uint{
							"iron-ingot": 3,
							"stick":      2,
						},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					}},
			},
			want: recipe.Response{
				FlatMap: map[string]uint{
					"iron-ingot": 3,
					"log":        1,
				},
			},
		},
		{
			name: "test multiplying by yield amount",
			inputs: recipe.Request{
				RequestedItem:   "planks",
				RequestedAmount: 8,
				Recipes: map[string]recipe.Recipe{
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					}},
			},
			want: recipe.Response{FlatMap: map[string]uint{"log": 2}},
		},
		{
			name: "test multiplier on varying levels of base materials",
			inputs: recipe.Request{
				RequestedItem:   "iron-pickaxe",
				RequestedAmount: 3,
				Recipes: map[string]recipe.Recipe{
					"iron-pickaxe": {
						Yield: 1,
						Inputs: map[string]uint{
							"iron-ingot": 3,
							"stick":      2,
						},
					},
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					}},
			},
			want: recipe.Response{FlatMap: map[string]uint{
				"iron-ingot": 9,
				"log":        1,
			}},
		},
		{
			name: "test adding extra items to inventory",
			inputs: recipe.Request{
				RequestedItem:   "stick",
				RequestedAmount: 3,
				Recipes: map[string]recipe.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
				},
				Inventory: map[string]uint{},
			},
			want: recipe.Response{
				FlatMap: map[string]uint{
					"log": 1,
				},
				Inventory: map[string]uint{
					"planks": 2,
					"stick":  1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recipe.ResolveBaseMaterials(tt.inputs)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ResolveBaseMaterials() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
