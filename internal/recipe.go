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
	Preferences     map[string][]string
	BuildItems      map[string]uint
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
	requestedItem := req.RequestedItem
	requestedAmount := req.RequestedAmount

	hasInventory := req.Inventory != nil

	if strings.HasPrefix(requestedItem, "#") {
		var err error
		requestedItem, err = chooseIngredient([]string{requestedItem}, req)
		if err != nil {
			return Response{}, err
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
		item, err := chooseIngredient(ingredient.Options, reqWithInventory(req, currentInventory))
		if err != nil {
			return Response{}, err
		}
		subResult, err := resolve(Request{
			RequestedItem:   item,
			RequestedAmount: totalNeeded,
			Recipes:         recipes,
			Tags:            req.Tags,
			Preferences:     req.Preferences,
			BuildItems:      req.BuildItems,
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
