package recipe

type Recipe struct {
	Yield  int
	Inputs map[string]int
}

func ResolveBaseMaterials(item string, amount int, recipes map[string]Recipe) map[string]int {
	baseMaterials := make(map[string]int)

	recipe, ok := recipes[item]
	if !ok {
		baseMaterials[item] += amount
		return baseMaterials
	}

	multiplier := getRecipeMultiplier(amount, recipe.Yield)

	for inputItem, inputAmount := range recipe.Inputs {
		materials := ResolveBaseMaterials(inputItem, inputAmount, recipes)
		multipliedMaterials := applyMultiplier(materials, multiplier)
		appendMaps(baseMaterials, multipliedMaterials)
	}

	return baseMaterials
}

func applyMultiplier(materials map[string]int, multiplier int) map[string]int {
	multipliedMaterials := make(map[string]int)
	for mItem, mAmount := range materials {
		multipliedMaterials[mItem] = mAmount * multiplier
	}
	return multipliedMaterials
}

func appendMaps(destMap map[string]int, srcMap map[string]int) {
	for item, amount := range srcMap {
		destMap[item] += amount
	}
}

func getRecipeMultiplier(target int, yield int) int {
	return (target + yield - 1) / yield
}
