package internal_test

import (
	"minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_parseJsonRecipe(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		jsonInput string
		output    *internal.ParseJsonResponse
	}{
		{
			name: "Shapeless Recipe",
			jsonInput: `{
			  "type": "minecraft:crafting_shapeless",
			  "category": "building",
			  "group": "planks",
			  "ingredients": [
			    "#minecraft:oak_logs"
			  ],
			  "result": {
			    "count": 4,
			    "id": "minecraft:oak_planks"
			  }
			}`,
			output: &internal.ParseJsonResponse{
				Id: "minecraft:oak_planks",
				Recipe: internal.Recipe{
					Yield: 4,
					Inputs: inputs(
						ingredient("#minecraft:oak_logs", 1),
					),
				},
			},
		},
		{
			name: "Shaped Recipe",
			jsonInput: `	
				{
				  "type": "minecraft:crafting_shaped",
				  "category": "misc",
				  "key": {
				    "I": "minecraft:iron_block",
				    "i": "minecraft:iron_ingot"
				  },
				  "pattern": [
				    "III",
				    " i ",
				    "iii"
				  ],
				  "result": {
				    "count": 1,
				    "id": "minecraft:anvil"
				  }
				}
			`,
			output: &internal.ParseJsonResponse{
				Id: "minecraft:anvil",
				Recipe: internal.Recipe{
					Yield: 1,
					Inputs: inputs(
						ingredient("minecraft:iron_block", 3),
						ingredient("minecraft:iron_ingot", 4),
					),
				},
			},
		},
		{
			name: "Smelting Recipe",
			jsonInput: `
				{
					"type": "minecraft:smelting",
					"category": "food",
					"cookingtime": 200,
					"experience": 0.35,
					"ingredient": "minecraft:potato",
					"result": {
						"id": "minecraft:baked_potato"
					}
				}
			`,
			output: &internal.ParseJsonResponse{
				Id: "minecraft:baked_potato",
				Recipe: internal.Recipe{
					Yield: 1,
					Inputs: inputs(
						ingredient("minecraft:potato", 1),
					),
				},
			},
		},
		{
			name: "Shapeless Recipe with array in ingredients",
			jsonInput: `
				{
				  "type": "minecraft:crafting_shapeless",
				  "category": "misc",
				  "group": "bed_dye",
				  "ingredients": [
				    "minecraft:red_dye",
				    [
				      "minecraft:black_bed",
				      "minecraft:blue_bed",
				      "minecraft:brown_bed",
				      "minecraft:cyan_bed",
				      "minecraft:gray_bed",
				      "minecraft:green_bed",
				      "minecraft:light_blue_bed",
				      "minecraft:light_gray_bed",
				      "minecraft:lime_bed",
				      "minecraft:magenta_bed",
				      "minecraft:orange_bed",
				      "minecraft:pink_bed",
				      "minecraft:purple_bed",
				      "minecraft:yellow_bed",
				      "minecraft:white_bed"
				    ]
				  ],
				  "result": {
				    "count": 1,
				    "id": "minecraft:red_bed"
				  }
				}
			`,
			output: &internal.ParseJsonResponse{
				Id: "minecraft:red_bed",
				Recipe: internal.Recipe{
					Yield: 1,
					Inputs: inputs(
						ingredient("minecraft:red_dye", 1),
						internal.Ingredient{
							Options: []string{
								"minecraft:black_bed",
								"minecraft:blue_bed",
								"minecraft:brown_bed",
								"minecraft:cyan_bed",
								"minecraft:gray_bed",
								"minecraft:green_bed",
								"minecraft:light_blue_bed",
								"minecraft:light_gray_bed",
								"minecraft:lime_bed",
								"minecraft:magenta_bed",
								"minecraft:orange_bed",
								"minecraft:pink_bed",
								"minecraft:purple_bed",
								"minecraft:yellow_bed",
								"minecraft:white_bed",
							},
							Count: 1,
						},
					),
				},
			},
		},
		{
			name: "Shaped Recipe with array in key",
			jsonInput: `
				{
				  "type": "minecraft:crafting_shaped",
				  "category": "misc",
				  "key": {
				    "#": "minecraft:stick",
				    "X": [
				      "minecraft:coal",
				      "minecraft:charcoal"
				    ]
				  },
				  "pattern": [
				    "X",
				    "#"
				  ],
				  "result": {
				    "count": 4,
				    "id": "minecraft:torch"
				  }
				}
			`,
			output: &internal.ParseJsonResponse{
				Id: "minecraft:torch",
				Recipe: internal.Recipe{
					Yield: 4,
					Inputs: inputs(
						ingredient("minecraft:stick", 1),
						internal.Ingredient{
							Options: []string{
								"minecraft:coal",
								"minecraft:charcoal",
							},
							Count: 1,
						},
					),
				},
			},
		},
		{
			name: "Smelting Recipe with array in ingredient",
			jsonInput: `
				{
				  "type": "minecraft:smelting",
				  "category": "misc",
				  "cookingtime": 200,
				  "experience": 0.1,
				  "ingredient": [
				    "minecraft:copper_pickaxe",
				    "minecraft:copper_shovel",
				    "minecraft:copper_axe",
				    "minecraft:copper_hoe",
				    "minecraft:copper_sword",
				    "minecraft:copper_spear",
				    "minecraft:copper_helmet",
				    "minecraft:copper_chestplate",
				    "minecraft:copper_leggings",
				    "minecraft:copper_boots",
				    "minecraft:copper_horse_armor",
				    "minecraft:copper_nautilus_armor"
				  ],
				  "result": {
				    "id": "minecraft:copper_nugget"
				  }
				}
			`,
			output: &internal.ParseJsonResponse{
				Id: "minecraft:copper_nugget",
				Recipe: internal.Recipe{
					Yield: 1,
					Inputs: []internal.Ingredient{
						{
							Options: []string{
								"minecraft:copper_pickaxe",
								"minecraft:copper_shovel",
								"minecraft:copper_axe",
								"minecraft:copper_hoe",
								"minecraft:copper_sword",
								"minecraft:copper_spear",
								"minecraft:copper_helmet",
								"minecraft:copper_chestplate",
								"minecraft:copper_leggings",
								"minecraft:copper_boots",
								"minecraft:copper_horse_armor",
								"minecraft:copper_nautilus_armor",
							},
							Count: 1,
						},
					},
				},
			},
		},
	}
	lessIngredient := func(a, b internal.Ingredient) bool {
		if len(a.Options) == 0 || len(b.Options) == 0 {
			return len(a.Options) < len(b.Options)
		}
		return a.Options[0] < b.Options[0]
	}

	lessString := func(a, b string) bool {
		return a < b
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := internal.ParseJsonRecipe([]byte(tt.jsonInput))

			diff := cmp.Diff(tt.output, got,
				cmpopts.SortSlices(lessIngredient),
				cmpopts.SortSlices(lessString),
			)
			if diff != "" {
				t.Errorf("ParseJsonRecipe() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
