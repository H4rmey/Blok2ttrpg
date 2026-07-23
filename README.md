# Blok2ttrpg

A flexible, theme-agnostic tabletop RPG system, together with a config-driven
companion web app for building characters and abilities.

Blok2ttrpg is a long-running personal project to build a TTRPG that is flexible
enough to be applied to almost any setting. The system deliberately ships with
**no theme of its own** - it relies on the players' own worldbuilding. In
principle any genre should work without re-theming anything, because nothing has
a theme baked in yet.

> **Heads up:** this is still *highly* work in progress. There will be bugs,
> gaps, and mechanics that are not written down yet. It has also not been
> playtested.

## What's in this repo

The project has two halves:

1. **The system (the rules).** Hand-written documentation living in `docs/`.
   These describe the core mechanics and optional modules.
2. **The app (the tooling).** A Go web application that turns a directory of
   YAML config into an interactive character and ability builder, plus generated
   documentation.

The rules are written by hand. The surrounding application is optional but very
useful when creating abilities.

## The system

The rules are sometimes very specific and sometimes intentionally vague; the
docs generally explain why. A lot of common mechanics (attacking, initiative,
movement, and so on) are borrowed from other TTRPGs, so if you already know
other systems, much of this will feel familiar.

### Core documentation (`docs/core/`)

| Document | Topic |
| --- | --- |
| `abilities.md` | How abilities work |
| `character-attributes.md` | Character attributes |
| `character-traits.md` | Character traits |
| `combat.md` | Combat rules |
| `dice-rolling.md` | Dice rolling |
| `items.md` | Items |
| `leveling.md` | Leveling |
| `states.md` | States |

### Modules

Modules are optional extensions to the core system. Current and planned modules
live in `docs/modules/`:

- **Character Creation** (core)
- **Ability Builder**
- **Leveling**
- **Skill Trees**
- **Magic System**
- **World**

Planned modules include character presets (races/classes), predefined ability
templates, an items list, and predefined abilities with skill trees.

## The app

A config-driven character and ability builder written in Go with
`html/template` + [HTMX](https://htmx.org). No Node, no npm, no build step.

**The config leads.** Everything the app renders and costs is derived from a
directory of YAML files. There are no hardcoded ability types, enactments,
traits, or character attributes anywhere in the Go code.

### Features

- Config files drive every rendered page.
- Live, advisory cost calculation while building abilities.
- Export documentation (Markdown + browser-print PDF).
- Export character sheets (YAML + browser-print PDF).
- Export abilities as YAML.
- YAML import/export of characters.
- JSON persistence of all characters.
- Dark/light mode and breadcrumb navigation.
- PDFs produced by the browser's own `window.print()` - no headless browser
  dependency.
- Cost is advisory, never blocking: you can always save an ability even if it is
  over budget; the UI just flags it.

### Running

```sh
go run .
```

Then open http://localhost:8080.

Or with Docker Compose:

```sh
docker compose up
```

### Flags / environment

- `-config` (or `CONFIG`): config directory or file. Default `config/ability-builder`.
- `-templates`: template directory. Default `templates`.
- `PORT`: listen port. Default `8080`.

Characters are stored in `data/<profile_id>/characters.json`.

## Configuration

A ruleset is a directory of YAML files, merged in filename order. See
`config/ability-builder/` for the reference ruleset:

| File | Purpose |
| --- | --- |
| `general.yaml` | version, profile id, title |
| `attributes.yaml` | character attribute groups/fields |
| `traits.yaml` | trait groups and reusable option lists |
| `ability_types.yaml` | ability types |
| `enactments.yaml` | enactment building blocks |
| `interactions.yaml` | interaction building blocks |
| `proficiencies.yaml` | proficiencies |
| `leveling.yaml` | leveling rules |
| `states.yaml` | states |
| `file_order.yaml` | merge/load order for the ruleset |

To make your own ruleset, copy `config/ability-builder` to a new directory, edit
the YAML, and run with `-config path/to/your-ruleset`.

### Field types

A `Field` uses a `type` discriminator: `text`, `textarea`, `number`,
`checkbox`, `dropdown`, `list`. Costs can attach at the field level (`cost`),
per number step (`per_step`), per dropdown option (`options[].cost`), or per
list row (`per_item`). Fields can be conditionally shown with `show_when`.

## Documentation in the app

Visit `/docs` for HTML docs generated from the config (with a print/PDF button),
or `/docs/markdown` to download the Markdown. Because the docs iterate over the
same `Config` the app runs on, they never drift out of sync.

## Project layout

```
main.go                 entrypoint / flags
cmd/gendocs             standalone docs generator
internal/config         schema, loading/merging, validation, lookups
internal/model          generic Character / Ability model
internal/store          JSON-backed character store
internal/engine         advisory cost calculation
internal/docs           config -> markdown/HTML docs
internal/export         YAML import/export
internal/web            HTTP handlers, routing, template funcs
templates               html/template views
static                  css, app.js, vendored htmx
config/ability-builder  reference ruleset
docs/                   hand-written system documentation
```

## A note on authorship

The system documentation is written by hand. The application code is largely
AI-assisted - I would rather spend my limited free time on my
friends and family than hand-writing all the tooling - but it is heavily steered
and managed by me.
