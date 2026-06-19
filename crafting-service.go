package main

import (
	"cmp"
	"fmt"
	"slices"
)

type ItemRequest struct {
	ItemID string `json:"item_id"`
	Amount int    `json:"amount"`
}

type CraftingResult struct {
	Trees         []*Node        `json:"trees"`
	BaseMaterials map[string]int `json:"base_materials"`
}

type craftingService struct {
	state   CraftingState
	recipes map[string]*Recipe
	tags    TagRegistry
}

func NewCraftingService(recipes map[string]*Recipe, tags TagRegistry) *craftingService {
	return &craftingService{
		recipes: recipes,
		tags:    tags,
	}
}

func (s *craftingService) ComputeDependencies(itemRequests []*ItemRequest) (*CraftingResult, error) {
	craftingResult := CraftingResult{
		Trees:         []*Node{},
		BaseMaterials: map[string]int{},
	}

	for _, itemRequest := range itemRequests {
		var trackingPath []string
		itemID := itemRequest.ItemID
		amount := itemRequest.Amount
		itemTreeResult, err := CalculateRecipeTree(itemID, amount, s.recipes, s.tags, trackingPath)
		if err != nil {
			return nil, err
		}
		craftingResult.Trees = append(craftingResult.Trees, itemTreeResult)

		queue := []*Node{itemTreeResult}
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]

			if node == nil {
				continue
			}

			for i := range node.Children {
				queue = append(queue, &node.Children[i])
			}

			if node.RecipeMultiplier == 0 {
				craftingResult.BaseMaterials[node.Item] += node.TargetAmount
			}
		}
	}

	return &craftingResult, nil
}

func CalculateRecipeTree(targetID string, targetAmount int, recipes map[string]*Recipe, tags TagRegistry, path []string) (*Node, error) {
	// Check for cycles
	if slices.Contains(path, targetID) {
		// If we've already seen this item in the current recursion path,
		// treat it as a base material to avoid infinite loops.
		return nil, nil
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

		childNode, err := CalculateRecipeTree(itemID, itemAmountNeeded, recipes, tags, newPath)
		if err != nil {
			return nil, err
		}
		if childNode != nil {
			children = append(children, *childNode)
		}
	}

	slices.SortFunc(children, func(a, b Node) int {
		return cmp.Compare(a.Item, b.Item)
	})
	node.Children = children

	return &node, nil
}
