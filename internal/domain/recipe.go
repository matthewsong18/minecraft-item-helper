package domain

import "context"

type ItemID string
type Amount uint

type ItemStack struct {
	Item   ItemID
	Amount Amount
}

type RecipeInput interface {
	isRecipeInput()
	Amount() Amount
}

type ItemIngredient struct {
	ID    ItemID
	Count Amount
}

func (ItemIngredient) isRecipeInput()   {}
func (i ItemIngredient) Amount() Amount { return i.Count }

var _ RecipeInput = (*ItemIngredient)(nil)

type TagIngredient struct {
	Items []ItemID
	Count Amount
}

func (TagIngredient) isRecipeInput()   {}
func (t TagIngredient) Amount() Amount { return t.Count }

var _ RecipeInput = (*TagIngredient)(nil)

type Recipe struct {
	Inputs []RecipeInput
	Output ItemStack
}

type RecipeRepository interface {
	FindRecipesFor(ctx context.Context, item ItemID) ([]Recipe, error)
}

type TagID string

type TagRepository interface {
	GetItemsIn(ctx context.Context, tag TagID) ([]ItemID, error)
}
