package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("parse %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("parse %q: expected a single JSON value", path)
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
