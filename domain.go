package main

type Recipe struct {
	ResultID    string
	Yield       int
	Ingredients map[string]IngredientAmount
}

type IngredientAmount struct {
	Ingredient Ingredient
	Amount     int
}

type Ingredient interface {
	getID() string
	Matches(itemID string) bool
	getItems() []string
}

type Item struct{ ID string }

func (i Item) getID() string {
	return i.ID
}

// Matches implements [Ingredient].
func (i Item) Matches(itemID string) bool {
	return i.ID == itemID
}

// getItems implements [Ingredient].
func (i Item) getItems() []string {
	return []string{i.ID}
}

type Tag struct {
	ID    string
	Items []Ingredient
}

// getID implements [Ingredient].
func (t *Tag) getID() string {
	return t.ID
}

// Matches implements [Ingredient].
func (t *Tag) Matches(itemID string) bool {
	for _, item := range t.Items {
		if item.Matches(itemID) {
			return true
		}
	}
	return false
}

// getItems implements [Ingredient].
func (t *Tag) getItems() []string {
	var items []string
	for _, child := range t.Items {
		items = append(items, child.getItems()...)
	}
	return items
}

type CraftingState struct {
	Remainders   map[string]int
	RawMaterials map[string]int
}

type Node struct {
	Item             string
	Recipe           *Recipe
	TargetAmount     int
	RecipeMultiplier int
	Children         []Node
}

var _ Ingredient = (*Item)(nil)
var _ Ingredient = (*Tag)(nil)
