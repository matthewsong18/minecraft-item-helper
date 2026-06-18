package main

import (
	"encoding/json"
	"fmt"
)

type RecipeDTO interface {
	RecipeType() string
}

type CraftingShapedRecipeDTO struct {
	Type    string            `json:"type"`
	Pattern []string          `json:"pattern"`
	Key     map[string]string `json:"key"`
	Result  RecipeResultDTO   `json:"result"`
}

type CraftingShapelessRecipeDTO struct {
	Type        string          `json:"type"`
	Ingredients []string        `json:"ingredients"`
	Result      RecipeResultDTO `json:"result"`
}

type RecipeResultDTO struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

func (c CraftingShapedRecipeDTO) RecipeType() string    { return c.Type }
func (c CraftingShapelessRecipeDTO) RecipeType() string { return c.Type }

type TagDTO struct {
	Values []string `json:"values"`
}

func parseRecipeDTO(recipeData []byte) (RecipeDTO, error) {
	var jsonResult struct {
		RecipeType string `json:"type"`
	}

	if err := json.Unmarshal(recipeData, &jsonResult); err != nil {
		return nil, err
	}

	switch jsonResult.RecipeType {
	case "minecraft:crafting_shaped":
		var craftingShapedRecipeDTO CraftingShapedRecipeDTO
		if err := json.Unmarshal(recipeData, &craftingShapedRecipeDTO); err != nil {
			return nil, err
		}
		return craftingShapedRecipeDTO, nil
	case "minecraft:crafting_shapeless":
		var craftingShapelessRecipeDTO CraftingShapelessRecipeDTO
		if err := json.Unmarshal(recipeData, &craftingShapelessRecipeDTO); err != nil {
			return nil, err
		}
		return craftingShapelessRecipeDTO, nil
	default:
		return nil, fmt.Errorf("unsupported recipe type: %s", jsonResult.RecipeType)
	}
}

func parseTagDTO(tagData []byte) (*TagDTO, error) {
	var tagDTO TagDTO
	if err := json.Unmarshal(tagData, &tagDTO); err != nil {
		return nil, err
	}
	return &tagDTO, nil
}
