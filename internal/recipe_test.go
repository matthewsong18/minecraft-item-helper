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
		inputs  map[string]int
		want    map[string]int
	}{
		{
			name: "test raw materials for a direct recipe",
			recipes: map[string]recipe.Recipe{
				"planks": {
					Yield:  4,
					Inputs: map[string]int{"log": 1},
				},
			},
			inputs: map[string]int{"planks": 4},
			want:   map[string]int{"log": 1},
		},
		{
			name: "test nested two-step recipe",
			recipes: map[string]recipe.Recipe{
				"planks": {
					Yield:  4,
					Inputs: map[string]int{"log": 1},
				},
				"stick": {
					Yield:  4,
					Inputs: map[string]int{"planks": 2},
				},
			},
			inputs: map[string]int{"stick": 4},
			want:   map[string]int{"log": 1},
		},
		{
			name: "test recipe with varying levels of base materials",
			recipes: map[string]recipe.Recipe{
				"iron-pickaxe": {
					Yield: 1,
					Inputs: map[string]int{
						"iron-ingot": 3,
						"stick":      2,
					},
				},
				"planks": {
					Yield:  4,
					Inputs: map[string]int{"log": 1},
				},
				"stick": {
					Yield:  4,
					Inputs: map[string]int{"planks": 2},
				},
			},
			inputs: map[string]int{"iron-pickaxe": 1},
			want: map[string]int{
				"iron-ingot": 3,
				"log":        1,
			},
		},
		{
			name: "test multiplying by yield amount",
			recipes: map[string]recipe.Recipe{
				"planks": {
					Yield:  4,
					Inputs: map[string]int{"log": 1},
				},
			},
			inputs: map[string]int{"planks": 8},
			want:   map[string]int{"log": 2},
		},
		{
			name: "test multiplier on varying levels of base materials",
			recipes: map[string]recipe.Recipe{
				"iron-pickaxe": {
					Yield: 1,
					Inputs: map[string]int{
						"iron-ingot": 3,
						"stick":      2,
					},
				},
				"stick": {
					Yield:  4,
					Inputs: map[string]int{"planks": 2},
				},
				"planks": {
					Yield:  4,
					Inputs: map[string]int{"log": 1},
				},
			},
			inputs: map[string]int{"iron-pickaxe": 3},
			want: map[string]int{
				"iron-ingot": 9,
				"log":        1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for item, amount := range tt.inputs {
				got := recipe.ResolveBaseMaterials(item, amount, tt.recipes)

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ResolveBaseMaterials() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
