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

func handleShapelessRecipe(data []byte) (*ParseJsonResponse, error) {
	var dto struct {
		Ingredients []string `json:"ingredients"`
		Result      struct {
			Count uint   `json:"count"`
			Id    string `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	inputs := make(map[string]uint)
	for _, item := range dto.Ingredients {
		inputs[item] += 1
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
		Key     map[string]string `json:"key"`
		Pattern []string          `json:"pattern"`
		Result  struct {
			Count uint   `json:"count"`
			Id    string `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to parse minecraft shaped recipe: %w", err)
	}

	inputs := make(map[string]uint)
	for key, item := range dto.Key {
		for _, line := range dto.Pattern {
			count := strings.Count(line, key)
			inputs[item] += uint(count)
		}
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
		Ingredient string `json:"ingredient"`
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
			Inputs: map[string]uint{
				dto.Ingredient: 1,
			},
		},
	}, nil
}
