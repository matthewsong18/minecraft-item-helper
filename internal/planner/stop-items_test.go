package planner_test

import (
	"minecraft-item-helper/internal/planner"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStopItemsWithOverrides(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		add    []string
		remove []string
		want   map[string]bool
	}{
		{
			name:   "Get default stop items",
			add:    nil,
			remove: nil,
			want: map[string]bool{
				"minecraft:coal":            true,
				"minecraft:copper_ingot":    true,
				"minecraft:diamond":         true,
				"minecraft:emerald":         true,
				"minecraft:gold_ingot":      true,
				"minecraft:iron_ingot":      true,
				"minecraft:lapis_lazuli":    true,
				"minecraft:leather":         true,
				"minecraft:netherite_ingot": true,
				"minecraft:raw_copper":      true,
				"minecraft:raw_gold":        true,
				"minecraft:raw_iron":        true,
				"minecraft:redstone":        true,
				"minecraft:slime_ball":      true,
				"minecraft:wheat":           true,
			},
		},
		{
			name:   "Add and remove stop items",
			add:    []string{"minecraft:iron_block"},
			remove: []string{"minecraft:coal"},
			want: map[string]bool{
				"minecraft:copper_ingot":    true,
				"minecraft:diamond":         true,
				"minecraft:emerald":         true,
				"minecraft:gold_ingot":      true,
				"minecraft:iron_block":      true,
				"minecraft:iron_ingot":      true,
				"minecraft:lapis_lazuli":    true,
				"minecraft:leather":         true,
				"minecraft:netherite_ingot": true,
				"minecraft:raw_copper":      true,
				"minecraft:raw_gold":        true,
				"minecraft:raw_iron":        true,
				"minecraft:redstone":        true,
				"minecraft:slime_ball":      true,
				"minecraft:wheat":           true,
			},
		},
		{
			name:   "Add and remove same item removes the item",
			add:    []string{"minecraft:iron_block"},
			remove: []string{"minecraft:iron_block"},
			want: map[string]bool{
				"minecraft:coal":            true,
				"minecraft:copper_ingot":    true,
				"minecraft:diamond":         true,
				"minecraft:emerald":         true,
				"minecraft:gold_ingot":      true,
				"minecraft:iron_ingot":      true,
				"minecraft:lapis_lazuli":    true,
				"minecraft:leather":         true,
				"minecraft:netherite_ingot": true,
				"minecraft:raw_copper":      true,
				"minecraft:raw_gold":        true,
				"minecraft:raw_iron":        true,
				"minecraft:redstone":        true,
				"minecraft:slime_ball":      true,
				"minecraft:wheat":           true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := planner.StopItemsWithOverrides(tt.add, tt.remove)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("StopItemsWithOverrides() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
