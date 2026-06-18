package main

import (
	"cmp"
	"fmt"
	"slices"
)

func CalculateRecipe(targetID string, targetAmount int, recipes map[string]*Recipe, tags TagRegistry, path []string) (*Node, error) {
	// 1. Check for cycles
	if slices.Contains(path, targetID) {
		// If we've already seen this item in the current recursion path,
		// treat it as a base material to avoid infinite loops.
		return &Node{
			Item:         targetID,
			TargetAmount: targetAmount,
		}, nil
	}

	// base case: a recipe doesn't exist for the target, hence a leaf node
	node := Node{
		Item:             targetID,
		Recipe:           nil,
		TargetAmount:     targetAmount,
		RecipeMultiplier: 0,
		Children:         nil,
	}

	recipe, exists := recipes[targetID]
	if !exists {
		return &node, nil
	}

	node.Recipe = recipe

	// Calculate how many times we need to run the recipe
	recipeMultiplier := (targetAmount + recipe.Yield - 1) / recipe.Yield
	node.RecipeMultiplier = recipeMultiplier

	// Update the path for child calculations
	newPath := append(path, targetID)

	children := []Node{}
	for _, ingredientAmount := range recipe.Ingredients {
		var itemID string
		switch ing := ingredientAmount.Ingredient.(type) {
		case Item:
			itemID = ing.ID
		case *Tag:
			items := ing.getItems()
			if len(items) == 0 {
				return nil, fmt.Errorf("tag %s has no items", ing.ID)
			}
			itemID = items[0]
		default:
			return nil, fmt.Errorf("unsupported ingredient type: %T", ing)
		}

		itemAmountNeeded := ingredientAmount.Amount * recipeMultiplier

		childNode, err := CalculateRecipe(itemID, itemAmountNeeded, recipes, tags, newPath)
		if err != nil {
			return nil, err
		}
		children = append(children, *childNode)
	}

	slices.SortFunc(children, func(a, b Node) int {
		return cmp.Compare(a.Item, b.Item)
	})
	node.Children = children

	return &node, nil
}
