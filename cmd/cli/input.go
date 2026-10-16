package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("parse %q: %w", path, err)
	}
	return nil
}

func loadQuantities(path string) (map[string]uint, error) {
	var quantities map[string]uint
	if err := readJSON(path, &quantities); err != nil {
		return nil, err
	}
	if quantities == nil {
		return nil, fmt.Errorf("%q must contain an object of item IDs and quantities", path)
	}
	for item := range quantities {
		if strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") {
			return nil, fmt.Errorf("invalid concrete item ID %q in %q", item, path)
		}
	}
	return quantities, nil
}
