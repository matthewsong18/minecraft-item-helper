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

type Response struct {
	FlatMap   map[string]uint
	Inventory map[string]uint
}

func ResolveBaseMaterials(req Request) Response {
	response := Response{
		FlatMap:   map[string]uint{},
		Inventory: nil,
	}

	recipes := req.Recipes
	item := req.RequestedItem
	amount := req.RequestedAmount

	recipe, ok := recipes[item]
	if !ok {
		response.FlatMap[item] += amount
		return response
	}

	multiplier := getRecipeMultiplier(amount, recipe.Yield)

	for inputItem, inputAmount := range recipe.Inputs {
		totalRequired := multiplier * inputAmount
		materials := ResolveBaseMaterials(Request{
			RequestedItem:   inputItem,
			RequestedAmount: totalRequired,
			Recipes:         recipes,
		})
		appendMaps(response.FlatMap, materials.FlatMap)
	}

	return response
}

func appendMaps(destMap map[string]uint, srcMap map[string]uint) {
	for item, amount := range srcMap {
		destMap[item] += amount
	}
}

func getRecipeMultiplier(target uint, yield uint) uint {
	return (target + yield - 1) / yield
}
