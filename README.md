# Minecraft Item Helper

A CLI MVP for turning a Minecraft build list into a shopping list. It expands recipes, subtracts owned materials, and shares crafting leftovers across the build.

## Quick start

Requires Go 1.27 or newer. Run commands from the repository root:

```sh
go run ./cmd/cli --item minecraft:lectern
```

This reports three leather, three oak logs, and nine sugar cane. Leather is a default **stop item**, so the planner does not expand it into rabbit hide.

For a complete build with existing inventory and saved preferences:

```sh
go run ./cmd/cli --requests examples/build.json \
  --inventory examples/inventory.json --config examples/config.json \
  --show-inventory
```

The example requests a crafting table and oak planks. Oak in the build takes priority over the stocked cherry planks and saved cherry preference:

```text
minecraft:oak_log: 1

Projected inventory after crafting (requested items set aside):
minecraft:cherry_planks: 64
minecraft:oak_planks: 3
```

## Build lists and inventory

Both files are JSON objects mapping concrete item IDs to quantities:

```json
{
  "minecraft:oak_planks": 128,
  "minecraft:torch": 32
}
```

Build quantities must be positive integers; inventory quantities may also be zero. Use namespaced IDs such as `minecraft:oak_planks`. Build files use one entry per item ID; schematic files and copied guide text are not imported directly.

The planner first sets aside owned items explicitly requested as finished outputs. It processes remaining requests in item-ID order, sharing available stock and crafting leftovers. Output lists additional materials to gather, sorted by ID; an empty shopping list means no additional materials are needed. `--show-inventory` displays projected unused stock and crafting surplus, excluding the requested finished outputs. Input files are never rewritten.

## Material choices and settings

Pass an optional JSON settings file with `--config`. See [examples/config.json](examples/config.json) for a complete example. Its fields are:

- `preferences`: tag IDs mapped to ordered lists of preferred concrete items.
- `stop_items.add` and `stop_items.remove`: changes to the default stop list. Removal wins if an item appears in both lists.
- `recipe_choices`: output item IDs mapped to recipe filenames without `.json`.

For flexible ingredients, the planner chooses a compatible item present in the build list first, then the item with the most available inventory, then a saved preference, then the first listed option. Within the build's compatible candidates, inventory quantity and preference order break ties. Exact requests keep their requested material.

Preferences apply separately to each tag: preferring cherry logs does not imply cherry planks or slabs. A tag preference also applies to an explicit OR list when the tag contains every alternative. Unavailable preferred items are ignored; equally applicable OR-tag preferences are considered in tag-ID order. Nested tags are expanded before choosing.

Stop items consume matching inventory and report any shortage without expanding their recipe. Defaults are leather; iron, gold, copper, and netherite ingots; coal, diamond, emerald, redstone, lapis lazuli; raw iron, gold, and copper; wheat; and slime balls. These defaults avoid expanding common gathering targets through reversible storage-block recipes.

Removing a stop item allows expansion again. The example config removes iron ingots and selects smelting raw iron:

```sh
go run ./cmd/cli --item minecraft:iron_ingot --amount 9 \
  --config examples/config.json
```

This reports nine raw iron. Stopping at ingots does not convert owned blocks or nuggets into ingots.

## Flags

| Flag | Purpose | Default |
| --- | --- | --- |
| `--item` | One concrete item to request | Required unless using `--requests` |
| `--amount` | Quantity for `--item` | `1` |
| `--requests` | Build-list JSON file; cannot combine with `--item` or `--amount` | None |
| `--inventory` | Owned quantities JSON file | Empty inventory |
| `--config` | Preferences, stop-item overrides, and recipe choices | Built-in defaults |
| `--show-inventory` | Display projected leftovers | `false` |
| `--recipes` | Recipe JSON directory | `assets/recipes` |
| `--tags` | Item tag JSON directory, read recursively | `assets/item-tag-groups` |

Tag IDs come from relative filenames: `coals.json` provides `#minecraft:coals`.

## MVP limits

Shaped crafting, shapeless crafting, and smelting are supported; other recipe types are skipped. Items without a supported recipe become shopping materials. Smelting fuel is not included.

When an output has multiple recipes, the default is its matching filename (`oak_planks.json` for oak planks), otherwise the first supported filename alphabetically. `recipe_choices` overrides that route. The planner reports recipe or tag cycles with dependency paths; it does not automatically try a different route.

Planning is deterministic and greedy. Each ingredient group chooses one alternative, and results are not guaranteed to minimize total gathering across every possible recipe or combination of inventory items.

## Development

Start with `internal/planner/` for material-planning rules and `internal/datafiles/` for reading recipe, tag, settings, and inventory files. `cmd/cli/` handles command-line flags and printed output. Tests sit beside the code they cover; CLI behavior is tested through `run(args, output)` without launching subprocesses.

```sh
just test
just test-watch
```

These shortcuts require `just` and `gotestsum`. Alternatively, use `go test ./...`.
