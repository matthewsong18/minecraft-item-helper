package crafting

import (
	"context"
	"minecraft-item-helper/internal/domain"
)

type CraftService struct {
	recipeRepo domain.RecipeRepository
	tagRepo    domain.TagRepository
}

func NewCraftService(r domain.RecipeRepository, t domain.TagRepository) *CraftService {
	return &CraftService{
		recipeRepo: r,
		tagRepo:    t,
	}
}

type CalculateItemTreeRequest struct {
	DesiredItems   map[domain.ItemID]domain.Amount // Map of desired items and amount
	Inventory      map[domain.ItemID]domain.Amount // (Optional) Map of items already owned or gathered as leftovers during crafting
	CustomRawTypes map[domain.ItemID]bool          // (Optional) Map of items to mark as new endpoints in calculation or overrides for existing endpoints
}

type CalculateItemTreeResponse struct {
	Root       *domain.TreeNode
	FlatMap    map[domain.ItemID]domain.Amount // Map of necessary items and amounts to fulfill the request
	Remainders map[domain.ItemID]domain.Amount // Map of items remaining from crafting process
}

func (c *CraftService) CalculateItemTree(ctx context.Context, req CalculateItemTreeRequest) (CalculateItemTreeResponse, error) {
	itemList := domain.ParseItemList(req.DesiredItems)
	inventory := domain.ParseInventory(req.Inventory)
	rawTypes := domain.ParseRawTypes(req.CustomRawTypes)
	planCalculation := domain.PlanCalculation{
		ItemList:   *itemList,
		Inventory:  *inventory,
		RawTypes:   *rawTypes,
		RecipeRepo: c.recipeRepo,
		TagRepo:    c.tagRepo,
	}

	treeResult, err := planCalculation.Execute(ctx)
	if err != nil {
		return CalculateItemTreeResponse{}, err
	}

	response := CalculateItemTreeResponse{
		Root:       treeResult.Root,
		FlatMap:    treeResult.FlatMap,
		Remainders: treeResult.Remainders,
	}

	return response, nil
}
