package internal

import (
	"encoding/json"
)

type Shapeless struct {
	Ingredients []string `json:"ingredients"`
	Result      struct {
		Count uint   `json:"count"`
		Id    string `json:"id"`
	} `json:"result"`
}

func ParseJsonRecipe(data []byte) (*Recipe, error) {
	var dto Shapeless
	err := json.Unmarshal(data, &dto)
	if err != nil {
		return nil, err
	}

	ing := dto.Ingredients[0]

	return &Recipe{
		Yield: dto.Result.Count,
		Inputs: map[string]uint{
			ing: 1,
		},
	}, nil
}
