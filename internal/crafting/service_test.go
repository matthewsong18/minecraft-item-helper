package crafting_test

import (
	"context"
	"minecraft-item-helper/internal/crafting"
	"minecraft-item-helper/internal/domain"
	"testing"
)

type fakeRecipeRepo struct {
	recipes map[domain.ItemID][]domain.Recipe
}

func newFakeRecipeRepo() *fakeRecipeRepo {
	return &fakeRecipeRepo{
		recipes: make(map[domain.ItemID][]domain.Recipe),
	}
}

func (f *fakeRecipeRepo) Add(r domain.Recipe) {
	f.recipes[r.Output.Item] = append(f.recipes[r.Output.Item], r)
}

func (f *fakeRecipeRepo) FindRecipesFor(ctx context.Context, item domain.ItemID) ([]domain.Recipe, error) {
	return f.recipes[item], nil
}

func item(id domain.ItemID, count domain.Amount) domain.ItemStack {
	return domain.ItemStack{Item: id, Amount: count}
}

func ing(id domain.ItemID, count domain.Amount) domain.ItemIngredient {
	return domain.ItemIngredient{ID: id, Count: count}
}

func (f *fakeRecipeRepo) addRecipe(output domain.ItemStack, inputs ...domain.RecipeInput) {
	f.recipes[output.Item] = append(f.recipes[output.Item], domain.Recipe{
		Inputs: inputs,
		Output: output,
	})
}

var _ domain.RecipeRepository = (*fakeRecipeRepo)(nil)

type fakeTagRepo struct {
	tags map[domain.TagID][]domain.ItemID
}

func newFakeTagRepo() *fakeTagRepo {
	return &fakeTagRepo{
		tags: map[domain.TagID][]domain.ItemID{},
	}
}

func (f *fakeTagRepo) AddTag(tag domain.TagID, items ...domain.ItemID) {
	f.tags[tag] = append(f.tags[tag], items...)
}

// GetItemsIn implements [domain.TagRepository].
func (f *fakeTagRepo) GetItemsIn(ctx context.Context, tag domain.TagID) ([]domain.ItemID, error) {
	panic("unimplemented")
}

var _ domain.TagRepository = (*fakeTagRepo)(nil)

func TestCalculateItemTree_SingleStepCraft(t *testing.T) {
	recipeRepo := newFakeRecipeRepo()
	recipeRepo.addRecipe(
		item("minecraft:oak_planks", 4),
		ing("minecraft:oak_log", 1),
	)
	tagRepo := newFakeTagRepo()

	svc := crafting.NewCraftService(recipeRepo, tagRepo)

	req := crafting.CalculateItemTreeRequest{
		DesiredItems: map[domain.ItemID]domain.Amount{
			"minecraft:oak_planks": 4,
		},
	}

	res, err := svc.CalculateItemTree(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := res.FlatMap["minecraft:oak_log"]; got != 1 {
		t.Errorf("expected 1 oak log, got %d", got)
	}
	if got := res.Remainders["minecraft:oak_planks"]; got != 0 {
		t.Errorf("expected 0 leftover planks, got %d", got)
	}
}
