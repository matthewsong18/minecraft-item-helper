package main_test

import (
	main "minecraft-item-helper"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCalculateRecipe(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		targetID     string
		targetAmount int
		recipes      map[string]*main.Recipe
		tags         main.TagRegistry
		want         *main.Node
		wantErr      bool
	}{
		{
			name:         "simple tree",
			targetID:     "minecraft:oak_boat",
			targetAmount: 1,
			recipes: map[string]*main.Recipe{
				"minecraft:oak_boat": &main.Recipe{
					ResultID: "minecraft:oak_boat",
					Yield:    1,
					Ingredients: map[string]main.IngredientAmount{
						"minecraft:oak_planks": main.IngredientAmount{
							Ingredient: main.Item{ID: "minecraft:oak_planks"},
							Amount:     5,
						},
					},
				},
				"minecraft:oak_planks": &main.Recipe{
					ResultID: "minecraft:oak_planks",
					Yield:    4,
					Ingredients: map[string]main.IngredientAmount{
						"minecraft:oak_logs": main.IngredientAmount{
							Ingredient: main.Item{ID: "minecraft:oak_logs"},
							Amount:     1,
						},
					},
				},
			},
			tags: main.TagRegistry{},
			want: &main.Node{
				Item: "minecraft:oak_boat",
				Recipe: &main.Recipe{
					ResultID: "minecraft:oak_boat",
					Yield:    1,
					Ingredients: map[string]main.IngredientAmount{
						"minecraft:oak_planks": main.IngredientAmount{
							Ingredient: main.Item{ID: "minecraft:oak_planks"},
							Amount:     5,
						},
					},
				},
				TargetAmount:     1,
				RecipeMultiplier: 1,
				Children: []main.Node{
					{
						Item: "minecraft:oak_planks",
						Recipe: &main.Recipe{
							ResultID: "minecraft:oak_planks",
							Yield:    4,
							Ingredients: map[string]main.IngredientAmount{
								"minecraft:oak_logs": main.IngredientAmount{
									Ingredient: main.Item{ID: "minecraft:oak_logs"},
									Amount:     1,
								},
							},
						},
						TargetAmount:     5,
						RecipeMultiplier: 2,
						Children: []main.Node{
							{
								Item:             "minecraft:oak_logs",
								Recipe:           nil,
								TargetAmount:     2,
								RecipeMultiplier: 0,
								Children:         nil,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := main.CalculateRecipe(tt.targetID, tt.targetAmount, tt.recipes, tt.tags, nil)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CalculateRecipe() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CalculateRecipe() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CalculateRecipe() mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
