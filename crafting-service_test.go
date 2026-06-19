package main_test

import (
	"encoding/json"
	"fmt"
	main "minecraft-item-helper"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func setup() (map[string]*main.Recipe, main.TagRegistry) {
	// 1. Parse Tags
	rawTags := make(map[string]main.TagDTO)
	tagFiles, _ := filepath.Glob("assets/item-tag-groups/*.json")
	for _, path := range tagFiles {
		data, _ := os.ReadFile(path)
		var dto main.TagDTO
		json.Unmarshal(data, &dto)

		// Use the filename (minus .json) as the tag ID, prefixed with minecraft:
		id := "minecraft:" + strings.TrimSuffix(filepath.Base(path), ".json")
		rawTags[id] = dto
	}

	// 2. Parse Recipes
	recipeFiles, _ := filepath.Glob("assets/recipes/*.json")
	recipeDTOs := make([]main.RecipeDTO, 0)
	for _, path := range recipeFiles {
		data, _ := os.ReadFile(path)

		if dto, err := main.ParseRecipeDTO(data); err == nil {
			recipeDTOs = append(recipeDTOs, dto)
		}
	}

	// 3. Compile Tags
	tags, err := main.ResolveTags(rawTags)
	if err != nil {
		fmt.Printf("Error compiling tags: %v\n", err)
		return nil, nil
	}

	// 4. Compile Recipes
	recipes := make(map[string]*main.Recipe)
	for _, dto := range recipeDTOs {
		var recipe *main.Recipe
		var err error

		switch d := dto.(type) {
		case main.CraftingShapedRecipeDTO:
			recipe, err = main.ResolveShapedRecipe(d, tags)
		case main.CraftingShapelessRecipeDTO:
			recipe, err = main.ResolveShapelessRecipe(d, tags)
		case main.SmeltingRecipeDTO:
			recipe, err = main.ResolveSmeltingRecipe(d, tags)
		}

		if err != nil {
			fmt.Printf("Error compiling recipe: %v\n", err)
			continue
		}
		if recipe != nil {
			// If multiple recipes produce the same result, we just keep the last one for now.
			// Smelting often has blasting/smelting/campfire variants.
			recipes[recipe.ResultID] = recipe
		}
	}

	return recipes, tags
}

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
			got, gotErr := main.CalculateRecipeTree(tt.targetID, tt.targetAmount, tt.recipes, tt.tags, nil)
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

func TestCraftingService_ComputeDependencies(t *testing.T) {
	recipes, tags := setup()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		recipes map[string]*main.Recipe
		tags    main.TagRegistry
		// Named input parameters for target function.
		itemRequests []*main.ItemRequest
		want         *main.CraftingResult
		wantErr      bool
	}{
		{
			name:    "simple test",
			recipes: recipes,
			tags:    tags,
			itemRequests: []*main.ItemRequest{
				{
					ItemID: "minecraft:lectern",
					Amount: 1,
				},
				{
					ItemID: "minecraft:book",
					Amount: 1,
				},
			},
			want: &main.CraftingResult{
				Trees: []*main.Node{
					&main.Node{
						Item: "minecraft:lectern",
						Recipe: &main.Recipe{
							ResultID: "minecraft:lectern",
							Yield:    1,
							Ingredients: map[string]main.IngredientAmount{
								"minecraft:bookshelf": {
									Ingredient: main.Item{ID: "minecraft:bookshelf"},
									Amount:     1,
								},
								"minecraft:wooden_slabs": {
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
									Amount: 4,
								},
							},
						},
						TargetAmount:     1,
						RecipeMultiplier: 1,
						Children: []main.Node{
							{
								Item: "minecraft:bookshelf",
								Recipe: &main.Recipe{
									ResultID: "minecraft:bookshelf",
									Yield:    1,
									Ingredients: map[string]main.IngredientAmount{
										"minecraft:book": {Ingredient: main.Item{ID: "minecraft:book"}, Amount: 3},
										"minecraft:planks": {
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
									},
								},
								TargetAmount:     1,
								RecipeMultiplier: 1,
								Children: []main.Node{
									{
										Item: "minecraft:book",
										Recipe: &main.Recipe{
											ResultID: "minecraft:book",
											Yield:    1,
											Ingredients: map[string]main.IngredientAmount{
												"minecraft:leather": {
													Ingredient: main.Item{ID: "minecraft:leather"},
													Amount:     1,
												},
												"minecraft:paper": {
													Ingredient: main.Item{ID: "minecraft:paper"},
													Amount:     3,
												},
											},
										},
										TargetAmount:     3,
										RecipeMultiplier: 3,
										Children: []main.Node{
											{
												Item: "minecraft:leather",
												Recipe: &main.Recipe{
													ResultID: "minecraft:leather",
													Yield:    1,
													Ingredients: map[string]main.IngredientAmount{
														"minecraft:rabbit_hide": {
															Ingredient: main.Item{ID: "minecraft:rabbit_hide"},
															Amount:     4,
														},
													},
												},
												TargetAmount:     3,
												RecipeMultiplier: 3,
												Children: []main.Node{
													{
														Item:             "minecraft:rabbit_hide",
														Recipe:           nil,
														TargetAmount:     12,
														RecipeMultiplier: 0,
														Children:         nil,
													},
												},
											},
											{
												Item: "minecraft:paper",
												Recipe: &main.Recipe{
													ResultID: "minecraft:paper",
													Yield:    3,
													Ingredients: map[string]main.IngredientAmount{
														"minecraft:sugar_cane": {
															Ingredient: main.Item{ID: "minecraft:sugar_cane"},
															Amount:     3,
														},
													},
												},
												TargetAmount:     9,
												RecipeMultiplier: 3,
												Children: []main.Node{
													{
														Item:             "minecraft:sugar_cane",
														Recipe:           nil,
														TargetAmount:     9,
														RecipeMultiplier: 0,
														Children:         nil,
													},
												},
											},
										},
									},
									{
										Item: "minecraft:oak_planks",
										Recipe: &main.Recipe{
											ResultID: "minecraft:oak_planks",
											Yield:    4,
											Ingredients: map[string]main.IngredientAmount{
												"minecraft:oak_logs": {
													Ingredient: &main.Tag{
														ID: "minecraft:oak_logs",
														Items: []main.Ingredient{
															main.Item{ID: "minecraft:oak_log"},
															main.Item{ID: "minecraft:oak_wood"},
															main.Item{ID: "minecraft:stripped_oak_log"},
															main.Item{ID: "minecraft:stripped_oak_wood"},
														},
													},
													Amount: 1,
												},
											},
										},
										TargetAmount:     6,
										RecipeMultiplier: 2,
										Children: []main.Node{
											{
												Item:             "minecraft:oak_log",
												Recipe:           nil,
												TargetAmount:     2,
												RecipeMultiplier: 0,
												Children:         nil,
											},
										},
									},
								},
							},
							{
								Item: "minecraft:oak_slab",
								Recipe: &main.Recipe{
									ResultID: "minecraft:oak_slab",
									Yield:    6,
									Ingredients: map[string]main.IngredientAmount{
										"minecraft:oak_planks": {
											Ingredient: main.Item{ID: "minecraft:oak_planks"},
											Amount:     3,
										},
									},
								},
								TargetAmount:     4,
								RecipeMultiplier: 1,
								Children: []main.Node{
									{
										Item: "minecraft:oak_planks",
										Recipe: &main.Recipe{
											ResultID: "minecraft:oak_planks",
											Yield:    4,
											Ingredients: map[string]main.IngredientAmount{
												"minecraft:oak_logs": {
													Ingredient: &main.Tag{
														ID: "minecraft:oak_logs",
														Items: []main.Ingredient{
															main.Item{ID: "minecraft:oak_log"},
															main.Item{ID: "minecraft:oak_wood"},
															main.Item{ID: "minecraft:stripped_oak_log"},
															main.Item{ID: "minecraft:stripped_oak_wood"},
														},
													},
													Amount: 1,
												},
											},
										},
										TargetAmount:     3,
										RecipeMultiplier: 1,
										Children: []main.Node{
											{
												Item:             "minecraft:oak_log",
												Recipe:           nil,
												TargetAmount:     1,
												RecipeMultiplier: 0,
												Children:         nil,
											},
										},
									},
								},
							},
						},
					},
					&main.Node{
						Item: "minecraft:book",
						Recipe: &main.Recipe{
							ResultID: "minecraft:book",
							Yield:    1,
							Ingredients: map[string]main.IngredientAmount{
								"minecraft:leather": {
									Ingredient: main.Item{ID: "minecraft:leather"},
									Amount:     1,
								},
								"minecraft:paper": {
									Ingredient: main.Item{ID: "minecraft:paper"},
									Amount:     3,
								},
							},
						},
						TargetAmount:     1,
						RecipeMultiplier: 1,
						Children: []main.Node{
							{
								Item: "minecraft:leather",
								Recipe: &main.Recipe{
									ResultID: "minecraft:leather",
									Yield:    1,
									Ingredients: map[string]main.IngredientAmount{
										"minecraft:rabbit_hide": {
											Ingredient: main.Item{ID: "minecraft:rabbit_hide"},
											Amount:     4,
										},
									},
								},
								TargetAmount:     1,
								RecipeMultiplier: 1,
								Children: []main.Node{
									{
										Item:             "minecraft:rabbit_hide",
										Recipe:           nil,
										TargetAmount:     4,
										RecipeMultiplier: 0,
										Children:         nil,
									},
								},
							},
							{
								Item: "minecraft:paper",
								Recipe: &main.Recipe{
									ResultID: "minecraft:paper",
									Yield:    3,
									Ingredients: map[string]main.IngredientAmount{
										"minecraft:sugar_cane": {
											Ingredient: main.Item{ID: "minecraft:sugar_cane"},
											Amount:     3,
										},
									},
								},
								TargetAmount:     3,
								RecipeMultiplier: 1,
								Children: []main.Node{
									{
										Item:             "minecraft:sugar_cane",
										Recipe:           nil,
										TargetAmount:     3,
										RecipeMultiplier: 0,
										Children:         nil,
									},
								},
							},
						},
					},
				},
				BaseMaterials: map[string]int{
					"minecraft:oak_log":     3,
					"minecraft:rabbit_hide": 16,
					"minecraft:sugar_cane":  12,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := main.NewCraftingService(tt.recipes, tt.tags)
			got, gotErr := s.ComputeDependencies(tt.itemRequests)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ComputeDependencies() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ComputeDependencies() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ComputeDependencies() mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
