package jsonloader

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

type SmeltingRecipeDTO struct {
	Type       string          `json:"type"`
	Ingredient json.RawMessage `json:"ingredient"`
	Result     RecipeResultDTO `json:"result"`
}

type RecipeResultDTO struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

func (r *RecipeResultDTO) UnmarshalJSON(data []byte) error {
	type Alias RecipeResultDTO
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.Count == 0 {
		r.Count = 1
	}
	return nil
}

func (c CraftingShapedRecipeDTO) RecipeType() string    { return c.Type }
func (c CraftingShapelessRecipeDTO) RecipeType() string { return c.Type }
func (c SmeltingRecipeDTO) RecipeType() string          { return c.Type }

type TagDTO struct {
	Values []string `json:"values"`
}

func ParseRecipeDTO(recipeData []byte) (RecipeDTO, error) {
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
	case "minecraft:smelting", "minecraft:blasting", "minecraft:campfire_cooking", "minecraft:smoking":
		var smeltingRecipeDTO SmeltingRecipeDTO
		if err := json.Unmarshal(recipeData, &smeltingRecipeDTO); err != nil {
			return nil, err
		}
		return smeltingRecipeDTO, nil
	default:
		return nil, fmt.Errorf("unsupported recipe type: %s", jsonResult.RecipeType)
	}
}

func ParseTagDTO(tagData []byte) (*TagDTO, error) {
	var tagDTO TagDTO
	if err := json.Unmarshal(tagData, &tagDTO); err != nil {
		return nil, err
	}
	return &tagDTO, nil
}
