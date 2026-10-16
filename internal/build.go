package internal

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type BuildRequest struct {
	Items       map[string]uint
	Recipes     map[string]Recipe
	Tags        Tags
	Inventory   Inventory
	StopItems   map[string]bool
	Preferences map[string][]string
}

// PlanBuild sets aside requested finished items and shares remaining stock and
// crafting leftovers across requests. It never changes the supplied inventory.
func PlanBuild(req BuildRequest) (Response, error) {
	if len(req.Items) == 0 {
		return Response{}, errors.New("build list is empty")
	}
	inventory := maps.Clone(req.Inventory)
	if inventory == nil {
		inventory = Inventory{}
	}
	remaining := maps.Clone(req.Items)
	items := slices.Sorted(maps.Keys(req.Items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") || req.Items[item] == 0 {
			return Response{}, fmt.Errorf("build item %q must have a concrete item ID and positive quantity", item)
		}
		inventory, remaining[item] = grabItemFromInventory(inventory, item, remaining[item])
	}

	response := Response{FlatMap: map[string]uint{}, Inventory: inventory}
	for _, item := range items {
		if remaining[item] == 0 {
			continue
		}
		result, err := ResolveBaseMaterials(Request{
			RequestedItem: item, RequestedAmount: remaining[item],
			Recipes: req.Recipes, Tags: req.Tags, Inventory: response.Inventory,
			StopItems: req.StopItems, Preferences: req.Preferences, BuildItems: req.Items,
		})
		if err != nil {
			return Response{}, fmt.Errorf("plan %s: %w", item, err)
		}
		appendMaps(response.FlatMap, result.FlatMap)
		response.Inventory = result.Inventory
	}
	return response, nil
}
