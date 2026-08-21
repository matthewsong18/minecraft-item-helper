package jsonloader_test

import (
	"minecraft-item-helper/internal/adapters/secondary/jsonloader"
	"minecraft-item-helper/internal/core/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveTags(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		rawTags map[string]jsonloader.TagDTO
		want    jsonloader.TagRegistry
		wantErr bool
	}{
		{
			name: "valid single tag",
			rawTags: map[string]jsonloader.TagDTO{
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
			want: jsonloader.TagRegistry{
				"minecraft:planks": &domain.Tag{
					ID: "minecraft:planks",
					Items: []domain.Ingredient{
						domain.Item{ID: "minecraft:oak_planks"},
						domain.Item{ID: "minecraft:spruce_planks"},
						domain.Item{ID: "minecraft:birch_planks"},
						domain.Item{ID: "minecraft:jungle_planks"},
						domain.Item{ID: "minecraft:acacia_planks"},
						domain.Item{ID: "minecraft:dark_oak_planks"},
						domain.Item{ID: "minecraft:pale_oak_planks"},
						domain.Item{ID: "minecraft:crimson_planks"},
						domain.Item{ID: "minecraft:warped_planks"},
						domain.Item{ID: "minecraft:mangrove_planks"},
						domain.Item{ID: "minecraft:bamboo_planks"},
						domain.Item{ID: "minecraft:cherry_planks"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nested tag recipe",
			rawTags: map[string]jsonloader.TagDTO{
				"minecraft:wooden_tool_materials": jsonloader.TagDTO{
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
			want: jsonloader.TagRegistry{
				"minecraft:wooden_tool_materials": &domain.Tag{
					ID: "minecraft:wooden_tool_materials",
					Items: []domain.Ingredient{
						&domain.Tag{
							ID: "minecraft:planks",
							Items: []domain.Ingredient{
								domain.Item{ID: "minecraft:oak_planks"},
								domain.Item{ID: "minecraft:spruce_planks"},
								domain.Item{ID: "minecraft:birch_planks"},
								domain.Item{ID: "minecraft:jungle_planks"},
								domain.Item{ID: "minecraft:acacia_planks"},
								domain.Item{ID: "minecraft:dark_oak_planks"},
								domain.Item{ID: "minecraft:pale_oak_planks"},
								domain.Item{ID: "minecraft:crimson_planks"},
								domain.Item{ID: "minecraft:warped_planks"},
								domain.Item{ID: "minecraft:mangrove_planks"},
								domain.Item{ID: "minecraft:bamboo_planks"},
								domain.Item{ID: "minecraft:cherry_planks"},
							},
						},
					},
				},
				"minecraft:planks": &domain.Tag{
					ID: "minecraft:planks",
					Items: []domain.Ingredient{
						domain.Item{ID: "minecraft:oak_planks"},
						domain.Item{ID: "minecraft:spruce_planks"},
						domain.Item{ID: "minecraft:birch_planks"},
						domain.Item{ID: "minecraft:jungle_planks"},
						domain.Item{ID: "minecraft:acacia_planks"},
						domain.Item{ID: "minecraft:dark_oak_planks"},
						domain.Item{ID: "minecraft:pale_oak_planks"},
						domain.Item{ID: "minecraft:crimson_planks"},
						domain.Item{ID: "minecraft:warped_planks"},
						domain.Item{ID: "minecraft:mangrove_planks"},
						domain.Item{ID: "minecraft:bamboo_planks"},
						domain.Item{ID: "minecraft:cherry_planks"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := jsonloader.ResolveTags(tt.rawTags)
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
		dto     jsonloader.CraftingShapedRecipeDTO
		tags    jsonloader.TagRegistry
		want    *domain.Recipe
		wantErr bool
	}{
		{
			name: "Valid simple recipe",
			dto: jsonloader.CraftingShapedRecipeDTO{
				Type: "minecraft:crafting_shaped",
				Pattern: []string{
					"# #",
					"###",
				},
				Key: map[string]string{
					"#": "minecraft:acacia_planks",
				},
				Result: jsonloader.RecipeResultDTO{
					ID:    "minecraft:acacia_boat",
					Count: 1,
				},
			},
			tags: jsonloader.TagRegistry{},
			want: &domain.Recipe{
				ResultID: "minecraft:acacia_boat",
				Yield:    1,
				Ingredients: map[string]domain.IngredientAmount{
					"minecraft:acacia_planks": {
						Ingredient: domain.Item{ID: "minecraft:acacia_planks"},
						Amount:     5,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid tag recipe",
			dto: jsonloader.CraftingShapedRecipeDTO{
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
				Result: jsonloader.RecipeResultDTO{
					ID:    "minecraft:barrel",
					Count: 1,
				},
			},
			tags: jsonloader.TagRegistry{
				"minecraft:planks": &domain.Tag{
					ID: "minecraft:planks",
					Items: []domain.Ingredient{
						domain.Item{ID: "minecraft:oak_planks"},
						domain.Item{ID: "minecraft:spruce_planks"},
						domain.Item{ID: "minecraft:birch_planks"},
						domain.Item{ID: "minecraft:jungle_planks"},
						domain.Item{ID: "minecraft:acacia_planks"},
						domain.Item{ID: "minecraft:dark_oak_planks"},
						domain.Item{ID: "minecraft:pale_oak_planks"},
						domain.Item{ID: "minecraft:crimson_planks"},
						domain.Item{ID: "minecraft:warped_planks"},
						domain.Item{ID: "minecraft:mangrove_planks"},
						domain.Item{ID: "minecraft:bamboo_planks"},
						domain.Item{ID: "minecraft:cherry_planks"},
					},
				},
				"minecraft:wooden_slabs": &domain.Tag{
					ID: "minecraft:wooden_slabs",
					Items: []domain.Ingredient{
						domain.Item{ID: "minecraft:oak_slab"},
						domain.Item{ID: "minecraft:spruce_slab"},
						domain.Item{ID: "minecraft:birch_slab"},
						domain.Item{ID: "minecraft:jungle_slab"},
						domain.Item{ID: "minecraft:acacia_slab"},
						domain.Item{ID: "minecraft:dark_oak_slab"},
						domain.Item{ID: "minecraft:pale_oak_slab"},
						domain.Item{ID: "minecraft:crimson_slab"},
						domain.Item{ID: "minecraft:warped_slab"},
						domain.Item{ID: "minecraft:mangrove_slab"},
						domain.Item{ID: "minecraft:bamboo_slab"},
						domain.Item{ID: "minecraft:cherry_slab"},
					},
				},
			},
			want: &domain.Recipe{
				ResultID: "minecraft:barrel",
				Yield:    1,
				Ingredients: map[string]domain.IngredientAmount{
					"minecraft:planks": domain.IngredientAmount{
						Ingredient: &domain.Tag{
							ID: "minecraft:planks",
							Items: []domain.Ingredient{
								domain.Item{ID: "minecraft:oak_planks"},
								domain.Item{ID: "minecraft:spruce_planks"},
								domain.Item{ID: "minecraft:birch_planks"},
								domain.Item{ID: "minecraft:jungle_planks"},
								domain.Item{ID: "minecraft:acacia_planks"},
								domain.Item{ID: "minecraft:dark_oak_planks"},
								domain.Item{ID: "minecraft:pale_oak_planks"},
								domain.Item{ID: "minecraft:crimson_planks"},
								domain.Item{ID: "minecraft:warped_planks"},
								domain.Item{ID: "minecraft:mangrove_planks"},
								domain.Item{ID: "minecraft:bamboo_planks"},
								domain.Item{ID: "minecraft:cherry_planks"},
							},
						},
						Amount: 6,
					},
					"minecraft:wooden_slabs": domain.IngredientAmount{
						Ingredient: &domain.Tag{
							ID: "minecraft:wooden_slabs",
							Items: []domain.Ingredient{
								domain.Item{ID: "minecraft:oak_slab"},
								domain.Item{ID: "minecraft:spruce_slab"},
								domain.Item{ID: "minecraft:birch_slab"},
								domain.Item{ID: "minecraft:jungle_slab"},
								domain.Item{ID: "minecraft:acacia_slab"},
								domain.Item{ID: "minecraft:dark_oak_slab"},
								domain.Item{ID: "minecraft:pale_oak_slab"},
								domain.Item{ID: "minecraft:crimson_slab"},
								domain.Item{ID: "minecraft:warped_slab"},
								domain.Item{ID: "minecraft:mangrove_slab"},
								domain.Item{ID: "minecraft:bamboo_slab"},
								domain.Item{ID: "minecraft:cherry_slab"},
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
			dto: jsonloader.CraftingShapedRecipeDTO{
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
				Result: jsonloader.RecipeResultDTO{
					ID:    "minecraft:wooden_sword",
					Count: 1,
				},
			},
			tags: jsonloader.TagRegistry{
				"minecraft:wooden_tool_materials": &domain.Tag{
					ID: "minecraft:wooden_tool_materials",
					Items: []domain.Ingredient{
						&domain.Tag{
							ID: "minecraft:planks",
							Items: []domain.Ingredient{
								domain.Item{ID: "minecraft:oak_planks"},
								domain.Item{ID: "minecraft:spruce_planks"},
								domain.Item{ID: "minecraft:birch_planks"},
								domain.Item{ID: "minecraft:jungle_planks"},
								domain.Item{ID: "minecraft:acacia_planks"},
								domain.Item{ID: "minecraft:dark_oak_planks"},
								domain.Item{ID: "minecraft:pale_oak_planks"},
								domain.Item{ID: "minecraft:crimson_planks"},
								domain.Item{ID: "minecraft:warped_planks"},
								domain.Item{ID: "minecraft:mangrove_planks"},
								domain.Item{ID: "minecraft:bamboo_planks"},
								domain.Item{ID: "minecraft:cherry_planks"},
							},
						},
					},
				},
				"minecraft:planks": &domain.Tag{
					ID: "minecraft:planks",
					Items: []domain.Ingredient{
						domain.Item{ID: "minecraft:oak_planks"},
						domain.Item{ID: "minecraft:spruce_planks"},
						domain.Item{ID: "minecraft:birch_planks"},
						domain.Item{ID: "minecraft:jungle_planks"},
						domain.Item{ID: "minecraft:acacia_planks"},
						domain.Item{ID: "minecraft:dark_oak_planks"},
						domain.Item{ID: "minecraft:pale_oak_planks"},
						domain.Item{ID: "minecraft:crimson_planks"},
						domain.Item{ID: "minecraft:warped_planks"},
						domain.Item{ID: "minecraft:mangrove_planks"},
						domain.Item{ID: "minecraft:bamboo_planks"},
						domain.Item{ID: "minecraft:cherry_planks"},
					},
				},
			},
			want: &domain.Recipe{
				ResultID: "minecraft:wooden_sword",
				Yield:    1,
				Ingredients: map[string]domain.IngredientAmount{
					"minecraft:stick": domain.IngredientAmount{
						Ingredient: domain.Item{ID: "minecraft:stick"},
						Amount:     1,
					},
					"minecraft:wooden_tool_materials": domain.IngredientAmount{
						Ingredient: &domain.Tag{
							ID: "minecraft:wooden_tool_materials",
							Items: []domain.Ingredient{
								&domain.Tag{
									ID: "minecraft:planks",
									Items: []domain.Ingredient{
										domain.Item{ID: "minecraft:oak_planks"},
										domain.Item{ID: "minecraft:spruce_planks"},
										domain.Item{ID: "minecraft:birch_planks"},
										domain.Item{ID: "minecraft:jungle_planks"},
										domain.Item{ID: "minecraft:acacia_planks"},
										domain.Item{ID: "minecraft:dark_oak_planks"},
										domain.Item{ID: "minecraft:pale_oak_planks"},
										domain.Item{ID: "minecraft:crimson_planks"},
										domain.Item{ID: "minecraft:warped_planks"},
										domain.Item{ID: "minecraft:mangrove_planks"},
										domain.Item{ID: "minecraft:bamboo_planks"},
										domain.Item{ID: "minecraft:cherry_planks"},
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
			got, gotErr := jsonloader.ResolveShapedRecipe(tt.dto, tt.tags)
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
