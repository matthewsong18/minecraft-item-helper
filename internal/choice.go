package internal

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var ErrTagCycle = errors.New("tag cycle")

func reqWithInventory(req Request, inventory Inventory) Request {
	req.Inventory = inventory
	return req
}

func chooseIngredient(options []string, req Request) (string, error) {
	candidates, err := expandOptions(options, req.Tags, nil)
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", errors.New("ingredient has no options")
	}

	var preferred []string
	for _, option := range options {
		preferred = append(preferred, req.Preferences[option]...)
	}
	rank := func(item string) int {
		if index := slices.Index(preferred, item); index >= 0 {
			return index
		}
		return len(preferred)
	}
	best := candidates[0]
	for _, item := range candidates[1:] {
		itemInBuild, bestInBuild := req.BuildItems[item] > 0, req.BuildItems[best] > 0
		if itemInBuild != bestInBuild {
			if itemInBuild {
				best = item
			}
			continue
		}
		if req.Inventory[item] > req.Inventory[best] ||
			(req.Inventory[item] == req.Inventory[best] && rank(item) < rank(best)) {
			best = item
		}
	}
	return best, nil
}

func expandOptions(options []string, tags Tags, path []string) ([]string, error) {
	var items []string
	for _, option := range options {
		if !strings.HasPrefix(option, "#") {
			if !slices.Contains(items, option) {
				items = append(items, option)
			}
			continue
		}
		if slices.Contains(path, option) {
			return nil, fmt.Errorf("%w: %s", ErrTagCycle, strings.Join(append(path, option), " -> "))
		}
		values := tags[option]
		if len(values) == 0 {
			return nil, fmt.Errorf("tag %s is missing or empty", option)
		}
		expanded, err := expandOptions(values, tags, append(path, option))
		if err != nil {
			return nil, err
		}
		for _, item := range expanded {
			if !slices.Contains(items, item) {
				items = append(items, item)
			}
		}
	}
	return items, nil
}
