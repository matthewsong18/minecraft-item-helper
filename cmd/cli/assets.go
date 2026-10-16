package main

import (
	"errors"
	"fmt"
	"minecraft-item-helper/internal"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func loadRecipes(dir string, choices map[string]string) (map[string]internal.Recipe, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read recipes directory %q: %w", dir, err)
	}

	type candidate struct {
		name   string
		recipe internal.Recipe
	}
	candidates := make(map[string][]candidate)
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read recipe %q: %w", path, err)
		}
		parsed, err := internal.ParseJsonRecipe(data)
		if errors.Is(err, internal.ErrUnsupportedRecipeType) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("parse recipe %q: %w", path, err)
		}
		name := strings.TrimSuffix(file.Name(), ".json")
		candidates[parsed.Id] = append(candidates[parsed.Id], candidate{name, parsed.Recipe})
	}
	for item, name := range choices {
		if !slices.ContainsFunc(candidates[item], func(c candidate) bool { return c.name == name }) {
			return nil, fmt.Errorf("recipe choice %q for %s is not a supported recipe in %q", name, item, dir)
		}
	}
	recipes := make(map[string]internal.Recipe)
	for item, options := range candidates {
		selected := options[0]
		wanted := choices[item]
		if wanted == "" {
			wanted = strings.TrimPrefix(item, "minecraft:")
		}
		for _, option := range options {
			if option.name == wanted {
				selected = option
				break
			}
		}
		recipes[item] = selected.recipe
	}
	return recipes, nil
}

func loadTags(dir string) (internal.Tags, error) {
	tags := make(internal.Tags)
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read tag %q: %w", path, err)
		}
		parsed, err := internal.ParseJsonTag(data)
		if err != nil {
			return fmt.Errorf("parse tag %q: %w", path, err)
		}

		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("resolve tag path %q: %w", path, err)
		}
		tagName := strings.TrimSuffix(filepath.ToSlash(relativePath), ".json")
		tags["#minecraft:"+tagName] = parsed.Tag
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read tags directory %q: %w", dir, err)
	}
	return tags, nil
}
