package main

import (
	"errors"
	"fmt"
	"minecraft-item-helper/internal"
	"os"
	"path/filepath"
	"strings"
)

func loadRecipes(dir string) (map[string]internal.Recipe, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read recipes directory %q: %w", dir, err)
	}

	recipes := make(map[string]internal.Recipe)
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
		recipes[parsed.Id] = parsed.Recipe
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
