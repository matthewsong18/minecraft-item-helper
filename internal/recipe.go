package main

import "maps"

type Recipe struct {
	Yield  int
	Inputs map[string]int
}

func ResolveBaseMaterials(inputs map[string]int, recipes map[string]Recipe) map[string]int {
	baseMaterials := make(map[string]int)
	for item := range inputs {
		recipe := recipes[item]

		maps.Copy(baseMaterials, recipe.Inputs)
	}
	return baseMaterials
}
