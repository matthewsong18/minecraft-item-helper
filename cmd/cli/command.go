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
	inventoryPath := flags.String("inventory", "", "JSON object of owned item quantities")
	requestsPath := flags.String("requests", "", "JSON object of requested item quantities")
	showInventory := flags.Bool("show-inventory", false, "show projected inventory after crafting")
	configPath := flags.String("config", "", "JSON settings for preferences and stop items")
	if err := flags.Parse(args); err != nil {
		return err
	}
	amountSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "amount" {
			amountSet = true
		}
	})
	if *requestsPath != "" && (*item != "" || amountSet) {
		return errors.New("cannot combine --requests with --item or --amount")
	}
	if *item == "" && *requestsPath == "" {
		return errors.New("provide --item (for example --item minecraft:torch) or --requests")
	}
	if *amount == 0 {
		return errors.New("--amount must be greater than zero")
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments; use --item and --amount")
	}
	items := map[string]uint{*item: *amount}
	if *requestsPath != "" {
		var err error
		items, err = loadQuantities(*requestsPath)
		if err != nil {
			return err
		}
	}

	config, err := loadSettings(*configPath)
	if err != nil {
		return err
	}
	recipes, err := loadRecipes(*recipesPath, config.RecipeChoices)
	if err != nil {
		return err
	}
	tags, err := loadTags(*tagsPath)
	if err != nil {
		return err
	}

	var inventory internal.Inventory
	if *inventoryPath != "" {
		inventory, err = loadQuantities(*inventoryPath)
		if err != nil {
			return err
		}
	}
	response, err := internal.PlanBuild(internal.BuildRequest{
		Items:   items,
		Recipes: recipes, Tags: tags, Inventory: inventory,
		StopItems: config.stopItems(), Preferences: config.Preferences,
	})
	if err != nil {
		return err
	}
	if err := writeMaterials(output, response.FlatMap); err != nil {
		return err
	}
	if *showInventory {
		if _, err := fmt.Fprintln(output, "\nProjected inventory after crafting (requested items set aside):"); err != nil {
			return err
		}
		return writeMaterials(output, response.Inventory)
	}
	return nil
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
