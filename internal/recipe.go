package recipe

type Recipe struct {
	Yield  uint
	Inputs map[string]uint
}

type Request struct {
	RequestedItem   string
	RequestedAmount uint
	Recipes         map[string]Recipe
}

func ResolveBaseMaterials(req Request) map[string]uint {
	baseMaterials := make(map[string]uint)

	recipes := req.Recipes
	item := req.RequestedItem
	amount := req.RequestedAmount

	recipe, ok := recipes[item]
	if !ok {
		baseMaterials[item] += amount
		return baseMaterials
	}

	multiplier := getRecipeMultiplier(amount, recipe.Yield)

	for inputItem, inputAmount := range recipe.Inputs {
		totalRequired := multiplier * inputAmount
		materials := ResolveBaseMaterials(Request{
			RequestedItem:   inputItem,
			RequestedAmount: totalRequired,
			Recipes:         recipes,
		})
		appendMaps(baseMaterials, materials)
	}

	return baseMaterials
}

func appendMaps(destMap map[string]uint, srcMap map[string]uint) {
	for item, amount := range srcMap {
		destMap[item] += amount
	}
}

func getRecipeMultiplier(target uint, yield uint) uint {
	return (target + yield - 1) / yield
}
