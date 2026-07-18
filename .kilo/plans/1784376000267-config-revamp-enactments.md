# Config Revamp + Config-Driven Enactments

## Goal
Revamp configuration handling for the Blok2 TTRPG Ability Builder in three parts:
1. Split the single-file `ability-builder` config into a per-profile directory of section files.
2. Add a `states` config with cost mapping and wire it into the `Enact State` card.
3. Make enactment cards fully config-driven (field schema + cost mapping). Enactments only for this iteration; interactions, validations, and ability types stay hardcoded.

## Key Decisions (resolved)
- Config split = per-profile directory of section files.
- Leftover areas (`validations`, `dice`, `combat`, `additional_enactment`, `version`, `profile_id`) go into `general.yaml`.
- Only `ability-builder` is converted now. Loader must still accept single-file profiles (dnd, pathfinder2e).
- `states.yaml` = full states with cost mapping; wire Enact State costs now.
- Enactment data model = generic field-values map (not typed struct fields).
- Field cost mapping = per-type cost rules (no expression interpreter).
- Cascading = child fields nested under options.
- Server recomputes enactment costs from the shared schema (authoritative), replacing trust in hidden fields.
- YAML export = schema-driven output keys per field.
- Enactment type list + enactment option lists come from `enactments.yaml`; trait/dice dropdowns reference shared lists via `options_source`.
- No migration of existing saved abilities (accept breakage; saved data is disposable).

## Relevant Files
- `internal/config/config.go` — Config structs, `Load`, `Validate`, `DefaultPath`.
- `internal/config/costs.go` — server-side ability-type cost compute (`ComputeAbilityCosts`).
- `internal/handlers/ability.go` — builder handler, `parseNewEnactments`, `SaveAbilityHandler`, `buildInitialState`, `mustMarshalConfig`.
- `internal/models/enactment.go` — `Enactment` struct, `AllEnactmentTypes`.
- `internal/models/reference.go` — hardcoded option lists.
- `internal/export/yaml.go` — YAML export (`writeYAMLEnactment`).
- `static/js/builder.js` — client card rendering + cost calc.
- `templates/builder.html` — `BUILDER_DATA` bootstrap.
- `config/ability-builder.yaml` — source config to split.
- `main.go` — `resolveConfigPath`, profile/store setup.

---

## Part 1 — Config Split (ability-builder → directory)

### Tasks
1. Create `config/ability-builder/` directory with section files carved out of `config/ability-builder.yaml`:
   - `file_order.yaml` — the `file_order` list.
   - `ability_types.yaml` — `ability_types` map.
   - `enactments.yaml` — `enactments` map (will be extended in Part 3).
   - `interactions.yaml` — `interactions` map.
   - `proficiencies.yaml` — `proficiencies` list.
   - `leveling.yaml` — `leveling` table.
   - `traits.yaml` — `traits` block.
   - `states.yaml` — new (Part 2).
   - `general.yaml` — `version`, `profile_id`, `combat`, `additional_enactment`, `dice`, `validations`.
2. Decide top-level YAML keys inside each section file. Recommended: each file contains the same nested path it had under `ability_builder:` (e.g. `enactments.yaml` starts with `ability_builder:\n  enactments:`), OR flatten so each file holds just its section and the loader assembles. Choose flattened section files (simpler): each file holds only its own top-level section key (e.g. `enactments:` at root of `enactments.yaml`; `version:`/`profile_id:`/`combat:`/... at root of `general.yaml`). The loader maps each file to the right place in the `Config`/`AbilityBuilderConfig`.
3. Update loader in `internal/config/config.go`:
   - Add directory support to `Load(path)`: if `path` is a directory, read known section filenames, unmarshal each into the corresponding struct field, assemble a single `Config`, then run `Validate()`.
   - If `path` is a file, keep current single-file behavior.
   - Keep `DefaultPath()` returning the ability-builder path; update it to point at the directory `config/ability-builder` (verify `main.go`/env/Docker still resolve correctly; `ABILITY_BUILDER_CONFIG` and `-config` flag continue to work for both file and dir).
4. Update `README.md` and `docker-compose.yaml`/`Dockerfile` config mount references from `config/ability-builder.yaml` to the directory (verify volume mounts).
5. Remove the old `config/ability-builder.yaml` after the split (dnd.yaml, pathfinder2e.yaml stay as single files).

### Validation
- `go build ./...`
- `go run main.go` boots; log shows the directory profile loaded.
- Load dnd/pathfinder via `-config config/dnd.yaml` still works (single-file path).

---

## Part 2 — States Config

### Tasks
1. Define `states.yaml` schema and add structs to `internal/config/config.go`:
   - `StatesConfig{ GeneralStates []GeneralStateConfig; SpecificStates []SpecificStateConfig }`.
   - `GeneralStateConfig{ ID, Name, Description string; MinShift, MaxShift int; ShiftCost CostDefinition }` (per-shift add/energy).
   - `SpecificStateConfig{ ID, Name, Description string; AddCost, EnergyCost int }`.
   - Add `States StatesConfig` to `AbilityBuilderConfig`.
2. Populate `states.yaml` from `docs/core/states.md` (general examples like Blinded/Encumbered; specific list from the tables). Assign reasonable placeholder costs; note in the file they are tunable.
3. Wire into `Enact State` (uses Part 3 schema): the Enact State enactment's `fields` in `enactments.yaml` reference the states lists so the card offers a state dropdown (specific) and/or general-state + shift-amount, with costs mapped via per-type cost rules.
4. Expose states to `BUILDER_DATA` (already covered by shipping full `cfg` via `mustMarshalConfig`).

### Validation
- Config validates with states present.
- Enact State card renders state options and computes a non-zero cost when a costed state/shift is selected (both JS live view and server recompute agree).

---

## Part 3 — Config-Driven Enactments

### 3a. Field schema (config)
Extend each entry in `enactments.yaml` with:
```
<key>:
  type: "Enact Damage"          # display type (also the enactment type id)
  description: "..."
  base_cost: { add_cost, energy_cost }
  fields:
    - key: always               # unique within enactment
      label: "Will always resolve"
      type: checkbox
      cost: { add_cost, energy_cost }   # applied when checked
    - key: source
      label: "Source"
      type: cascade             # dropdown that reveals child fields per option
      options:
        - { value: d4,  label: "1d4",  cost: {add_cost:0, energy_cost:0} }
        - { value: d6,  label: "1d6",  cost: {add_cost:2, energy_cost:1} }
        - value: trait
          label: "Trait (1d10)"
          cost: {add_cost:3, energy_cost:0}
          fields:               # nested, shown when this option active
            - { key: source_trait, type: dropdown, options_source: traits_all, label: "Trait" }
        - value: other
          label: "Another roll result"
          cost: {add_cost:3, energy_cost:1}
          fields:
            - { key: other_roll_text, type: free_text, label: "Other Roll Text" }
    - key: flat
      label: "Flat Bonus"
      type: free_number
      default: 0
      min: 0
      max: 20
      step: 1
      per_step: { add_cost, energy_cost }   # increase/decrease may be split; see below
    - key: solutions
      label: "Solutions"
      type: solutions
      options_source: traits_all
      default_count: 2
      per_item:
        increase: { add_cost, energy_cost }
        decrease: { add_cost, energy_cost }
      export: { key: solutions }
```

Field types and cost rules:
- `dropdown`: options list; selected option's `cost` applies. Options inline or `options_source` (dynamic list, no per-option cost unless overridden).
- `free_text`: no cost. Optional `default`.
- `free_number`: `default`, `min`, `max`, `step`; cost = `((value-default)/step) * per_step` clamped at 0 when at/below default; allow `per_step.increase`/`per_step.decrease` split (mirrors phase reverse-duration / step costs).
- `checkbox`: `cost` when checked.
- `solutions`: dynamic add/remove list from `options_source`; cost = deviation from `default_count` × `per_item.increase|decrease`.
- `cascade`: `dropdown` or `checkbox` variant whose options/states embed nested `fields` that render (and cost) only when active.

`options_source` supported values: `traits_general`, `traits_offense`, `traits_defense`, `traits_all`, `dice_damage`, `dice_generic`, `states_general`, `states_specific`. Resolved against `traits.yaml`/`general.yaml.dice`/`states.yaml`.

Each field may declare `export: { key, suffix, omit_when_default }` for YAML output.

### 3b. Go config structs
- Add `Fields []FieldConfig` to `EnactmentConfig` (keep existing cost fields for now for compat, or migrate them into `fields`; prefer moving all into `fields`).
- `FieldConfig{ Key, Label, Type string; Cost CostDefinition; Options []FieldOption; OptionsSource string; Default any; Min, Max, Step int; PerStep StepCosts; DefaultCount int; PerItem StepCosts; Export FieldExport }`.
- `FieldOption{ Value, Label string; Cost CostDefinition; Fields []FieldConfig }`.
- `FieldExport{ Key, Suffix string; OmitWhenDefault bool }`.
- Drive `AllEnactmentTypes` and per-type option lists from config; remove enactment-specific entries from `reference.go` (`DamageDiceOptions`, `DirectionOptions`, `ShiftDirectionOptions`, `TriggerTimings`, `PersistentEffectTypes` as they pertain to enactments) and from JS `BUILDER_DATA`. Keep trait lists (used by interactions/validation too).

### 3c. Data model
- Replace enactment-specific typed fields in `models.Enactment` with `Fields map[string]any` (string / bool / number / []string values). Keep `Name`, `Description`, `Type`, `BuildCost`, `CastCost`, `Formula`, `Always` (or fold Always into Fields), and `Interaction *Interaction` (interactions stay typed).
- Update `Enactment.TotalCost()` accordingly.

### 3d. Server parse + cost recompute
- Rewrite `parseNewEnactments` in `internal/handlers/ability.go` to read generic `enact_<idx>_<fieldkey>` values (including nested cascade child keys and repeated `solution`/list keys) into `Fields`.
- Add a generic Go cost evaluator (new file, e.g. `internal/config/enact_costs.go`) that walks the enactment `fields` schema + submitted values to compute enactment `build`/`cast`. Include `always`/base_cost handling. This becomes authoritative; stop trusting `enact_<idx>_build/cast` hidden fields for enactments.
- Interaction/validation costs: keep current behavior (still trust JS hidden fields for those) since they're out of scope this iteration.
- Enforce point budget using server-computed enactment costs (existing budget check in `SaveAbilityHandler`).

### 3e. Client rendering + live cost
- Replace `renderEnact*` functions and the enactment branch of `calcEnact` in `static/js/builder.js` with:
  - A generic `renderEnactCard(type, data)` that reads the enactment's `fields` schema from `C.enactments[key].fields` and renders each field type (dropdown, free_text, free_number, checkbox, solutions, cascade with nested children).
  - A generic `calcEnact(card)` that walks the schema + DOM values applying per-type cost rules identical to the Go evaluator.
- Keep interaction/validation rendering + calc as-is.
- Keep hydration (`buildInitialState` / edit mode) working with the generic `Fields` map.

### 3f. Export
- Rewrite `writeYAMLEnactment` to walk the enactment schema + `Fields`, emitting `export.key` (with `suffix`, `omit_when_default`). Reproduce current damage/healing/movement/etc. output by setting appropriate `export` keys in `enactments.yaml`.
- Interaction/validation export unchanged.

### Validation
- `go build ./...` and `go test ./...` (update/adjust `internal/export/character_test.go` if it asserts old enactment YAML).
- Manual: build each enactment type in the UI; verify live cost equals server-computed cost (compare review page / saved ability) and YAML export matches expectations.
- Verify a new/edited config field (e.g. add a checkbox to `enactments.yaml`) appears in the card, affects cost, and exports — without code changes.

---

## Risks / Notes
- Two cost code paths (JS + Go) must stay in lockstep; the per-type cost rules are intentionally simple to keep parity. Add a small shared spec/table in the plan comments and mirror exactly.
- Cascade child field name collisions: namespace child keys within their field/option (e.g. `source__source_trait`) or rely on unique keys; ensure parse + hydrate agree.
- `options_source` requires trait/dice/states lists to be present; validate references at load time.
- Removing hardcoded lists may affect interaction/validation code that shares them (e.g. trait lists) — only remove enactment-specific lists.
- Existing saved abilities will not migrate and may render blank enactment fields; acceptable per decision.
- Confirm Docker/compose config mount path updates so the container finds the directory profile.

## Suggested Implementation Order
1. Part 1 (split + loader dir support) — smallest, unblocks everything.
2. Part 3b/3c (Go structs + model) then 3d (parse + evaluator).
3. Part 3e (JS generic renderer/calc) to parity.
4. Part 3f (export) + test updates.
5. Part 2 (states) using the new schema.
