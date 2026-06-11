package main_test

import (
	main "minecraft-item-helper"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveTags(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		rawTags map[string]main.TagDTO
		want    main.TagRegistry
		wantErr bool
	}{
		{
			name: "valid single tag",
			rawTags: map[string]main.TagDTO{
				"minecraft:planks": {
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
			},
			want: main.TagRegistry{
				"minecraft:planks": &main.Tag{
					ID: "minecraft:planks",
					Items: []main.Ingredient{
						main.Item{ID: "minecraft:oak_planks"},
						main.Item{ID: "minecraft:spruce_planks"},
						main.Item{ID: "minecraft:birch_planks"},
						main.Item{ID: "minecraft:jungle_planks"},
						main.Item{ID: "minecraft:acacia_planks"},
						main.Item{ID: "minecraft:dark_oak_planks"},
						main.Item{ID: "minecraft:pale_oak_planks"},
						main.Item{ID: "minecraft:crimson_planks"},
						main.Item{ID: "minecraft:warped_planks"},
						main.Item{ID: "minecraft:mangrove_planks"},
						main.Item{ID: "minecraft:bamboo_planks"},
						main.Item{ID: "minecraft:cherry_planks"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nested tag recipe",
			rawTags: map[string]main.TagDTO{
				"minecraft:wooden_tool_materials": main.TagDTO{
					Values: []string{
						"#minecraft:planks",
					},
				},
				"minecraft:planks": {
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
			},
			want: main.TagRegistry{
				"minecraft:wooden_tool_materials": &main.Tag{
					ID: "minecraft:wooden_tool_materials",
					Items: []main.Ingredient{
						&main.Tag{
							ID: "minecraft:planks",
							Items: []main.Ingredient{
								main.Item{ID: "minecraft:oak_planks"},
								main.Item{ID: "minecraft:spruce_planks"},
								main.Item{ID: "minecraft:birch_planks"},
								main.Item{ID: "minecraft:jungle_planks"},
								main.Item{ID: "minecraft:acacia_planks"},
								main.Item{ID: "minecraft:dark_oak_planks"},
								main.Item{ID: "minecraft:pale_oak_planks"},
								main.Item{ID: "minecraft:crimson_planks"},
								main.Item{ID: "minecraft:warped_planks"},
								main.Item{ID: "minecraft:mangrove_planks"},
								main.Item{ID: "minecraft:bamboo_planks"},
								main.Item{ID: "minecraft:cherry_planks"},
							},
						},
					},
				},
				"minecraft:planks": &main.Tag{
					ID: "minecraft:planks",
					Items: []main.Ingredient{
						main.Item{ID: "minecraft:oak_planks"},
						main.Item{ID: "minecraft:spruce_planks"},
						main.Item{ID: "minecraft:birch_planks"},
						main.Item{ID: "minecraft:jungle_planks"},
						main.Item{ID: "minecraft:acacia_planks"},
						main.Item{ID: "minecraft:dark_oak_planks"},
						main.Item{ID: "minecraft:pale_oak_planks"},
						main.Item{ID: "minecraft:crimson_planks"},
						main.Item{ID: "minecraft:warped_planks"},
						main.Item{ID: "minecraft:mangrove_planks"},
						main.Item{ID: "minecraft:bamboo_planks"},
						main.Item{ID: "minecraft:cherry_planks"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := main.ResolveTags(tt.rawTags)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResolveTags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResolveTags() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ResolveTags() mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestResolveShapedRecipe(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		dto     main.CraftingShapedRecipeDTO
		tags    main.TagRegistry
		want    *main.Recipe
		wantErr bool
	}{
		{
			name: "Valid simple recipe",
			dto: main.CraftingShapedRecipeDTO{
				Type: "minecraft:crafting_shaped",
				Pattern: []string{
					"# #",
					"###",
				},
				Key: map[string]string{
					"#": "minecraft:acacia_planks",
				},
				Result: main.RecipeResultDTO{
					ID:    "minecraft:acacia_boat",
					Count: 1,
				},
			},
			tags: main.TagRegistry{},
			want: &main.Recipe{
				ResultID: "minecraft:acacia_boat",
				Yield:    1,
				Ingredients: map[string]main.IngredientAmount{
					"minecraft:acacia_planks": {
						Ingredient: main.Item{ID: "minecraft:acacia_planks"},
						Amount:     5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid tag recipe",
			dto: main.CraftingShapedRecipeDTO{
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
				Result: main.RecipeResultDTO{
					ID:    "minecraft:barrel",
					Count: 1,
				},
			},
			tags: main.TagRegistry{
				"minecraft:planks": &main.Tag{
					ID: "minecraft:planks",
					Items: []main.Ingredient{
						main.Item{ID: "minecraft:oak_planks"},
						main.Item{ID: "minecraft:spruce_planks"},
						main.Item{ID: "minecraft:birch_planks"},
						main.Item{ID: "minecraft:jungle_planks"},
						main.Item{ID: "minecraft:acacia_planks"},
						main.Item{ID: "minecraft:dark_oak_planks"},
						main.Item{ID: "minecraft:pale_oak_planks"},
						main.Item{ID: "minecraft:crimson_planks"},
						main.Item{ID: "minecraft:warped_planks"},
						main.Item{ID: "minecraft:mangrove_planks"},
						main.Item{ID: "minecraft:bamboo_planks"},
						main.Item{ID: "minecraft:cherry_planks"},
					},
				},
				"minecraft:wooden_slabs": &main.Tag{
					ID: "minecraft:wooden_slabs",
					Items: []main.Ingredient{
						main.Item{ID: "minecraft:oak_slab"},
						main.Item{ID: "minecraft:spruce_slab"},
						main.Item{ID: "minecraft:birch_slab"},
						main.Item{ID: "minecraft:jungle_slab"},
						main.Item{ID: "minecraft:acacia_slab"},
						main.Item{ID: "minecraft:dark_oak_slab"},
						main.Item{ID: "minecraft:pale_oak_slab"},
						main.Item{ID: "minecraft:crimson_slab"},
						main.Item{ID: "minecraft:warped_slab"},
						main.Item{ID: "minecraft:mangrove_slab"},
						main.Item{ID: "minecraft:bamboo_slab"},
						main.Item{ID: "minecraft:cherry_slab"},
					},
				},
			},
			want: &main.Recipe{
				ResultID: "minecraft:barrel",
				Yield:    1,
				Ingredients: map[string]main.IngredientAmount{
					"minecraft:planks": main.IngredientAmount{
						Ingredient: &main.Tag{
							ID: "minecraft:planks",
							Items: []main.Ingredient{
								main.Item{ID: "minecraft:oak_planks"},
								main.Item{ID: "minecraft:spruce_planks"},
								main.Item{ID: "minecraft:birch_planks"},
								main.Item{ID: "minecraft:jungle_planks"},
								main.Item{ID: "minecraft:acacia_planks"},
								main.Item{ID: "minecraft:dark_oak_planks"},
								main.Item{ID: "minecraft:pale_oak_planks"},
								main.Item{ID: "minecraft:crimson_planks"},
								main.Item{ID: "minecraft:warped_planks"},
								main.Item{ID: "minecraft:mangrove_planks"},
								main.Item{ID: "minecraft:bamboo_planks"},
								main.Item{ID: "minecraft:cherry_planks"},
							},
						},
						Amount: 6,
					},
					"minecraft:wooden_slabs": main.IngredientAmount{
						Ingredient: &main.Tag{
							ID: "minecraft:wooden_slabs",
							Items: []main.Ingredient{
								main.Item{ID: "minecraft:oak_slab"},
								main.Item{ID: "minecraft:spruce_slab"},
								main.Item{ID: "minecraft:birch_slab"},
								main.Item{ID: "minecraft:jungle_slab"},
								main.Item{ID: "minecraft:acacia_slab"},
								main.Item{ID: "minecraft:dark_oak_slab"},
								main.Item{ID: "minecraft:pale_oak_slab"},
								main.Item{ID: "minecraft:crimson_slab"},
								main.Item{ID: "minecraft:warped_slab"},
								main.Item{ID: "minecraft:mangrove_slab"},
								main.Item{ID: "minecraft:bamboo_slab"},
								main.Item{ID: "minecraft:cherry_slab"},
							},
						},
						Amount: 2,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "double nested recipe",
			dto: main.CraftingShapedRecipeDTO{
				Type: "minecraft:crafting_shaped",
				Pattern: []string{
					"X",
					"X",
					"#",
				},
				Key: map[string]string{
					"#": "minecraft:stick",
					"X": "#minecraft:wooden_tool_materials",
				},
				Result: main.RecipeResultDTO{
					ID:    "minecraft:wooden_sword",
					Count: 1,
				},
			},
			tags: main.TagRegistry{
				"minecraft:wooden_tool_materials": &main.Tag{
					ID: "minecraft:wooden_tool_materials",
					Items: []main.Ingredient{
						&main.Tag{
							ID: "minecraft:planks",
							Items: []main.Ingredient{
								main.Item{ID: "minecraft:oak_planks"},
								main.Item{ID: "minecraft:spruce_planks"},
								main.Item{ID: "minecraft:birch_planks"},
								main.Item{ID: "minecraft:jungle_planks"},
								main.Item{ID: "minecraft:acacia_planks"},
								main.Item{ID: "minecraft:dark_oak_planks"},
								main.Item{ID: "minecraft:pale_oak_planks"},
								main.Item{ID: "minecraft:crimson_planks"},
								main.Item{ID: "minecraft:warped_planks"},
								main.Item{ID: "minecraft:mangrove_planks"},
								main.Item{ID: "minecraft:bamboo_planks"},
								main.Item{ID: "minecraft:cherry_planks"},
							},
						},
					},
				},
				"minecraft:planks": &main.Tag{
					ID: "minecraft:planks",
					Items: []main.Ingredient{
						main.Item{ID: "minecraft:oak_planks"},
						main.Item{ID: "minecraft:spruce_planks"},
						main.Item{ID: "minecraft:birch_planks"},
						main.Item{ID: "minecraft:jungle_planks"},
						main.Item{ID: "minecraft:acacia_planks"},
						main.Item{ID: "minecraft:dark_oak_planks"},
						main.Item{ID: "minecraft:pale_oak_planks"},
						main.Item{ID: "minecraft:crimson_planks"},
						main.Item{ID: "minecraft:warped_planks"},
						main.Item{ID: "minecraft:mangrove_planks"},
						main.Item{ID: "minecraft:bamboo_planks"},
						main.Item{ID: "minecraft:cherry_planks"},
					},
				},
			},
			want: &main.Recipe{
				ResultID: "minecraft:wooden_sword",
				Yield:    1,
				Ingredients: map[string]main.IngredientAmount{
					"minecraft:stick": main.IngredientAmount{
						Ingredient: main.Item{ID: "minecraft:stick"},
						Amount:     1,
					},
					"minecraft:wooden_tool_materials": main.IngredientAmount{
						Ingredient: &main.Tag{
							ID: "minecraft:wooden_tool_materials",
							Items: []main.Ingredient{
								&main.Tag{
									ID: "minecraft:planks",
									Items: []main.Ingredient{
										main.Item{ID: "minecraft:oak_planks"},
										main.Item{ID: "minecraft:spruce_planks"},
										main.Item{ID: "minecraft:birch_planks"},
										main.Item{ID: "minecraft:jungle_planks"},
										main.Item{ID: "minecraft:acacia_planks"},
										main.Item{ID: "minecraft:dark_oak_planks"},
										main.Item{ID: "minecraft:pale_oak_planks"},
										main.Item{ID: "minecraft:crimson_planks"},
										main.Item{ID: "minecraft:warped_planks"},
										main.Item{ID: "minecraft:mangrove_planks"},
										main.Item{ID: "minecraft:bamboo_planks"},
										main.Item{ID: "minecraft:cherry_planks"},
									},
								},
							},
						},
						Amount: 2,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := main.ResolveShapedRecipe(tt.dto, tt.tags)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResolveShapedRecipe() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResolveShapedRecipe() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ResolveShapedRecipe() mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
