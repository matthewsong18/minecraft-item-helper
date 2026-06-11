package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 1. Parse Tags
	rawTags := make(map[string]TagDTO)
	tagFiles, _ := filepath.Glob("assets/item-tag-groups/*.json")
	for _, path := range tagFiles {
		data, _ := os.ReadFile(path)
		var dto TagDTO
		json.Unmarshal(data, &dto)

		// Use the filename (minus .json) as the tag ID, prefixed with minecraft:
		id := "minecraft:" + strings.TrimSuffix(filepath.Base(path), ".json")
		rawTags[id] = dto
	}

	// 2. Parse Recipes
	rawRecipes := make(map[string]CraftingShapedRecipeDTO)
	recipeFiles, _ := filepath.Glob("assets/recipes/*.json")
	for _, path := range recipeFiles {
		data, _ := os.ReadFile(path)

		// For simplicity, we only handle shaped recipes for now
		if recipeDTO, err := parseRecipeDTO(data); err == nil && recipeDTO.RecipeType() == "minecraft:crafting_shaped" {
			shapedRecipeDTO := recipeDTO.(CraftingShapedRecipeDTO)
			rawRecipes[shapedRecipeDTO.Result.ID] = shapedRecipeDTO
		}
	}

	// 3. Compile Tags
	tags, err := ResolveTags(rawTags)
	if err != nil {
		fmt.Printf("Error compiling tags: %v\n", err)
		return
	}

	// 4. Compile Recipes
	recipes := make(map[string]*Recipe)
	for id, dto := range rawRecipes {
		recipe, err := ResolveShapedRecipe(dto, tags)
		if err != nil {
			fmt.Printf("Error compiling recipe for %s: %v\n", id, err)
			continue
		}
		recipes[id] = recipe
	}

	// 5. Calculate Tree for a Barrel
	fmt.Println("Calculating recipe for minecraft:lectern...")
	root, err := CalculateRecipe("minecraft:lectern", 9, recipes, tags)
	if err != nil {
		fmt.Printf("Calculation error: %v\n", err)
		return
	}

	printNode(root, 0)
}

func printNode(n *Node, indent int) {
	fmt.Printf("%s%d x %s (Multiplier: %d)\n",
		strings.Repeat("  ", indent),
		n.TargetAmount,
		n.Item,
		n.RecipeMultiplier)

	for _, child := range n.Children {
		printNode(&child, indent+1)
	}
}
