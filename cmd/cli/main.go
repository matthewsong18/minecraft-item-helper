package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"minecraft-item-helper/internal"
	"os"
)

func main() {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	recipesPath := flag.String("recipes", "assets/recipes/", "path to find recipe jsons")
	// tagsPath := flag.String("tags", "assets/item-tag-groups/", "path to find tag jsons")

	flag.Parse()

	files, err := os.ReadDir(*recipesPath)
	if err != nil {
		log.Fatal(err)
	}

	recipes := make(map[string]internal.Recipe)

	for _, file := range files {
		data, err := os.ReadFile(*recipesPath + file.Name())
		if err != nil {
			log.Fatal(err)
		}

		parseJsonResponse, err := internal.ParseJsonRecipe(data)
		if err != nil {
			if errors.Is(err, internal.ErrUnsupportedRecipeType) {
				continue
			}
			logger.Printf("%v: %v", file.Name(), err)
			continue
		}

		itemId := parseJsonResponse.Id
		recipe := parseJsonResponse.Recipe

		recipes[itemId] = recipe
	}

	fmt.Print(&buf)
}
