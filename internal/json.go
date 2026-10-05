package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Shapeless struct {
	Ingredients []string `json:"ingredients"`
	Result      struct {
		Count uint   `json:"count"`
		Id    string `json:"id"`
	} `json:"result"`
}

type Shaped struct {
	Key     map[string]string `json:"key"`
	Pattern []string          `json:"pattern"`
	Result  struct {
		Count uint   `json:"count"`
		Id    string `json:"id"`
	} `json:"result"`
}

func ParseJsonRecipe(data []byte) (*Recipe, error) {
	var router struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &router); err != nil {
		return nil, fmt.Errorf("failed to parse message type: %w", err)
	}

	switch router.Type {
	case "minecraft:crafting_shapeless":
		var dto Shapeless
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}

		inputs := make(map[string]uint)
		for _, item := range dto.Ingredients {
			inputs[item] += 1
		}

		return &Recipe{
			Yield:  dto.Result.Count,
			Inputs: inputs,
		}, nil
	case "minecraft:crafting_shaped":
		return handleShapedRecipe(data)
	default:
		return nil, fmt.Errorf("unknown recipe type: %s", router.Type)
	}
}

func handleShapedRecipe(data []byte) (*Recipe, error) {
	var dto Shaped
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

	return &Recipe{
		Yield:  dto.Result.Count,
		Inputs: inputs,
	}, nil
}
