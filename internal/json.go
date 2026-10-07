package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ParseJsonResponse struct {
	Id     string
	Recipe Recipe
}

var ErrUnsupportedRecipeType = errors.New("unsupported recipe type")

func ParseJsonRecipe(data []byte) (*ParseJsonResponse, error) {
	var router struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &router); err != nil {
		return nil, fmt.Errorf("failed to parse message type: %w", err)
	}

	switch router.Type {
	case "minecraft:crafting_shapeless":
		return handleShapelessRecipe(data)
	case "minecraft:crafting_shaped":
		return handleShapedRecipe(data)
	case "minecraft:smelting":
		return handleSmeltingRecipe(data)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedRecipeType, router.Type)
	}
}

type RawIngredient []string

func (r *RawIngredient) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*r = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*r = multiple
		return nil
	}

	return fmt.Errorf("invalid ingredient format: %s", string(data))
}

func handleShapelessRecipe(data []byte) (*ParseJsonResponse, error) {
	var dto struct {
		Ingredients []RawIngredient `json:"ingredients"`
		Result      struct {
			Count uint   `json:"count"`
			Id    string `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	groupedItems := make(map[string]Ingredient)
	for _, raw := range dto.Ingredients {
		if len(raw) > 1 {
			// Just using raw[0] to get a key
			groupedItems[raw[0]] = Ingredient{
				Options: raw,
				Count:   1,
			}
			continue
		}

		if existing, found := groupedItems[raw[0]]; found {
			existing.Count++
			groupedItems[raw[0]] = existing
		} else {
			groupedItems[raw[0]] = Ingredient{
				Options: raw,
				Count:   1,
			}
		}
	}

	var inputs []Ingredient
	for _, ingredient := range groupedItems {
		inputs = append(inputs, ingredient)
	}

	return &ParseJsonResponse{
		Id: dto.Result.Id,
		Recipe: Recipe{
			Yield:  dto.Result.Count,
			Inputs: inputs,
		},
	}, nil
}

func handleShapedRecipe(data []byte) (*ParseJsonResponse, error) {
	var dto struct {
		Key     map[string]RawIngredient `json:"key"`
		Pattern []string                 `json:"pattern"`
		Result  struct {
			Count uint   `json:"count"`
			Id    string `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to parse minecraft shaped recipe: %w", err)
	}

	itemCounts := make(map[string]Ingredient)
	for key, raw := range dto.Key {
		for _, line := range dto.Pattern {
			count := uint(strings.Count(line, key))
			if existing, found := itemCounts[raw[0]]; found {
				existing.Count += count
				itemCounts[raw[0]] = existing
			} else {
				itemCounts[raw[0]] = Ingredient{
					Options: raw,
					Count:   count,
				}
			}
		}
	}

	var inputs []Ingredient
	for _, ingredient := range itemCounts {
		inputs = append(inputs, ingredient)
	}

	return &ParseJsonResponse{
		Id: dto.Result.Id,
		Recipe: Recipe{
			Yield:  dto.Result.Count,
			Inputs: inputs,
		},
	}, nil
}

func handleSmeltingRecipe(data []byte) (*ParseJsonResponse, error) {
	var dto struct {
		Ingredient RawIngredient `json:"ingredient"`
		Result     struct {
			Id string `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	return &ParseJsonResponse{
		Id: dto.Result.Id,
		Recipe: Recipe{
			Yield: 1,
			Inputs: []Ingredient{
				{
					Options: dto.Ingredient,
					Count:   1,
				},
			},
		},
	}, nil
}
