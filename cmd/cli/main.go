package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"minecraft-item-helper/internal"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("minecraft-item-helper", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	item := flags.String("item", "", "Minecraft item ID to resolve (for example minecraft:torch)")
	amount := flags.Uint("amount", 1, "number of items requested")
	recipesPath := flags.String("recipes", "assets/recipes", "directory containing recipe JSON files")
	tagsPath := flags.String("tags", "assets/item-tag-groups", "directory containing item tag JSON files")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *item == "" {
		return errors.New("missing required --item (for example --item minecraft:torch)")
	}
	if *amount == 0 {
		return errors.New("--amount must be greater than zero")
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments; use --item and --amount")
	}

	recipes, err := loadRecipes(*recipesPath)
	if err != nil {
		return err
	}
	tags, err := loadTags(*tagsPath)
	if err != nil {
		return err
	}

	response := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem:   *item,
		RequestedAmount: *amount,
		Recipes:         recipes,
		Tags:            tags,
	})

	items := make([]string, 0, len(response.FlatMap))
	for id := range response.FlatMap {
		items = append(items, id)
	}
	sort.Strings(items)
	for _, id := range items {
		if _, err := fmt.Fprintf(output, "%s: %d\n", id, response.FlatMap[id]); err != nil {
			return fmt.Errorf("write result: %w", err)
		}
	}
	return nil
}

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
