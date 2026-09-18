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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for item, amount := range tt.inputs {

				got := recipe.ResolveBaseMaterials(item, amount, tt.recipes)

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Fatalf("ResolveBaseMaterials() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
