package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parseRecipeDTO(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		recipeData []byte
		want       RecipeDTO
		wantErr    bool
	}{
		{
			name: "simple shaped recipe",
			recipeData: []byte(`{
					  "type": "minecraft:crafting_shaped",
					  "category": "misc",
					  "group": "boat",
					  "key": {
					    "#": "minecraft:acacia_planks"
					  },
					  "pattern": [
					    "# #",
					    "###"
					  ],
					  "result": {
					    "count": 1,
					    "id": "minecraft:acacia_boat"
					  }
			}`),
			want: CraftingShapedRecipeDTO{
				Type: "minecraft:crafting_shaped",
				Pattern: []string{
					"# #",
					"###",
				},
				Key: map[string]string{
					"#": "minecraft:acacia_planks",
				},
				Result: RecipeResultDTO{
					ID:    "minecraft:acacia_boat",
					Count: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "tag shaped recipe",
			recipeData: []byte(`{
				"type": "minecraft:crafting_shaped",
				"category": "misc",
				"key": {
				"P": "#minecraft:planks",
				"S": "#minecraft:wooden_slabs"
				},
				"pattern": [
				"PSP",
				"P P",
				"PSP"
				],
				"result": {
				"count": 1,
				"id": "minecraft:barrel"
				}
			}`),
			want: CraftingShapedRecipeDTO{
				Type: "minecraft:crafting_shaped",
				Pattern: []string{
					"PSP",
					"P P",
					"PSP",
				},
				Key: map[string]string{
					"P": "#minecraft:planks",
					"S": "#minecraft:wooden_slabs",
				},
				Result: RecipeResultDTO{
					ID:    "minecraft:barrel",
					Count: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseRecipeDTO(tt.recipeData)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseRecipeDTO() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("parseRecipeDTO() succeeded unexpectedly")
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseTagDTO() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_parseTagDTO(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tagData []byte
		want    *TagDTO
		wantErr bool
	}{
		{
			name: "planks tag",
			tagData: []byte(`
				{
				  "values": [
				    "minecraft:oak_planks",
				    "minecraft:spruce_planks",
				    "minecraft:birch_planks",
				    "minecraft:jungle_planks",
				    "minecraft:acacia_planks",
				    "minecraft:dark_oak_planks",
				    "minecraft:pale_oak_planks",
				    "minecraft:crimson_planks",
				    "minecraft:warped_planks",
				    "minecraft:mangrove_planks",
				    "minecraft:bamboo_planks",
				    "minecraft:cherry_planks"
				  ]
				}
			`),
			want: &TagDTO{
				Values: []string{
					"minecraft:oak_planks",
					"minecraft:spruce_planks",
					"minecraft:birch_planks",
					"minecraft:jungle_planks",
					"minecraft:acacia_planks",
					"minecraft:dark_oak_planks",
					"minecraft:pale_oak_planks",
					"minecraft:crimson_planks",
					"minecraft:warped_planks",
					"minecraft:mangrove_planks",
					"minecraft:bamboo_planks",
					"minecraft:cherry_planks",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseTagDTO(tt.tagData)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseTagDTO() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseTagDTO() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseTagDTO() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
