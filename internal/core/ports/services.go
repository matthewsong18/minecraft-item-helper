// internal/core/ports/services.go
package ports

import "minecraft-item-helper/internal/core/domain"

type CraftingService interface {
	GetRecipe(itemName string) (*domain.Recipe, error)
	CalculateMaterials(itemName string, count int) (map[string]int, error)
}
