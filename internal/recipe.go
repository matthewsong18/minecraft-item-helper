package recipe

import "maps"

type Recipe struct {
	Yield  uint
	Inputs map[string]uint
}

type Request struct {
	RequestedItem   string
	RequestedAmount uint
	Recipes         map[string]Recipe
	Inventory       map[string]uint
}

type Response struct {
	FlatMap   map[string]uint
	Inventory map[string]uint
}

func ResolveBaseMaterials(req Request) Response {
	response := Response{
		FlatMap:   map[string]uint{},
		Inventory: nil,
	}

	recipes := req.Recipes
	requestedItem := req.RequestedItem
	requestedAmount := req.RequestedAmount

	hasInventory := req.Inventory != nil
	if hasInventory {
		response.Inventory = maps.Clone(req.Inventory)

		if ownedAmount, haveItem := req.Inventory[requestedItem]; haveItem {
			if requestedAmount < ownedAmount {
				req.Inventory[requestedItem] -= requestedAmount
				response.Inventory = maps.Clone(req.Inventory)
				return response
			} else if requestedAmount == ownedAmount {
				delete(response.Inventory, requestedItem)
				return response
			}

			requestedAmount -= ownedAmount
			delete(response.Inventory, requestedItem)
		}
	}

	recipe, hasRecipe := recipes[requestedItem]
	if !hasRecipe {
		response.FlatMap[requestedItem] += requestedAmount
		return response
	}

	multiplier := getRecipeMultiplier(requestedAmount, recipe.Yield)

	totalProduced := multiplier * recipe.Yield
	if leftover := totalProduced - requestedAmount; leftover > 0 && hasInventory {
		response.Inventory[requestedItem] += leftover
	}

	currentInventory := response.Inventory

	for inputItem, inputAmount := range recipe.Inputs {
		totalNeeded := multiplier * inputAmount
		subResult := ResolveBaseMaterials(Request{
			RequestedItem:   inputItem,
			RequestedAmount: totalNeeded,
			Recipes:         recipes,
			Inventory:       currentInventory,
		})

		appendMaps(response.FlatMap, subResult.FlatMap)
		currentInventory = subResult.Inventory
	}

	response.Inventory = currentInventory

	return response
}

func appendMaps(destMap map[string]uint, srcMap map[string]uint) {
	for item, amount := range srcMap {
		destMap[item] += amount
	}
}

func getRecipeMultiplier(target uint, yield uint) uint {
	return (target + yield - 1) / yield
}
