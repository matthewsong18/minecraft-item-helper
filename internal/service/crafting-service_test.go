package service_test

import (
	"encoding/json"
	"fmt"
	"minecraft-item-helper/internal/adapters/secondary/jsonloader"
	"minecraft-item-helper/internal/core/domain"
	"minecraft-item-helper/internal/service"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func setup() (map[string]*domain.Recipe, jsonloader.TagRegistry) {
	// 1. Parse Tags
	rawTags := make(map[string]jsonloader.TagDTO)
	tagFiles, _ := filepath.Glob("assets/item-tag-groups/*.json")
	for _, path := range tagFiles {
		data, _ := os.ReadFile(path)
		var dto jsonloader.TagDTO
		json.Unmarshal(data, &dto)

		// Use the filename (minus .json) as the tag ID, prefixed with minecraft:
		id := "minecraft:" + strings.TrimSuffix(filepath.Base(path), ".json")
		rawTags[id] = dto
	}

	// 2. Parse Recipes
	recipeFiles, _ := filepath.Glob("assets/recipes/*.json")
	recipeDTOs := make([]jsonloader.RecipeDTO, 0)
	for _, path := range recipeFiles {
		data, _ := os.ReadFile(path)

		if dto, err := jsonloader.ParseRecipeDTO(data); err == nil {
			recipeDTOs = append(recipeDTOs, dto)
		}
	}

	// 3. Compile Tags
	tags, err := jsonloader.ResolveTags(rawTags)
	if err != nil {
		fmt.Printf("Error compiling tags: %v\n", err)
		return nil, nil
	}

	// 4. Compile Recipes
	recipes := make(map[string]*domain.Recipe)
	for _, dto := range recipeDTOs {
		var recipe *domain.Recipe
		var err error

		switch d := dto.(type) {
		case jsonloader.CraftingShapedRecipeDTO:
			recipe, err = jsonloader.ResolveShapedRecipe(d, tags)
		case jsonloader.CraftingShapelessRecipeDTO:
			recipe, err = jsonloader.ResolveShapelessRecipe(d, tags)
		case jsonloader.SmeltingRecipeDTO:
			recipe, err = jsonloader.ResolveSmeltingRecipe(d, tags)
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
		recipes      map[string]*domain.Recipe
		tags         jsonloader.TagRegistry
		want         *domain.Node
		wantErr      bool
	}{
		{
			name:         "simple tree",
			targetID:     "minecraft:oak_boat",
			targetAmount: 1,
			recipes: map[string]*domain.Recipe{
				"minecraft:oak_boat": &domain.Recipe{
					ResultID: "minecraft:oak_boat",
					Yield:    1,
					Ingredients: map[string]domain.IngredientAmount{
						"minecraft:oak_planks": domain.IngredientAmount{
							Ingredient: domain.Item{ID: "minecraft:oak_planks"},
							Amount:     5,
						},
					},
				},
				"minecraft:oak_planks": &domain.Recipe{
					ResultID: "minecraft:oak_planks",
					Yield:    4,
					Ingredients: map[string]domain.IngredientAmount{
						"minecraft:oak_logs": domain.IngredientAmount{
							Ingredient: domain.Item{ID: "minecraft:oak_logs"},
							Amount:     1,
						},
					},
				},
			},
			tags: jsonloader.TagRegistry{},
			want: &domain.Node{
				Item: "minecraft:oak_boat",
				Recipe: &domain.Recipe{
					ResultID: "minecraft:oak_boat",
					Yield:    1,
					Ingredients: map[string]domain.IngredientAmount{
						"minecraft:oak_planks": domain.IngredientAmount{
							Ingredient: domain.Item{ID: "minecraft:oak_planks"},
							Amount:     5,
						},
					},
				},
				TargetAmount:     1,
				RecipeMultiplier: 1,
				Children: []domain.Node{
					{
						Item: "minecraft:oak_planks",
						Recipe: &domain.Recipe{
							ResultID: "minecraft:oak_planks",
							Yield:    4,
							Ingredients: map[string]domain.IngredientAmount{
								"minecraft:oak_logs": domain.IngredientAmount{
									Ingredient: domain.Item{ID: "minecraft:oak_logs"},
									Amount:     1,
								},
							},
						},
						TargetAmount:     5,
						RecipeMultiplier: 2,
						Children: []domain.Node{
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
			got, gotErr := service.CalculateRecipeTree(tt.targetID, tt.targetAmount, tt.recipes, tt.tags, nil)
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
		recipes map[string]*domain.Recipe
		tags    jsonloader.TagRegistry
		// Named input parameters for target function.
		itemRequests []*service.ItemRequest
		want         *service.CraftingResult
		wantErr      bool
	}{
		{
			name:    "simple test",
			recipes: recipes,
			tags:    tags,
			itemRequests: []*service.ItemRequest{
				{
					ItemID: "minecraft:lectern",
					Amount: 1,
				},
				{
					ItemID: "minecraft:book",
					Amount: 1,
				},
			},
			want: &service.CraftingResult{
				Trees: []*domain.Node{
					&domain.Node{
						Item: "minecraft:lectern",
						Recipe: &domain.Recipe{
							ResultID: "minecraft:lectern",
							Yield:    1,
							Ingredients: map[string]domain.IngredientAmount{
								"minecraft:bookshelf": {
									Ingredient: domain.Item{ID: "minecraft:bookshelf"},
									Amount:     1,
								},
								"minecraft:wooden_slabs": {
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
									Amount: 4,
								},
							},
						},
						TargetAmount:     1,
						RecipeMultiplier: 1,
						Children: []domain.Node{
							{
								Item: "minecraft:bookshelf",
								Recipe: &domain.Recipe{
									ResultID: "minecraft:bookshelf",
									Yield:    1,
									Ingredients: map[string]domain.IngredientAmount{
										"minecraft:book": {Ingredient: domain.Item{ID: "minecraft:book"}, Amount: 3},
										"minecraft:planks": {
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
									},
								},
								TargetAmount:     1,
								RecipeMultiplier: 1,
								Children: []domain.Node{
									{
										Item: "minecraft:book",
										Recipe: &domain.Recipe{
											ResultID: "minecraft:book",
											Yield:    1,
											Ingredients: map[string]domain.IngredientAmount{
												"minecraft:leather": {
													Ingredient: domain.Item{ID: "minecraft:leather"},
													Amount:     1,
												},
												"minecraft:paper": {
													Ingredient: domain.Item{ID: "minecraft:paper"},
													Amount:     3,
												},
											},
										},
										TargetAmount:     3,
										RecipeMultiplier: 3,
										Children: []domain.Node{
											{
												Item: "minecraft:leather",
												Recipe: &domain.Recipe{
													ResultID: "minecraft:leather",
													Yield:    1,
													Ingredients: map[string]domain.IngredientAmount{
														"minecraft:rabbit_hide": {
															Ingredient: domain.Item{ID: "minecraft:rabbit_hide"},
															Amount:     4,
														},
													},
												},
												TargetAmount:     3,
												RecipeMultiplier: 3,
												Children: []domain.Node{
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
												Recipe: &domain.Recipe{
													ResultID: "minecraft:paper",
													Yield:    3,
													Ingredients: map[string]domain.IngredientAmount{
														"minecraft:sugar_cane": {
															Ingredient: domain.Item{ID: "minecraft:sugar_cane"},
															Amount:     3,
														},
													},
												},
												TargetAmount:     9,
												RecipeMultiplier: 3,
												Children: []domain.Node{
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
										Recipe: &domain.Recipe{
											ResultID: "minecraft:oak_planks",
											Yield:    4,
											Ingredients: map[string]domain.IngredientAmount{
												"minecraft:oak_logs": {
													Ingredient: &domain.Tag{
														ID: "minecraft:oak_logs",
														Items: []domain.Ingredient{
															domain.Item{ID: "minecraft:oak_log"},
															domain.Item{ID: "minecraft:oak_wood"},
															domain.Item{ID: "minecraft:stripped_oak_log"},
															domain.Item{ID: "minecraft:stripped_oak_wood"},
														},
													},
													Amount: 1,
												},
											},
										},
										TargetAmount:     6,
										RecipeMultiplier: 2,
										Children: []domain.Node{
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
								Recipe: &domain.Recipe{
									ResultID: "minecraft:oak_slab",
									Yield:    6,
									Ingredients: map[string]domain.IngredientAmount{
										"minecraft:oak_planks": {
											Ingredient: domain.Item{ID: "minecraft:oak_planks"},
											Amount:     3,
										},
									},
								},
								TargetAmount:     4,
								RecipeMultiplier: 1,
								Children: []domain.Node{
									{
										Item: "minecraft:oak_planks",
										Recipe: &domain.Recipe{
											ResultID: "minecraft:oak_planks",
											Yield:    4,
											Ingredients: map[string]domain.IngredientAmount{
												"minecraft:oak_logs": {
													Ingredient: &domain.Tag{
														ID: "minecraft:oak_logs",
														Items: []domain.Ingredient{
															domain.Item{ID: "minecraft:oak_log"},
															domain.Item{ID: "minecraft:oak_wood"},
															domain.Item{ID: "minecraft:stripped_oak_log"},
															domain.Item{ID: "minecraft:stripped_oak_wood"},
														},
													},
													Amount: 1,
												},
											},
										},
										TargetAmount:     3,
										RecipeMultiplier: 1,
										Children: []domain.Node{
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
					&domain.Node{
						Item: "minecraft:book",
						Recipe: &domain.Recipe{
							ResultID: "minecraft:book",
							Yield:    1,
							Ingredients: map[string]domain.IngredientAmount{
								"minecraft:leather": {
									Ingredient: domain.Item{ID: "minecraft:leather"},
									Amount:     1,
								},
								"minecraft:paper": {
									Ingredient: domain.Item{ID: "minecraft:paper"},
									Amount:     3,
								},
							},
						},
						TargetAmount:     1,
						RecipeMultiplier: 1,
						Children: []domain.Node{
							{
								Item: "minecraft:leather",
								Recipe: &domain.Recipe{
									ResultID: "minecraft:leather",
									Yield:    1,
									Ingredients: map[string]domain.IngredientAmount{
										"minecraft:rabbit_hide": {
											Ingredient: domain.Item{ID: "minecraft:rabbit_hide"},
											Amount:     4,
										},
									},
								},
								TargetAmount:     1,
								RecipeMultiplier: 1,
								Children: []domain.Node{
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
								Recipe: &domain.Recipe{
									ResultID: "minecraft:paper",
									Yield:    3,
									Ingredients: map[string]domain.IngredientAmount{
										"minecraft:sugar_cane": {
											Ingredient: domain.Item{ID: "minecraft:sugar_cane"},
											Amount:     3,
										},
									},
								},
								TargetAmount:     3,
								RecipeMultiplier: 1,
								Children: []domain.Node{
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
			s := service.NewCraftingService(tt.recipes, tt.tags)
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
