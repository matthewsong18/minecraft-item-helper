package planner

// StopItemsWithOverrides returns the default stop items with additions and removals applied.
// Removals take precedence when an item appears in both lists.
func StopItemsWithOverrides(add, remove []string) map[string]bool {
	stops := map[string]bool{
		"minecraft:coal":            true,
		"minecraft:copper_ingot":    true,
		"minecraft:diamond":         true,
		"minecraft:emerald":         true,
		"minecraft:gold_ingot":      true,
		"minecraft:iron_ingot":      true,
		"minecraft:lapis_lazuli":    true,
		"minecraft:leather":         true,
		"minecraft:netherite_ingot": true,
		"minecraft:raw_copper":      true,
		"minecraft:raw_gold":        true,
		"minecraft:raw_iron":        true,
		"minecraft:redstone":        true,
		"minecraft:slime_ball":      true,
		"minecraft:wheat":           true,
	}

	for _, itemToAdd := range add {
		stops[itemToAdd] = true
	}

	for _, itemToDel := range remove {
		delete(stops, itemToDel)
	}

	return stops
}
