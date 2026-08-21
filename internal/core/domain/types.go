package domain

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
	GetID() string
	Matches(itemID string) bool
	GetItems() []string
}

type Item struct{ ID string }

func (i Item) GetID() string {
	return i.ID
}

// Matches implements [Ingredient].
func (i Item) Matches(itemID string) bool {
	return i.ID == itemID
}

// GetItems implements [Ingredient].
func (i Item) GetItems() []string {
	return []string{i.ID}
}

type Tag struct {
	ID    string
	Items []Ingredient
}

// getID implements [Ingredient].
func (t *Tag) GetID() string {
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

// GetItems implements [Ingredient].
func (t *Tag) GetItems() []string {
	var items []string
	for _, child := range t.Items {
		items = append(items, child.GetItems()...)
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
