package main

import (
	"fmt"
	"minecraft-item-helper/internal"
	"strings"
)

type settings struct {
	RecipeChoices map[string]string   `json:"recipe_choices"`
	Preferences   map[string][]string `json:"preferences"`
	StopItems     struct {
		Add    []string `json:"add"`
		Remove []string `json:"remove"`
	} `json:"stop_items"`
}

func loadSettings(path string) (settings, error) {
	if path == "" {
		return settings{}, nil
	}
	var config *settings
	if err := readJSON(path, &config); err != nil {
		return settings{}, err
	}
	if config == nil {
		return settings{}, fmt.Errorf("%q must contain a settings object", path)
	}
	for tag, items := range config.Preferences {
		if !strings.HasPrefix(tag, "#") {
			return settings{}, fmt.Errorf("preference key %q must be a tag beginning with #", tag)
		}
		for _, item := range items {
			if strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") {
				return settings{}, fmt.Errorf("preference for %s must name concrete items", tag)
			}
		}
	}
	for _, item := range append(append([]string{}, config.StopItems.Add...), config.StopItems.Remove...) {
		if strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") {
			return settings{}, fmt.Errorf("stop item %q must be a concrete item", item)
		}
	}
	return *config, nil
}

func (config settings) stopItems() map[string]bool {
	return internal.StopItemsWithOverrides(
		config.StopItems.Add,
		config.StopItems.Remove,
	)
}
