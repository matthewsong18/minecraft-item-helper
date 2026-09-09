package domain

import "context"

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

// TreeNode represents a node in the crafting tree
type TreeNode struct {
	Item     ItemID
	Amount   Amount
	Recipe   *Recipe
	Children []*TreeNode
}

// TreeResult represents the complete crafting tree result
type TreeResult struct {
	Root     *TreeNode
	FlatMap  map[ItemID]Amount
	Remainders map[ItemID]Amount
}

func (c PlanCalculation) Execute(ctx context.Context) (*TreeResult, error) {
	seenRecipes := make(map[ItemID]Recipe)
	itemQueue := make([]ItemID, 0)
	
	// Initialize queue with desired items
	for itemID := range c.ItemList.itemStacks {
		itemQueue = append(itemQueue, itemID)
	}
	
	// Build the tree structure
	treeNodes := make(map[ItemID]*TreeNode)
	
	for len(itemQueue) > 0 {
		itemID := itemQueue[0]
		itemQueue = itemQueue[1:]

		recipe, seen := seenRecipes[itemID]
		if !seen {
			recipes, err := c.RecipeRepo.FindRecipesFor(ctx, itemID)
			if err != nil {
				return nil, err
			}

			if len(recipes) == 0 {
				// No recipe found, this is a raw item
				continue
			}
			
			recipe = recipes[0]
			seenRecipes[itemID] = recipe
		}

		// Create or get the tree node for this item
		node, exists := treeNodes[itemID]
		if !exists {
			node = &TreeNode{
				Item:   itemID,
				Amount: c.ItemList.itemStacks[itemID],
			}
			treeNodes[itemID] = node
		}
		
		node.Recipe = &recipe
		
		// Process inputs and create child nodes
		for _, input := range recipe.Inputs {
			var inputItemID ItemID
			var inputAmount Amount
			
			switch ing := input.(type) {
			case ItemIngredient:
				inputItemID = ing.ID
				inputAmount = ing.Count
			case TagIngredient:
				// For tag ingredients, we'll use the first item in the tag
				items, err := c.TagRepo.GetItemsIn(ctx, TagID(ing.Items[0]))
				if err != nil {
					return nil, err
				}
				if len(items) > 0 {
					inputItemID = items[0]
					inputAmount = ing.Count
				}
			}
			
			if inputItemID != "" {
				// Create child node
				childNode := &TreeNode{
					Item:   inputItemID,
					Amount: inputAmount,
				}
				node.Children = append(node.Children, childNode)
				treeNodes[inputItemID] = childNode
				
				// Add to queue for processing
				itemQueue = append(itemQueue, inputItemID)
			}
		}
	}
	
	// Build flat map and remainders
	flatMap := make(map[ItemID]Amount)
	remainders := make(map[ItemID]Amount)
	
	// Calculate totals from tree
	for itemID, node := range treeNodes {
		flatMap[itemID] = node.Amount
	}
	
	// Account for inventory
	for itemID, amount := range c.Inventory.itemStacks {
		if existing, ok := flatMap[itemID]; ok {
			if existing > amount {
				flatMap[itemID] = existing - amount
				remainders[itemID] = 0
			} else {
				flatMap[itemID] = 0
				remainders[itemID] = amount - existing
			}
		} else {
			remainders[itemID] = amount
		}
	}
	
	// Find root nodes (items in desired list)
	var rootNodes []*TreeNode
	for itemID := range c.ItemList.itemStacks {
		if node, exists := treeNodes[itemID]; exists {
			rootNodes = append(rootNodes, node)
		}
	}
	
	return &TreeResult{
		Root:       &TreeNode{Children: rootNodes}, // Virtual root
		FlatMap:    flatMap,
		Remainders: remainders,
	}, nil
}
