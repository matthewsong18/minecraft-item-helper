package internal

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Ingredient struct {
	Options []string // Some recipes have OR options like coal or charcoal for torches
	Count   uint
}

type Recipe struct {
	Yield  uint
	Inputs []Ingredient
}

type Inventory map[string]uint

type Tags map[string][]string

type Request struct {
	RequestedItem   string
	RequestedAmount uint
	Recipes         map[string]Recipe
	Tags            Tags
	Inventory       Inventory
	StopItems       map[string]bool
}

type Response struct {
	FlatMap   map[string]uint
	Inventory map[string]uint
}

var ErrRecipeCycle = errors.New("recipe cycle")

func ResolveBaseMaterials(req Request) (Response, error) {
	return resolve(req, nil)
}

func resolve(req Request, path []string) (Response, error) {
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
		return response, nil
	}

	recipe, hasRecipe := recipes[requestedItem]
	if !hasRecipe || req.StopItems[requestedItem] {
		response.FlatMap[requestedItem] += requestedAmount
		return response, nil
	}
	if slices.Contains(path, requestedItem) {
		return Response{}, fmt.Errorf("%w: %s", ErrRecipeCycle, strings.Join(append(path, requestedItem), " -> "))
	}
	if recipe.Yield == 0 {
		return Response{}, fmt.Errorf("recipe for %s has zero yield", requestedItem)
	}
	path = append(path, requestedItem)

	multiplier := getRecipeMultiplier(requestedAmount, recipe.Yield)

	totalProduced := multiplier * recipe.Yield
	leftover := totalProduced - requestedAmount

	currentInventory := response.Inventory

	for _, ingredient := range recipe.Inputs {
		totalNeeded := multiplier * ingredient.Count
		if len(ingredient.Options) == 0 {
			return Response{}, fmt.Errorf("recipe for %s has an ingredient without options", requestedItem)
		}
		subResult, err := resolve(Request{
			RequestedItem:   ingredient.Options[0],
			RequestedAmount: totalNeeded,
			Recipes:         recipes,
			Tags:            tags,
			Inventory:       currentInventory,
			StopItems:       req.StopItems,
		}, path)
		if err != nil {
			return Response{}, err
		}

		appendMaps(response.FlatMap, subResult.FlatMap)
		currentInventory = subResult.Inventory
	}

	response.Inventory = currentInventory
	if leftover > 0 && hasInventory {
		response.Inventory[requestedItem] += leftover
	}

	return response, nil
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
