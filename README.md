# Minecraft Item Helper

Minecraft Item Helper is an early CLI MVP that calculates the base materials needed to craft one requested item and amount. It reads Minecraft recipe and item tag JSON files, resolves supported recipes recursively, and prints the remaining materials and quantities.

## Requirements

- Go 1.27 or newer
- Optional: [just](https://just.systems/) and `gotestsum` to use the test shortcuts

## Run

From the repository root, pass a namespaced item ID. Amount defaults to `1`:

```sh
go run ./cmd/cli --item minecraft:torch --amount 16
```

The command prints one `item ID: quantity` line for each base material, sorted by item ID. You can select alternate data directories with `--recipes` and `--tags`:

```sh
go run ./cmd/cli --item minecraft:oak_planks --amount 8 \
  --recipes ./data/recipes --tags ./data/item-tag-groups
```

By default, recipes load from `assets/recipes/` and tags load recursively from `assets/item-tag-groups/`. Tag IDs are inferred from their relative JSON file paths; for example, `coals.json` provides `#minecraft:coals`.

| Flag | Meaning | Default |
| --- | --- | --- |
| `--item` | Required namespaced item ID to resolve | — |
| `--amount` | Requested quantity; must be greater than zero | `1` |
| `--recipes` | Recipe JSON directory | `assets/recipes` |
| `--tags` | Item tag JSON directory | `assets/item-tag-groups` |

## Current Scope

The parser supports shaped crafting, shapeless crafting, and smelting recipes. Unsupported recipe types are skipped. Ingredients that have multiple alternatives currently resolve to the first listed option, and tag ingredients use the first value in the tag. The CLI does not yet accept inventory or multiple requests.

Recipe and tag data are stored as JSON under `assets/`. Domain models, parsing, and material resolution are in `internal/`; CLI argument handling and data loading are in `cmd/cli/`.

## Development

Run the full test suite with Go:

```sh
go test ./...
```

Or use the `just` shortcut, which runs `gotestsum`:

```sh
just test
```
