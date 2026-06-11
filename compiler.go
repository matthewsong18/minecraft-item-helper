package main

import (
	"fmt"
	"strings"
)

type TagRegistry map[string]Ingredient

func ResolveShapedRecipe(dto CraftingShapedRecipeDTO, tags TagRegistry) (*Recipe, error) {
	ingredients := make(map[string]IngredientAmount)

	for key, ingredientString := range dto.Key {
		var amount int
		for _, line := range dto.Pattern {
			amount += strings.Count(line, key)
		}
		if amount == 0 {
			return nil, fmt.Errorf("key for %s not found in pattern: %s", ingredientString, key)
		}

		newIngredient, err := resolveIngredient(ingredientString, tags)
		if err != nil {
			return nil, err
		}

		newIngredientID := newIngredient.getID()

		if ingredientAmount, exists := ingredients[newIngredientID]; exists {
			ingredientAmount.Amount += amount
			continue
		}

		newIngredientAmountEntry := IngredientAmount{
			Ingredient: newIngredient,
			Amount:     amount,
		}

		ingredients[newIngredientID] = newIngredientAmountEntry
	}

	newRecipe := Recipe{
		ResultID:    dto.Result.ID,
		Yield:       dto.Result.Count,
		Ingredients: ingredients,
	}

	return &newRecipe, nil
}

func resolveIngredient(ingredientString string, tags TagRegistry) (Ingredient, error) {
	if tagID, ok := strings.CutPrefix(ingredientString, "#"); ok {
		tag, exists := tags[tagID]
		if !exists {
			return nil, fmt.Errorf("tag reference not found: %s", tagID)
		}
		return tag, nil
	}
	return Item{ID: ingredientString}, nil
}

func ResolveTags(rawTags map[string]TagDTO) (TagRegistry, error) {
	registry := make(TagRegistry)

	// We declare the function signature first so it can call itself recursively
	var resolve func(tagID string) (Ingredient, error)

	resolve = func(tagID string) (Ingredient, error) {
		// 1. If we already resolved this tag, just return the cached pointer!
		if resolvedTag, exists := registry[tagID]; exists {
			return resolvedTag, nil
		}

		// 2. Fetch the raw DTO
		dto, exists := rawTags[tagID]
		if !exists {
			return nil, fmt.Errorf("tag reference not found: %s", tagID)
		}

		// 3. Create the Domain Tag. We put it in the registry early
		// to prevent infinite loops if two tags reference each other.
		domainTag := &Tag{ID: tagID}
		registry[tagID] = domainTag

		// 4. Resolve all its children (Items and Sub-Tags)
		for _, val := range dto.Values {
			if after, ok := strings.CutPrefix(val, "#"); ok {
				// It's a nested tag! Recursively resolve it.
				subTagID := after
				subTag, err := resolve(subTagID)
				if err != nil {
					return nil, err
				}
				domainTag.Items = append(domainTag.Items, subTag)
			} else {
				// It's just a regular item
				domainTag.Items = append(domainTag.Items, Item{ID: val})
			}
		}

		return domainTag, nil
	}

	// 5. Loop through all raw DTOs and trigger the resolver
	for tagID := range rawTags {
		if _, err := resolve(tagID); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
