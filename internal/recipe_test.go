package recipe_test

import (
	recipe "minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveBaseMaterials(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		recipes map[string]recipe.Recipe
		inputs  map[string]uint
		want    map[string]uint
	}{
		{
			name: "test raw materials for a direct recipe",
			recipes: map[string]recipe.Recipe{
				"planks": {
					Yield:  4,
					Inputs: map[string]uint{"log": 1},
				},
			},
			inputs: map[string]uint{"planks": 4},
			want:   map[string]uint{"log": 1},
		},
		{
			name: "test nested two-step recipe",
			recipes: map[string]recipe.Recipe{
				"planks": {
					Yield:  4,
					Inputs: map[string]uint{"log": 1},
				},
				"stick": {
					Yield:  4,
					Inputs: map[string]uint{"planks": 2},
				},
			},
			inputs: map[string]uint{"stick": 4},
			want:   map[string]uint{"log": 1},
		},
		{
			name: "test recipe with varying levels of base materials",
			recipes: map[string]recipe.Recipe{
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
				},
			},
			inputs: map[string]uint{"iron-pickaxe": 1},
			want: map[string]uint{
				"iron-ingot": 3,
				"log":        1,
			},
		},
		{
			name: "test multiplying by yield amount",
			recipes: map[string]recipe.Recipe{
				"planks": {
					Yield:  4,
					Inputs: map[string]uint{"log": 1},
				},
			},
			inputs: map[string]uint{"planks": 8},
			want:   map[string]uint{"log": 2},
		},
		{
			name: "test multiplier on varying levels of base materials",
			recipes: map[string]recipe.Recipe{
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
				},
			},
			inputs: map[string]uint{"iron-pickaxe": 3},
			want: map[string]uint{
				"iron-ingot": 9,
				"log":        1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for item, amount := range tt.inputs {
				got := recipe.ResolveBaseMaterials(recipe.Request{
					RequestedItem:   item,
					RequestedAmount: amount,
					Recipes:         tt.recipes,
				})

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ResolveBaseMaterials() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
