package main

import (
	"encoding/json"
	"fmt"
	"minecraft-item-helper/internal/adapters/secondary/jsonloader"
	"minecraft-item-helper/internal/core/domain"
	"minecraft-item-helper/internal/service"
	"os"
	"path/filepath"
	"strings"
)

func main() {
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
		return
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

	// 5. Calculate Tree
	for _, target := range []string{"minecraft:lectern", "minecraft:book", "minecraft:glass", "minecraft:hopper_minecart"} {
		fmt.Printf("\nCalculating recipe for %s...\n", target)
		root, err := service.CalculateRecipeTree(target, 80, recipes, tags, nil)
		if err != nil {
			fmt.Printf("Calculation error for %s: %v\n", target, err)
			continue
		}
		printNode(root, 0)
	}
}

func printNode(n *domain.Node, indent int) {
	fmt.Printf("%s%d x %s (Multiplier: %d)\n",
		strings.Repeat("  ", indent),
		n.TargetAmount,
		n.Item,
		n.RecipeMultiplier)

	for _, child := range n.Children {
		printNode(&child, indent+1)
	}
}
