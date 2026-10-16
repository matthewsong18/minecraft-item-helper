package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"minecraft-item-helper/internal"
	"sort"
)

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

	response, err := internal.ResolveBaseMaterials(internal.Request{
		RequestedItem:   *item,
		RequestedAmount: *amount,
		Recipes:         recipes,
		Tags:            tags,
	})
	if err != nil {
		return err
	}
	return writeMaterials(output, response.FlatMap)
}

func writeMaterials(output io.Writer, materials map[string]uint) error {
	items := make([]string, 0, len(materials))
	for id := range materials {
		items = append(items, id)
	}
	sort.Strings(items)
	for _, id := range items {
		if _, err := fmt.Fprintf(output, "%s: %d\n", id, materials[id]); err != nil {
			return fmt.Errorf("write result: %w", err)
		}
	}
	return nil
}
