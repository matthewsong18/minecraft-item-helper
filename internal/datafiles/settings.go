package datafiles

import (
	"fmt"
	"minecraft-item-helper/internal/planner"
	"strings"
)

type Settings struct {
	RecipeChoices map[string]string   `json:"recipe_choices"`
	Preferences   map[string][]string `json:"preferences"`
	StopItems     struct {
		Add    []string `json:"add"`
		Remove []string `json:"remove"`
	} `json:"stop_items"`
}

func LoadSettings(path string) (Settings, error) {
	if path == "" {
		return Settings{}, nil
	}
	var config *Settings
	if err := readJSON(path, &config); err != nil {
		return Settings{}, err
	}
	if config == nil {
		return Settings{}, fmt.Errorf("%q must contain a settings object", path)
	}
	for tag, items := range config.Preferences {
		if !strings.HasPrefix(tag, "#") {
			return Settings{}, fmt.Errorf("preference key %q must be a tag beginning with #", tag)
		}
		for _, item := range items {
			if strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") {
				return Settings{}, fmt.Errorf("preference for %s must name concrete items", tag)
			}
		}
	}
	for _, item := range append(append([]string{}, config.StopItems.Add...), config.StopItems.Remove...) {
		if strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") {
			return Settings{}, fmt.Errorf("stop item %q must be a concrete item", item)
		}
	}
	return *config, nil
}

func (config Settings) StopItemSet() map[string]bool {
	return planner.StopItemsWithOverrides(
		config.StopItems.Add,
		config.StopItems.Remove,
	)
}
