# Ability Builder Configuration

## Overview

The Ability Builder loads its default rules from the split YAML directory `config/ability-builder/`. Set `ABILITY_BUILDER_CONFIG` to point at another config directory or a legacy single YAML file.

The split directory is loaded into `AbilityBuilderConfig` from these section files:

- `general.yaml`
- `file_order.yaml`
- `ability_types.yaml`
- `enactments.yaml`
- `interactions.yaml`
- `proficiencies.yaml`
- `traits.yaml`
- `leveling.yaml`
- `states.yaml`

Split config loading is strict: unknown YAML keys are rejected. Keep examples and edits aligned with the schema names exactly.

Legacy single-file configs, such as `config/dnd.yaml` and `config/pathfinder2e.yaml`, are still supported. When a config section has no generic `fields` schema, the system falls back to older hardcoded cost paths for that section.

## File-by-file reference

### `general.yaml`

Controls root metadata and global defaults:

- `version`
- `profile_id`
- `combat.actions.amount`
- `additional_enactment`
- `dice.damage`
- `dice.generic`
- `validations`
- generic validation `fields`

Example:

```yaml
version: 1
profile_id: ability-builder
combat:
  actions:
    amount: 3
additional_enactment:
  add_cost: 1
  energy_cost: 1
  description: "Adding an additional Enactment beyond the first"
```

### `file_order.yaml`

Controls the order in which Markdown files under `./docs/` are appended to generated output.

```yaml
file_order:
  - ./docs/modules/ability-builder/introduction.md
  - ./docs/modules/ability-builder/guide.md
```

The list must be exhaustive: every `.md` file under `./docs/` must appear exactly once.

### `ability_types.yaml`

Defines ability type display names, descriptions, base energy/action values, legacy cost settings, compatible enactments, and generic `fields` for Execution, Reaction, Phase, Minion, Preparation, and Concentration.

Example:

```yaml
ability_types:
  execution:
    name: "Execution"
    description: "Performed instantly during a character's turn."
    base_energy: 3
    base_action: 2
    compatible_enactments:
      - Enact Damage
      - Enact Healing
    fields:
      - key: item_dep
        label: "Has Item Dependency"
        type: checkbox
        cost:
          add_cost: -1
          energy_cost: 0
```

### `enactments.yaml`

Defines enactment type names, descriptions, base costs, legacy cost settings, and generic `fields` for Enact Damage, Healing, Movement, Proficiency Shift, Persistent Effect, State, and Negation.

Example:

```yaml
enactments:
  damage:
    type: "Enact Damage"
    description: "Inflicts damage to a target."
    base_cost:
      add_cost: 2
      energy_cost: 1
```

### `interactions.yaml`

Defines interaction type names, descriptions, base costs, legacy cost settings, and generic `fields` for Self, Direct, Ranged, Area, and Area of Effect.

Example:

```yaml
interactions:
  direct:
    type: "Direct"
    description: "Affects a single target within 1m."
    default_range: 1
    default_targets: 1
    base_cost:
      add_cost: 0
      energy_cost: 0
```

### `traits.yaml`

Provides trait lists used by option sources.

```yaml
traits:
  general:
    - Strength
    - Dexterity
  offense:
    - Precision
  defense:
    - Reflex
  vital:
    - HP
```

### `proficiencies.yaml`

Defines proficiency tiers, point cost, dice per category, and vital values.

```yaml
proficiencies:
  - id: trained
    name: "Trained"
    cost: 1
    dice:
      general: "d8"
      offense: "d8"
      defense: "d8"
    vitals:
      hp: 16
      movement: 5
      energy: 12
```

### `leveling.yaml`

Defines trait point and ability point progression.

```yaml
leveling:
  max_level: 10
  trait_points:
    standard_trait_count: 22
    starting_formula: "(trait_count + 2) / 3"
    levels:
      - level: 1
        points_gained: 0
        total: 8
```

### `states.yaml`

Defines the Enact State data set:

- `additional_state`: surcharge for each selected state after the first.
- `general_states`: flexible shift states with min/max bounds and per-shift cost.
- `specific_states`: fixed-cost named states.

```yaml
additional_state:
  add_cost: 1
  energy_cost: 0

general_states:
  - id: encouraged
    name: "Encouraged"
    description: "Positive trait shifts"
    min_shift: 1
    max_shift: 6
    shift_cost:
      add_cost: 2
      energy_cost: 1
```

## Generic field schema

Generic fields appear under `fields` or `row_fields`. Supported `FieldConfig` keys are:

| Key | Purpose |
| --- | --- |
| `key` | Stable submitted field key. This is persisted in generic field maps. |
| `label` | UI label. |
| `type` | Field type: `checkbox`, `dropdown`, `free_text`, `free_number`, `solutions`, or `states`. |
| `cost` | Field-level `add_cost` / `energy_cost`. |
| `options` | Inline dropdown options. |
| `options_source` | Dynamic option source name resolved by the browser. |
| `default` | Default submitted/display value. |
| `min` | Minimum numeric value. |
| `max` | Maximum numeric value. |
| `step` | Numeric increment size. Defaults to `1` for cost calculation when omitted. |
| `rounding` | Optional step rounding: `ceil` or `floor`. |
| `per_step` | `increase` / `decrease` costs for `free_number` deltas. |
| `default_count` | Default row count for repeatable fields. |
| `per_item` | `increase` / `decrease` costs for repeatable row count deltas. |
| `export` | Export mapping with `key`, optional `suffix`, and `omit_when_default`. |
| `row_fields` | Sub-fields used by `solutions` and `states` rows. |
| `stores_to` | Maps generic values to typed model fields for export compatibility. |
| `visibility_when` | Controlling sibling field key. |
| `show_when` | Required controlling value for visibility. |

`FieldOption` supports `value`, `label`, optional `cost`, and optional child `fields`. `CostDefinition` supports `add_cost`, `energy_cost`, optional `description`, and optional `step`.

## Supported field types

### `checkbox`

Charges `cost` only when checked.

```yaml
- key: always
  label: "Will always resolve"
  type: checkbox
  cost:
    add_cost: 5
    energy_cost: 3
```

### `dropdown`

A dropdown can use inline `options` or an `options_source`. Do not mix both on the same field.

A field-level `cost` is charged when any non-empty option is selected:

```yaml
- key: offense
  label: "Offensive Trait (extra die)"
  type: dropdown
  options_source: traits_offense
  cost:
    add_cost: 4
    energy_cost: 2
```

Inline options can carry per-option cost:

```yaml
- key: shift_dir
  label: "Direction"
  type: dropdown
  options:
    - value: UP
      label: "UP"
      cost:
        add_cost: 0
        energy_cost: 0
    - value: DOWN
      label: "DOWN"
      cost:
        add_cost: 0
        energy_cost: 0
```

### `free_text`

Stores text and has no direct cost.

```yaml
- key: other
  label: "Other Roll Text"
  type: free_text
  visibility_when: source
  show_when: other
```

### `free_number`

Uses `default`, `min`, `max`, `step`, optional `rounding`, and optional `per_step`.

```yaml
- key: range
  label: "Range"
  type: free_number
  default: 0
  min: 0
  max: 10
  step: 2
  rounding: ceil
  per_step:
    increase:
      add_cost: 1
      energy_cost: 0
```

For values above `default`, `per_step.increase` is multiplied by the number of steps. For values below `default`, `per_step.decrease` is multiplied by the absolute number of steps. `rounding: ceil` rounds positive partial steps up; `rounding: floor` rounds down through integer division.

### `solutions`

Repeatable row field. The form submits parallel arrays named `<field>__<subfield>`, and the server validates/evaluates each row using `row_fields`.

Blank rows are ignored before `per_item` calculation. A `default` value can seed initial rows; when it is an array of objects with keys matching `row_fields`, each object populates one row:

```yaml
- key: counter_trait
  label: "Counter Trait"
  type: solutions
  default_count: 2
  default:
    - type: defense
      value: Reflex
    - type: defense
      value: Constitution
  options_source: traits_all
  row_fields:
    - key: type
      label: "Counter Type"
      type: dropdown
      default: defense
      options:
        - value: defense
          label: "Defensive Trait"
          cost:
            add_cost: 0
            energy_cost: 0
        - value: general
          label: "General Trait"
          cost:
            add_cost: 4
            energy_cost: 0
        - value: offense
          label: "Offensive Trait"
          cost:
            add_cost: 4
            energy_cost: 0
        - value: previous
          label: "Use result of previous"
          cost:
            add_cost: 3
            energy_cost: 1
    - key: value
      label: "Counter Trait"
      type: dropdown
      options_source: traits_all
  per_item:
    increase:
      add_cost: 0
      energy_cost: 0
    decrease:
      add_cost: 3
      energy_cost: 1
```

Add/remove buttons on a `solutions` field always add or remove exactly one row at a time. `default_count` controls how many rows are shown initially.

### `states`

Repeatable Enact State rows backed by `states.yaml`. Blank rows are ignored; partially filled rows are invalid.

```yaml
- key: states
  label: "States"
  type: states
  default_count: 1
  row_fields:
    - key: state_kind
      label: "State Type"
      type: dropdown
      default: ""
      options:
        - value: specific
          label: "Specific State"
          cost:
            add_cost: 0
            energy_cost: 0
        - value: general
          label: "General State (shift)"
          cost:
            add_cost: 0
            energy_cost: 0
    - key: specific_state
      label: "Specific State"
      type: dropdown
      options_source: states_specific
      visibility_when: state_kind
      show_when: specific
    - key: general_state
      label: "General State"
      type: dropdown
      options_source: states_general
      visibility_when: state_kind
      show_when: general
    - key: shift_amount
      label: "Shift Amount"
      type: free_number
      default: 1
      step: 1
      visibility_when: state_kind
      show_when: general
```

## Cost evaluation rules

Cost calculation is server-authoritative:

- The browser mirrors the calculation for live feedback only.
- The server recomputes costs on save.
- Schema-backed cards do not trust hidden build/cast values from the form.
- Legacy no-schema paths still use submitted build/cast fallback behavior.

Rules:

- Enactments and interactions start from `base_cost`.
- The first enactment added to an ability is free: its component `base_cost` is waived, so adding the first enactment costs no build or energy. Field-driven costs on that first enactment (checkboxes, dropdowns, numbers, states, etc.) still apply normally.
- Each enactment beyond the first pays its full `base_cost` plus the `additional_enactment` surcharge from `general.yaml`.
- Ability types start from `base_energy` and `base_action` where applicable.

- `checkbox` cost applies only when checked.
- `dropdown` field-level `cost` applies when the selected value is non-empty.
- Inline dropdown option `cost` also applies for the selected option.
- `free_number` applies `per_step.increase` or `per_step.decrease` from the configured `default`.
- `rounding: ceil` rounds positive partial steps up; `rounding: floor` rounds down.
- Enact State adds specific/general state row costs from `states.yaml` plus `additional_state` once per selected state after the first.

Example: this Ranged field charges `+1` build for every full 2m above 10m and floors partial steps.

```yaml
- key: range
  label: "Range"
  type: free_number
  default: 10
  min: 10
  max: 20
  step: 2
  rounding: floor
  per_step:
    increase:
      add_cost: 1
      energy_cost: 0
    decrease:
      add_cost: 0
      energy_cost: 0
```

## Visibility rules

`visibility_when` references another field's `key`. The field is active only when that controlling value equals `show_when`. Hidden/inactive fields do not contribute cost.

For checkboxes, the submitted checked value is the string `"true"`, not `"on"`.

```yaml
- key: item_dep
  label: "Has Item Dependency"
  type: checkbox
  cost:
    add_cost: -1
    energy_cost: 0
- key: item_name
  label: "Item Name"
  type: free_text
  visibility_when: item_dep
  show_when: "true"
```

Dropdown-controlled visibility:

```yaml
- key: source_trait
  label: "Trait"
  type: dropdown
  options_source: traits_offense
  visibility_when: source
  show_when: trait
```

## Default-driven visibility

When a controlling field has no submitted value, `visibility_when` falls back to that field's configured `default`. This ensures dependent fields become visible as soon as the default is active, without requiring an explicit user selection first.

```yaml
- key: engage_mode
  label: "Engage Roll Type"
  type: dropdown
  default: trait
  options:
    - value: trait
      label: "Trait Roll"
- key: engage_trait
  label: "Trait"
  type: dropdown
  options_source: traits_all
  visibility_when: engage_mode
  show_when: trait
```

With the default above, `engage_trait` is visible immediately when the card renders.

## Trait dropdown grouping and filtering

Dropdowns backed by `traits_general`, `traits_offense`, `traits_defense`, or `traits_all` are rendered as grouped `<optgroup>` lists by category.

For `solutions` rows that include a `type` field (for example `counter_trait`), the `value` dropdown is filtered to traits of the selected type. When the type is empty or `previous`, all traits are shown grouped.

## Option sources

Option sources are resolved in `static/js/builder.js`. Adding a new source name requires a JavaScript mapping.

| Source | Resolves to |
| --- | --- |
| `traits_general` | `D.generalTraits` from `traits.general` |
| `traits_offense` | `D.offenseTraits` from `traits.offense` |
| `traits_defense` | `D.defenseTraits` from `traits.defense` |
| `traits_all` | General + offense + defense traits |
| `dice_damage` | `D.damageDiceOptions` from `dice.damage` |
| `dice_generic` | `D.genericDieOptions` from `dice.generic` |
| `states_general` | `C.states.general_states` |
| `states_specific` | `C.states.specific_states` |
| `directions_all` | `D.directionOptions` |
| `directions` | `D.directionOptions` |
| `shift_directions` | `D.shiftDirectionOptions` |
| `trigger_timings` | `D.triggerTimings` |
| `aoe_trigger_timings` | `D.aoeTriggerTimings` |
| `knockout_options` | `D.knockoutOptions` |
| `reaction_triggers` | `D.reactionTriggers` |
| `ability_types` | `D.abilityTypes` |
| `enactment_types` | `D.allEnactmentTypes` |
| `interaction_types` | `D.interactionTypes` |

## States configuration

Specific states have fixed `add_cost` and `energy_cost`:

```yaml
specific_states:
  - id: taunted
    name: "Taunted"
    description: "You can only target a preset Target."
    add_cost: 2
    energy_cost: 0
```

General states use `min_shift`, `max_shift`, and `shift_cost` per absolute shift:

```yaml
general_states:
  - id: frightened
    name: "Frightened"
    description: "Negative trait shifts"
    min_shift: -6
    max_shift: 0
    shift_cost:
      add_cost: 1
      energy_cost: 0
```

Validation rules:

- `specific` rows require `specific_state`.
- `general` rows require `general_state` and `shift_amount` within that state's range.
- Unknown state IDs are rejected.
- Blank rows are ignored before surcharge calculation.
- `additional_state` is applied once per selected state after the first.

## How to add or change config

Safe workflows:

1. Change labels, costs, or existing option values directly in YAML.
2. Add an option to an existing field by appending to its `options` list.
3. Add a field to an existing type by appending a valid `FieldConfig` under `fields`.
4. Add a specific or general state in `states.yaml` and use an existing `states_*` option source.
5. Add a new ability, enactment, or interaction type by adding its config entry and updating compatibility lists such as `compatible_enactments` as needed.

After edits, validate with:

```bash
go test ./...
go vet ./...
go build ./...
node --check static/js/builder.js
go run .
```

For documentation changes, also run:

```bash
go run ./cmd/docs
```

## Known boundaries

- Generic YAML import/export is not fully schema-driven unless implemented separately.
- Existing saved abilities may not migrate cleanly when config keys change.
- Field keys are persisted in generic `Fields` maps, so renaming a field key changes saved-data compatibility.
- Config-defined type lists are used by the builder, but model/export compatibility may still depend on existing typed fields for some paths.
