package recipe

import (
	"fmt"
	"maps"
	"strings"
)

type Recipe struct {
	Yield  uint
	Inputs map[string]uint
}

type Inventory map[string]uint

type Tags map[string][]string

type Request struct {
	RequestedItem   string
	RequestedAmount uint
	Recipes         map[string]Recipe
	Tags            Tags
	Inventory       Inventory
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
	tags := req.Tags
	requestedItem := req.RequestedItem
	requestedAmount := req.RequestedAmount

	hasInventory := req.Inventory != nil

	isTag := strings.HasPrefix(requestedItem, "#")
	if isTag {
		requestedTag := requestedItem
		for tag, items := range tags {
			if tag == requestedTag {
				requestedItem = items[0]

				if !hasInventory {
					break
				}

				var bestItem string
				var bestItemAmount uint
				for _, item := range items {
					amount, err := req.Inventory.findAmountOf(item)

					if err == nil && amount > bestItemAmount {
						bestItemAmount = amount
						bestItem = item
					}
				}

				if bestItemAmount != 0 {
					requestedItem = bestItem
				}

				break
			}
		}
	}

	if hasInventory {
		response.Inventory, requestedAmount = grabItemFromInventory(req.Inventory, requestedItem, requestedAmount)
	}

	if requestedAmount == 0 {
		return response
	}

	recipe, hasRecipe := recipes[requestedItem]
	if !hasRecipe {
		response.FlatMap[requestedItem] += requestedAmount
		return response
	}

	multiplier := getRecipeMultiplier(requestedAmount, recipe.Yield)

	totalProduced := multiplier * recipe.Yield
	leftover := totalProduced - requestedAmount

	if leftover > 0 && hasInventory {
		response.Inventory[requestedItem] += leftover
	}

	currentInventory := response.Inventory

	for inputItem, inputAmount := range recipe.Inputs {
		totalNeeded := multiplier * inputAmount
		subResult := ResolveBaseMaterials(Request{
			RequestedItem:   inputItem,
			RequestedAmount: totalNeeded,
			Recipes:         recipes,
			Tags:            tags,
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

func grabItemFromInventory(inventory Inventory, requestedItem string, requestedAmount uint) (Inventory, uint) {
	returnInventory := maps.Clone(inventory)

	ownedAmount, haveItem := returnInventory[requestedItem]
	if !haveItem {
		return returnInventory, requestedAmount
	}

	ownMoreThanRequested := requestedAmount < ownedAmount
	if ownMoreThanRequested {
		returnInventory[requestedItem] -= requestedAmount
	} else if requestedAmount == ownedAmount {
		delete(returnInventory, requestedItem)
	} else {
		requestedAmount -= ownedAmount
		delete(returnInventory, requestedItem)
		return returnInventory, requestedAmount
	}

	return returnInventory, 0
}

func (i Inventory) findAmountOf(requestedItem string) (uint, error) {
	if amountOwned, foundItem := i[requestedItem]; foundItem {
		return amountOwned, nil
	}

	return 0, fmt.Errorf("Inventory does not contain %v", requestedItem)
}
