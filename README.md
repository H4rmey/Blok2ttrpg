# Blok2ttrpg v5

A config-driven tabletop RPG character and ability builder. Written in Go with
`html/template` + [HTMX](https://htmx.org). No Node, no npm, no build step.

## Why v5

This is a from-scratch rebuild focused on one idea: **the config leads.**
Everything the app renders and costs is derived from a directory of YAML files.
There are no hardcoded ability types, enactments, traits, or character
attributes anywhere in the Go code.

### Kept from earlier versions
- Config files render the pages.
- Export documentation (Markdown + browser-print PDF).
- Export character sheets (YAML + browser-print PDF).
- Export abilities as YAML.
- YAML import/export of characters.
- Go + templates + HTMX, dark/light mode, breadcrumb navigation.
- Live cost calculation while building.
- JSON persistence of all characters.
- Adding/removing traits in config updates the builder everywhere.
- Documentation generated from the same config the app uses, so docs and
  behaviour stay in sync.

### Fixed in v5
- **No Node dependency.** PDFs are produced by the browser's own
  `window.print()` (print-friendly pages), so `puppeteer-core` and
  `node_modules` are gone entirely.
- **Config-driven character attributes.** Add or change attributes in YAML;
  the character model stores everything generically.
- **Less verbose code.** A single generic `Component` type backs ability
  types, enactments, and interactions, and a single generic `Field` type
  drives both the UI and the cost engine.
- **Zero hardcoded type names.** All ability types, enactments, interactions,
  proficiencies, and traits come from config.
- **Cost is advisory, never blocking.** You can always save an ability even if
  it is over budget; the UI just flags it.

## Running

```sh
go run .
```

Then open http://localhost:8080.

### Flags / environment
- `-config` (or `CONFIG`): config directory or file. Default `config/default`.
- `-templates`: template directory. Default `templates`.
- `PORT`: listen port. Default `8080`.

Characters are stored in `data/<profile_id>/characters.json`.

## Configuration

A ruleset is a directory of YAML files, merged in filename order. See
`config/default/` for the reference ruleset:

| File | Purpose |
| --- | --- |
| `01-general.yaml` | version, profile id, title, leveling, proficiencies |
| `02-attributes.yaml` | character attribute groups/fields |
| `03-traits.yaml` | trait groups and reusable option lists |
| `04-ability-types.yaml` | ability types |
| `05-enactments.yaml` | enactment building blocks |
| `06-interactions.yaml` | interaction building blocks + docs order |

To make your own ruleset, copy `config/default` to a new directory, edit the
YAML, and run with `-config path/to/your-ruleset`.

### Field types
A `Field` uses a `type` discriminator: `text`, `textarea`, `number`,
`checkbox`, `dropdown`, `list`. Costs can attach at the field level (`cost`),
per number step (`per_step`), per dropdown option (`options[].cost`), or per
list row (`per_item`). Fields can be conditionally shown with `show_when`.

## Documentation

Visit `/docs` for HTML docs generated from the config (with a print/PDF
button), or `/docs/markdown` to download the Markdown. The docs templates in
`config/<ruleset>/docs/` iterate over the same `Config` the app runs on, so
they never drift out of sync.

## Project layout

```
main.go                 entrypoint / flags
internal/config         schema, loading/merging, validation, lookups
internal/model          generic Character / Ability model
internal/store          JSON-backed character store
internal/engine         advisory cost calculation
internal/docs           config -> markdown/HTML docs
internal/export         YAML import/export
internal/web            HTTP handlers, routing, template funcs
templates               html/template views
static                  css, app.js, vendored htmx
config/default          reference ruleset
```
