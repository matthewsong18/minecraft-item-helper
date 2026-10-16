package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ParseJsonRecipeResponse struct {
	Id     string
	Recipe Recipe
}

type ParseJsonTagResponse struct {
	Tag []string
}

var ErrUnsupportedRecipeType = errors.New("unsupported recipe type")

func ParseJsonRecipe(data []byte) (*ParseJsonRecipeResponse, error) {
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
	if err := json.Unmarshal(data, &single); err == nil && strings.TrimSpace(single) != "" {
		*r = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil && len(multiple) > 0 {
		for _, option := range multiple {
			if strings.TrimSpace(option) == "" {
				return fmt.Errorf("ingredient option is empty")
			}
		}
		*r = multiple
		return nil
	}

	return fmt.Errorf("invalid ingredient format: %s", string(data))
}

func handleShapelessRecipe(data []byte) (*ParseJsonRecipeResponse, error) {
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

	inputs, err := groupIngredients(dto.Ingredients)
	if err != nil {
		return nil, err
	}

	return &ParseJsonRecipeResponse{
		Id: dto.Result.Id,
		Recipe: Recipe{
			Yield:  dto.Result.Count,
			Inputs: inputs,
		},
	}, nil
}

func handleShapedRecipe(data []byte) (*ParseJsonRecipeResponse, error) {
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

	var rawInputs []RawIngredient
	for _, row := range dto.Pattern {
		for _, symbol := range row {
			if symbol == ' ' {
				continue
			}
			raw, ok := dto.Key[string(symbol)]
			if !ok {
				return nil, fmt.Errorf("unknown shaped recipe symbol %q", symbol)
			}
			rawInputs = append(rawInputs, raw)
		}
	}
	inputs, err := groupIngredients(rawInputs)
	if err != nil {
		return nil, err
	}

	return &ParseJsonRecipeResponse{
		Id: dto.Result.Id,
		Recipe: Recipe{
			Yield:  dto.Result.Count,
			Inputs: inputs,
		},
	}, nil
}

func handleSmeltingRecipe(data []byte) (*ParseJsonRecipeResponse, error) {
	var dto struct {
		Ingredient RawIngredient `json:"ingredient"`
		Result     struct {
			Id string `json:"id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	return &ParseJsonRecipeResponse{
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

func ParseJsonTag(data []byte) (*ParseJsonTagResponse, error) {
	var dto struct {
		Values []string `json:"values"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to parse tag values: %w", err)
	}

	return &ParseJsonTagResponse{Tag: dto.Values}, nil
}

// groupIngredients preserves first appearance, including the order of alternatives.
func groupIngredients(raws []RawIngredient) ([]Ingredient, error) {
	var inputs []Ingredient
	indices := make(map[string]int)
	for _, raw := range raws {
		if len(raw) == 0 {
			return nil, errors.New("ingredient has no options")
		}
		key, _ := json.Marshal(raw)
		if index, ok := indices[string(key)]; ok {
			inputs[index].Count++
		} else {
			indices[string(key)] = len(inputs)
			inputs = append(inputs, Ingredient{Options: raw, Count: 1})
		}
	}
	return inputs, nil
}
