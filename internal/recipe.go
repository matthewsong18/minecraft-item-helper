package recipe

type Recipe struct {
	Yield  int
	Inputs map[string]int
}

func ResolveBaseMaterials(item string, amount int, recipes map[string]Recipe) map[string]int {
	baseMaterials := make(map[string]int)
	recipe := recipes[item]

	if len(recipe.Inputs) <= 0 {
		baseMaterials[item] += amount
		return baseMaterials
	}

	for item, amount := range recipe.Inputs {
		materials := ResolveBaseMaterials(item, amount, recipes)
		for item, amount := range materials {
			baseMaterials[item] += amount
		}
	}

	return baseMaterials
}
