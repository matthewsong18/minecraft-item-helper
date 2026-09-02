package domain

type ItemList struct{ itemStacks map[ItemID]Amount }

func ParseItemList(itemStacks map[ItemID]Amount) *ItemList {
	return &ItemList{
		itemStacks: itemStacks,
	}
}

type Inventory struct{ itemStacks map[ItemID]Amount }

func ParseInventory(itemStacks map[ItemID]Amount) *Inventory {
	return &Inventory{
		itemStacks: itemStacks,
	}
}

type RawTypes struct{ items map[ItemID]bool }

func ParseRawTypes(items map[ItemID]bool) *RawTypes {
	return &RawTypes{
		items: items,
	}
}

type PlanCalculation struct {
	ItemList   ItemList
	Inventory  Inventory
	RawTypes   RawTypes
	RecipeRepo RecipeRepository
	TagRepo    TagRepository
}

func NewPlanCalculation(iL ItemList, i Inventory, r RawTypes, recipeRepo RecipeRepository, tagRepo TagRepository) PlanCalculation {
	return PlanCalculation{
		ItemList:   iL,
		Inventory:  i,
		RawTypes:   r,
		RecipeRepo: recipeRepo,
		TagRepo:    tagRepo,
	}
}

func (c PlanCalculation) Execute() error {
	panic("Unimplemented")
}
