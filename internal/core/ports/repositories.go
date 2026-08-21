package ports

import "minecraft-item-helper/internal/core/domain"

type RecipeRepository interface {
	GetRecipe(itemName string) (*domain.Recipe, error)
	GetAllRecipes() (map[string]domain.Recipe, error)
	GetTagGroup(tag string) ([]string, error)
}
