package jsonloader

import (
	"encoding/json"
	"fmt"
	"minecraft-item-helper/internal/core/domain"
	"strings"
)

type TagRegistry map[string]domain.Ingredient

func ResolveShapedRecipe(dto CraftingShapedRecipeDTO, tags TagRegistry) (*domain.Recipe, error) {
	ingredients := make(map[string]domain.IngredientAmount)

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

		newIngredientID := newIngredient.GetID()

		if ingredientAmount, exists := ingredients[newIngredientID]; exists {
			ingredientAmount.Amount += amount
			continue
		}

		newIngredientAmountEntry := domain.IngredientAmount{
			Ingredient: newIngredient,
			Amount:     amount,
		}

		ingredients[newIngredientID] = newIngredientAmountEntry
	}

	newRecipe := domain.Recipe{
		ResultID:    dto.Result.ID,
		Yield:       dto.Result.Count,
		Ingredients: ingredients,
	}

	return &newRecipe, nil
}

func ResolveShapelessRecipe(dto CraftingShapelessRecipeDTO, tags TagRegistry) (*domain.Recipe, error) {
	ingredients := make(map[string]domain.IngredientAmount)

	for _, ingredientString := range dto.Ingredients {
		newIngredient, err := resolveIngredient(ingredientString, tags)
		if err != nil {
			return nil, err
		}

		id := newIngredient.GetID()
		entry, exists := ingredients[id]
		if !exists {
			entry = domain.IngredientAmount{
				Ingredient: newIngredient,
				Amount:     0,
			}
		}
		entry.Amount++
		ingredients[id] = entry
	}

	newRecipe := domain.Recipe{
		ResultID:    dto.Result.ID,
		Yield:       dto.Result.Count,
		Ingredients: ingredients,
	}

	return &newRecipe, nil
}

func ResolveSmeltingRecipe(dto SmeltingRecipeDTO, tags TagRegistry) (*domain.Recipe, error) {
	var ingredientsStrings []string
	if err := json.Unmarshal(dto.Ingredient, &ingredientsStrings); err != nil {
		var single string
		if err := json.Unmarshal(dto.Ingredient, &single); err != nil {
			return nil, fmt.Errorf("failed to parse smelting ingredient: %v", err)
		}
		ingredientsStrings = append(ingredientsStrings, single)
	}

	ingredients := make(map[string]domain.IngredientAmount)

	if len(ingredientsStrings) == 0 {
		return nil, fmt.Errorf("smelting recipe has no ingredients")
	}

	firstIngredient, err := resolveIngredient(ingredientsStrings[0], tags)
	if err != nil {
		return nil, err
	}

	ingredients[firstIngredient.GetID()] = domain.IngredientAmount{
		Ingredient: firstIngredient,
		Amount:     1,
	}

	newRecipe := domain.Recipe{
		ResultID:    dto.Result.ID,
		Yield:       dto.Result.Count,
		Ingredients: ingredients,
	}

	return &newRecipe, nil
}

func resolveIngredient(ingredientString string, tags TagRegistry) (domain.Ingredient, error) {
	if tagID, ok := strings.CutPrefix(ingredientString, "#"); ok {
		tag, exists := tags[tagID]
		if !exists {
			return nil, fmt.Errorf("tag reference not found: %s", tagID)
		}
		return tag, nil
	}
	return domain.Item{ID: ingredientString}, nil
}

func ResolveTags(rawTags map[string]TagDTO) (TagRegistry, error) {
	registry := make(TagRegistry)

	// We declare the function signature first so it can call itself recursively
	var resolve func(tagID string) (domain.Ingredient, error)

	resolve = func(tagID string) (domain.Ingredient, error) {
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
		domainTag := &domain.Tag{ID: tagID}
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
				domainTag.Items = append(domainTag.Items, domain.Item{ID: val})
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
