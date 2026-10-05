package internal_test

import (
	"minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parseJsonRecipe(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		jsonInput string
		output    *internal.Recipe
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
			output: &internal.Recipe{
				Yield: 4,
				Inputs: map[string]uint{
					"#minecraft:oak_logs": 1,
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
			output: &internal.Recipe{
				Yield: 1,
				Inputs: map[string]uint{
					"minecraft:iron_block": 3,
					"minecraft:iron_ingot": 4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := internal.ParseJsonRecipe([]byte(tt.jsonInput))

			if diff := cmp.Diff(tt.output, got); diff != "" {
				t.Errorf("ParseJsonRecipe() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
