package planner_test

import (
	"minecraft-item-helper/internal/planner"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveBaseMaterials(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		inputs planner.Request
		want   planner.Response
	}{
		{
			name: "test raw materials for a direct recipe",
			inputs: planner.Request{
				RequestedItem:   "planks",
				RequestedAmount: 4,
				Recipes: map[string]planner.Recipe{
					"planks": {
						Yield: 4,
						Inputs: []planner.Ingredient{
							planner.Ingredient{
								Options: []string{"log"},
								Count:   1,
							},
						},
					},
				},
			},
			want: planner.Response{
				FlatMap: map[string]uint{"log": 1},
			},
		},
		{
			name: "test nested two-step recipe",
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]planner.Recipe{
					"planks": {
						Yield: 4,
						Inputs: []planner.Ingredient{
							{
								Options: []string{"log"},
								Count:   1,
							},
						},
					},
					"stick": {
						Yield: 4,
						Inputs: []planner.Ingredient{
							{
								Options: []string{"planks"},
								Count:   2,
							},
						},
					},
				},
			},
			want: planner.Response{
				FlatMap: map[string]uint{"log": 1},
			},
		},
		{
			name: "test recipe with varying levels of base materials",
			inputs: planner.Request{
				RequestedItem:   "iron-pickaxe",
				RequestedAmount: 1,
				Recipes: map[string]planner.Recipe{
					"iron-pickaxe": {
						Yield: 1,
						Inputs: []planner.Ingredient{
							{
								Options: []string{"iron-ingot"},
								Count:   3,
							},
							{
								Options: []string{"stick"},
								Count:   2,
							},
						},
					},
					"planks": {
						Yield: 4,
						Inputs: []planner.Ingredient{
							{
								Options: []string{"log"},
								Count:   1,
							},
						},
					},
					"stick": {
						Yield: 4,
						Inputs: []planner.Ingredient{
							{
								Options: []string{"planks"},
								Count:   2,
							},
						},
					}},
			},
			want: planner.Response{
				FlatMap: map[string]uint{
					"iron-ingot": 3,
					"log":        1,
				},
			},
		},
		{
			name: "test multiplying by yield amount",
			inputs: planner.Request{
				RequestedItem:   "planks",
				RequestedAmount: 8,
				Recipes: map[string]planner.Recipe{
					"planks": {
						Yield: 4,
						Inputs: []planner.Ingredient{
							{
								Options: []string{"log"},
								Count:   1,
							},
						},
					}},
			},
			want: planner.Response{FlatMap: map[string]uint{"log": 2}},
		},
		{
			name: "test multiplier on varying levels of base materials",
			inputs: planner.Request{
				RequestedItem:   "iron-pickaxe",
				RequestedAmount: 3,
				Recipes: map[string]planner.Recipe{
					"iron-pickaxe": {
						Yield: 1,
						Inputs: inputs(
							ingredient("iron-ingot", 3),
							ingredient("stick", 2),
						),
					},
					"stick": {
						Yield: 4,
						Inputs: inputs(
							ingredient("planks", 2),
						),
					},
					"planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("log", 1),
						),
					}},
			},
			want: planner.Response{FlatMap: map[string]uint{
				"iron-ingot": 9,
				"log":        1,
			}},
		},
		{
			name: "test adding extra items to inventory",
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 3,
				Recipes: map[string]planner.Recipe{
					"stick": {
						Yield: 4,
						Inputs: inputs(
							ingredient("planks", 2),
						),
					},
					"planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("log", 1),
						),
					},
				},
				Inventory: map[string]uint{},
			},
			want: planner.Response{
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
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 5,
				Recipes: map[string]planner.Recipe{
					"stick": {
						Yield: 4,
						Inputs: inputs(
							ingredient("planks", 2),
						),
					},
					"planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("log", 1),
						),
					},
				},
				Inventory: map[string]uint{
					"stick": 1,
				},
			},
			want: planner.Response{
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
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]planner.Recipe{
					"stick": {
						Yield: 4,
						Inputs: inputs(
							ingredient("planks", 2),
						),
					},
					"planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("log", 1),
						),
					},
				},
				Inventory: map[string]uint{
					"stick": 4,
				},
			},
			want: planner.Response{
				FlatMap:   map[string]uint{},
				Inventory: map[string]uint{},
			},
		},
		{
			name: "test resolving recipe requiring a tag using first candidate",
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]planner.Recipe{
					"stick": {
						Yield: 4,
						Inputs: inputs(
							ingredient("#planks", 2),
						),
					},
					"oak-planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("oak-log", 1),
						),
					},
				},
				Tags: planner.Tags{
					"#planks": []string{"oak-planks", "birch-planks"},
				},
				Inventory: planner.Inventory{},
			},
			want: planner.Response{
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
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]planner.Recipe{
					"stick": {
						Yield: 4,
						Inputs: inputs(
							ingredient("#planks", 2),
						),
					},
					"oak-planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("oak-log", 1),
						),
					},
					"birch-planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("birch-log", 1),
						),
					},
				},
				Tags: planner.Tags{
					"#planks": []string{"oak-planks", "birch-planks"},
				},
				Inventory: map[string]uint{
					"birch-planks": 2, // Already have birch, should pick birch instead of oak
				},
			},
			want: planner.Response{
				FlatMap:   map[string]uint{}, // No logs needed, birch was already available
				Inventory: map[string]uint{}, // 2 birch consumed completely
			},
		},
		{
			name: "test resolving recipe requiring a tag preferring existing inventory",
			inputs: planner.Request{
				RequestedItem:   "stick",
				RequestedAmount: 4,
				Recipes: map[string]planner.Recipe{
					"stick": {
						Yield: 4,
						// Inputs: map[string]uint{"#planks": 2},
						Inputs: inputs(
							ingredient("#planks", 2),
						),
					},
					"oak-planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("oak-log", 1),
						),
					},
					"birch-planks": {
						Yield: 4,
						Inputs: inputs(
							ingredient("birch-log", 1),
						),
					},
				},
				Tags: planner.Tags{
					"#planks": []string{"oak-planks", "birch-planks"},
				},
				Inventory: map[string]uint{
					"birch-planks": 2, // Already have birch, should pick birch instead of oak
				},
			},
			want: planner.Response{
				FlatMap:   map[string]uint{}, // No logs needed, birch was already available
				Inventory: map[string]uint{}, // 2 birch consumed completely
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := planner.ResolveBaseMaterials(tt.inputs)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ResolveBaseMaterials() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func inputs(ingredient ...planner.Ingredient) []planner.Ingredient {
	return ingredient
}

func ingredient(id string, count uint) planner.Ingredient {
	return planner.Ingredient{
		Options: []string{id},
		Count:   count,
	}
}
