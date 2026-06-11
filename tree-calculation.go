package main

import (
	"cmp"
	"fmt"
	"slices"
)

func CalculateRecipe(targetID string, targetAmount int, recipes map[string]*Recipe, tags TagRegistry) (*Node, error) {
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

	recipeMultiplier := (targetAmount + recipe.Yield - 1) / recipe.Yield
	node.RecipeMultiplier = recipeMultiplier

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

		childNode, err := CalculateRecipe(itemID, itemAmountNeeded, recipes, tags)
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
