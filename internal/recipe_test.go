package internal_test

import (
	"minecraft-item-helper/internal"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveBaseMaterials(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		inputs internal.Request
		want   internal.Response
	}{
		{
			name: "test raw materials for a direct recipe",
			inputs: internal.Request{
				RequestedItem:   "planks",
				RequestedAmount: 4,
				Recipes: map[string]internal.Recipe{
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
				},
			},
			want: internal.Response{
				FlatMap: map[string]uint{"log": 1},
			},
		},
		{
			name: "test nested two-step recipe",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]internal.Recipe{
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
				},
			},
			want: internal.Response{
				FlatMap: map[string]uint{"log": 1},
			},
		},
		{
			name: "test recipe with varying levels of base materials",
			inputs: internal.Request{
				RequestedItem:   "iron-pickaxe",
				RequestedAmount: 1,
				Recipes: map[string]internal.Recipe{
					"iron-pickaxe": {
						Yield: 1,
						Inputs: map[string]uint{
							"iron-ingot": 3,
							"stick":      2,
						},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					}},
			},
			want: internal.Response{
				FlatMap: map[string]uint{
					"iron-ingot": 3,
					"log":        1,
				},
			},
		},
		{
			name: "test multiplying by yield amount",
			inputs: internal.Request{
				RequestedItem:   "planks",
				RequestedAmount: 8,
				Recipes: map[string]internal.Recipe{
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					}},
			},
			want: internal.Response{FlatMap: map[string]uint{"log": 2}},
		},
		{
			name: "test multiplier on varying levels of base materials",
			inputs: internal.Request{
				RequestedItem:   "iron-pickaxe",
				RequestedAmount: 3,
				Recipes: map[string]internal.Recipe{
					"iron-pickaxe": {
						Yield: 1,
						Inputs: map[string]uint{
							"iron-ingot": 3,
							"stick":      2,
						},
					},
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					}},
			},
			want: internal.Response{FlatMap: map[string]uint{
				"iron-ingot": 9,
				"log":        1,
			}},
		},
		{
			name: "test adding extra items to inventory",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 3,
				Recipes: map[string]internal.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
				},
				Inventory: map[string]uint{},
			},
			want: internal.Response{
				FlatMap: map[string]uint{
					"log": 1,
				},
				Inventory: map[string]uint{
					"planks": 2,
					"stick":  1,
				},
			},
		},
		{
			name: "test having items in inventory",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 5,
				Recipes: map[string]internal.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
				},
				Inventory: map[string]uint{
					"stick": 1,
				},
			},
			want: internal.Response{
				FlatMap: map[string]uint{
					"log": 1,
				},
				Inventory: map[string]uint{
					"planks": 2,
				},
			},
		},
		{
			name: "test having exact items in inventory",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]internal.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"planks": 2},
					},
					"planks": {
						Yield:  4,
						Inputs: map[string]uint{"log": 1},
					},
				},
				Inventory: map[string]uint{
					"stick": 4,
				},
			},
			want: internal.Response{
				FlatMap:   map[string]uint{},
				Inventory: map[string]uint{},
			},
		},
		{
			name: "test resolving recipe requiring a tag using first candidate",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]internal.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"#planks": 2}, // Input uses tag
					},
					"oak-planks": {
						Yield:  4,
						Inputs: map[string]uint{"oak-log": 1},
					},
				},
				Tags: internal.Tags{
					"#planks": []string{"oak-planks", "birch-planks"},
				},
				Inventory: internal.Inventory{},
			},
			want: internal.Response{
				FlatMap: map[string]uint{
					"oak-log": 1,
				},
				Inventory: map[string]uint{
					"oak-planks": 2, // 4 produced, 2 consumed for sticks
				},
			},
		},
		{
			name: "test resolving recipe requiring a tag preferring existing inventory",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]internal.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"#planks": 2},
					},
					"oak-planks": {
						Yield:  4,
						Inputs: map[string]uint{"oak-log": 1},
					},
					"birch-planks": {
						Yield:  4,
						Inputs: map[string]uint{"birch-log": 1},
					},
				},
				Tags: internal.Tags{
					"#planks": []string{"oak-planks", "birch-planks"},
				},
				Inventory: map[string]uint{
					"birch-planks": 2, // Already have birch, should pick birch instead of oak
				},
			},
			want: internal.Response{
				FlatMap:   map[string]uint{}, // No logs needed, birch was already available
				Inventory: map[string]uint{}, // 2 birch consumed completely
			},
		},
		{
			name: "test resolving recipe requiring a tag preferring existing inventory",
			inputs: internal.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]internal.Recipe{
					"stick": {
						Yield:  4,
						Inputs: map[string]uint{"#planks": 2},
					},
					"oak-planks": {
						Yield:  4,
						Inputs: map[string]uint{"oak-log": 1},
					},
					"birch-planks": {
						Yield:  4,
						Inputs: map[string]uint{"birch-log": 1},
					},
				},
				Tags: internal.Tags{
					"#planks": []string{"oak-planks", "birch-planks"},
				},
				Inventory: map[string]uint{
					"birch-planks": 2, // Already have birch, should pick birch instead of oak
				},
			},
			want: internal.Response{
				FlatMap:   map[string]uint{}, // No logs needed, birch was already available
				Inventory: map[string]uint{}, // 2 birch consumed completely
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.ResolveBaseMaterials(tt.inputs)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ResolveBaseMaterials() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
