# Ability Builder: Code vs. System Documentation

This document lists the differences between the current ability-builder implementation and the original Blok2 system documentation in `docs/Blok2ttrpg/Blok2ttrpg/Modules/Ability Builder/`.

## Ability Types

| System Docs | Current Code | Notes |
|-------------|--------------|-------|
| **Execution** | Implemented | Basic instant ability. |
| **Reaction** | Implemented | Trigger-based ability with range/uses. |
| **Phase** | Implemented | Phase + reverse phase rounds, knockouts. |
| **Preparation** | **Not implemented** | Docs describe a separate Preparation type (Reaction-like but costs 2 actions). The code has no `Preparation` ability type. |
| **Minion** | Implemented (WIP) | Both mark Minions as work-in-progress. |

## Base Rules & Costing

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Add Cost** vs. **Energy Cost** | Docs keep these separate for every perk/rule. | `TotalCost()` sums `BuildCost + CastCost` together server-side. The live builder JS keeps them separate (`data-build` and `data-cast`). | Backend cost model merges the two; the UI does not. |
| **Execution energy** | 3 energy base. | 3 energy base (`energy = 3`). | Matches. |
| **Execution action cost** | 2 actions base. | Captured as `action_steps`, but no "Total Action Cost" display exists. | Action cost is stored but not surfaced in the header. |
| **Additional Enactment cost** | +1 build per extra enactment. | +1 build per extra enactment (`acts.length - 1`). | Matches for build cost. The docs also mention per-enactment base energy cost (e.g., 1 for non-first), which is not reflected. |

## Ability-Type Perks

| Type | Docs | Code | Notes |
|------|------|------|-------|
| **Execution** | Has item dependency, energy ±, action ± (with separate Add/Energy costs). | Two simple step-selects (`energy_steps`, `action_steps`) and an item dependency checkbox. | Code does not present each perk individually; only the net steps are configurable. |
| **Reaction** | Add range, add uses, energy ±, item dependency. | Range/uses inputs, item dependency, but no "Add uses per round" or explicit energy-step perks. | Range/uses cost math is hard-coded, not a perk list. |
| **Phase** | Add/remove phase rounds, all-knockout requirement, reverse-phase knockout, no-knockout, energy ±, item dependency. | Phase/reverse rounds, all-knockout, reverse knockout, no-knockout, item dependency. | Energy ± perks are missing. No explicit "item dependency" UI for Phase (the checkbox is rendered but the field exists in the form). |
| **Minion** | HP, attack, defense, speed, lifetime, extra actions, extra abilities, item dependency, energy ±. | Only HP bonus and extra lifetime (plus item dependency). | Minion implementation is far more limited than docs. |

## Triggers (Reaction / Preparation)

| Docs | Code | Notes |
|------|------|-------|
| e.g., "Someone runs away from you", "Someone has to do a Defense roll", "Someone does a skill check" | e.g., "Target moves away from engager", "Target makes a trait check", "Target fails a validation" | Trigger names differ and the code list is longer/more granular. First trigger is free in both; code does not charge for additional triggers. |

## Enactments

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Enact Damage** | Perks: shift dice tier, change to trait, flat bonus, offensive trait die, always resolve, use previous. | Source die tier, trait source, flat bonus, offensive trait, always resolve, use previous. | Functionally similar, but not presented as a purchasable perk list. |
| **Enact Healing** | Perks: shift dice tier, change to trait, flat bonus, Medicine trait, always resolve, use previous. | Same as damage plus Medicine trait. | Matches. |
| **Enact Movement** | Perks: distance, extra direction, change origin, change distance to trait, always resolve, use previous. | Distance, multiple directions, other origin, free direction, always resolve. | "Change the total movement to any other trait" perk not implemented. |
| **Enact Proficiency Shift** | Perks: extra uses, shift again, choose whether to use shifted proficiency, always resolve, use previous. | Shift amount, uses, direction, always resolve. | "Choose whether to use shifted proficiency" and "use previous" not implemented. |
| **Enact Persistent Effect** | Contains an array of `effects` (each an enactment); perks for duration, solutions, extra effects, always resolve. | Single `effect_type` dropdown plus duration/solutions; no nested enactment structure. | Code cannot represent multiple effects or a full nested enactment as in the docs. |
| **Enact State** | Mentioned as WIP in docs. | **Not present** at all. | Docs list a sixth type; code has only five. |
| **Optional perks** | Docs mark perks as `is_optional: True/False` so players can drop them to save energy. | **Not implemented** — no optional perk concept. | The code always applies the chosen options. |
| **Base enactment energy cost** | Docs note `base_enactment_energy_cost: 0` for first, `1` for additional. | Not explicitly modeled; only the +1 build cost for extra enactments is applied. | Energy surcharge for additional enactments is not reflected. |

## Interactions

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Self** | Has a validation; engager/target = self; counter = d8; perks for generic counter, always resolve, use previous. | Renders "Self + Target = Self + Counter = d8" with no validation config; always-resolve and use-previous available. | Validation card is still shown in the block but Self docs say it has a validation. |
| **Direct** | Range 1m, 1 target; perks for range, extra target, always resolve, use previous. | Range/targets inputs, always resolve, use previous. | No explicit "extra target" perk; just a numeric input. |
| **Ranged** | Range 10m, 1 target, visible, not obstructed, -2 engagement penalty; perks for range, visibility, obstruction, remove penalty, targets, always resolve. | Range/targets, visibility, obstruction, remove penalty, always resolve. | "Decrease range" perk is implemented as a negative step, not a separate perk. |
| **Area** | Radius 1m, range 0m, origin engager; perks for radius, range, origin, always resolve. | Radius/range/origin inputs, always resolve. | No explicit perk list. |
| **Area of Effect** | Radius 1m, range 0m, duration 2 rounds, engager immunity, timing conditions; perks for radius, range, duration, immunity, origin, always resolve. | Radius/range/duration/origin/immune/timing inputs, always resolve. | No explicit perk list. |
| **Interaction inheritance** | Docs say an Enactment without its own Interaction inherits the previous Enactment's Interaction. | **Not implemented** — every Enactment block must select its own Interaction. | The form does not auto-fill or inherit from the previous block. |
| **Allies / Enemies** | Docs ask whether the ability affects allies or only enemies. | **Not implemented** — no ally/enemy targeting flag. | |
| **Engager included/excluded** | Docs ask whether the user is included in the effect. | Only implemented for AoE via "engager immune"; no general flag. | |

## Validations

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Engagement roll** | Offensive Trait by default; perks to replace with generic, other trait, another roll, etc. | Trait/generic/other/previous modes; trait can be offense/defense/general. | Code allows more modes (generic, other, previous) and trait categories. |
| **Counter roll** | Two Defensive Traits by default; perks to replace, remove, generic, tier shift, etc. | Multiple counter entries; each can be defense/general/offense/previous trait. | Docs say two counters; code allows any number. |
| **Only one counter option** | Removing a counter option costs +3 build and +1 energy. | Implemented: 1 counter adds +3 build / +1 energy. | Matches. |
| **Counter as generic dice** | Default generic counter is d12. | Generic die options are `d6, d8, d10, d12`. | Close; docs say default d12 for counters. |
| **Generic engagement dice** | Default generic engagement is d6. | Generic die options are `d6, d8, d10, d12`. | Matches. |

## YAML / Export Format

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Perks in YAML** | Every template includes `perks:` arrays with `description`, `add_cost`, `amount`, `total_add_cost`, `energy_cost`, `is_optional`. | **Perks are not emitted.** The exporter explicitly says "Perks are no longer emitted." | Exported YAML cannot be used to reconstruct perk choices. |
| **Nested structure** | Docs use `enactments` → `interactions` (plural) → `validation`. | Code uses `enactments` → `interaction` (singular) → `validation`. | Model also uses `Interaction` pointer (singular) but JSON tag is `interactions`. |
| **Cost fields** | Docs show `base_enactment_energy_cost`, `total_add_cost`, `energy_cost` on perks. | Code exports `build_cost` and `cast_cost` only when non-zero. | No granular perk cost breakdown. |
| **Ability type fields** | `energy_cost`, `action_cost` are explicit. | `energy_cost`, `action_cost` exported, plus `energy_adjustment`, `action_adjustment` for Execution. | Extra adjustment fields not in docs. |
| **Validation** | Docs show `validation: n/a` for Self/Direct. | Code exports `validation: n/a` when missing. | Matches. |

## Leveling & Ability Points

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Ability Points** | Docs define a leveling table (10–36 points by level 10) and say points are spent to pay Add Cost. | **Not implemented** — the app is only a cost calculator; no character level or point pool is tracked. | |
| **Refunding points** | Docs say negative Add Cost perks refund points. | Negative-cost options exist (e.g., item dependency), but no point pool is enforced. | |
| **Upgrading abilities** | Docs describe spending new points to add perks. | **Not implemented** — edit mode loads a saved ability but there is no "upgrade" flow. | |

## UI / Workflow Differences

| Topic | Docs | Code | Notes |
|-------|------|------|-------|
| **Total Ability Cost** | Energy cost is a first-class cost. | Recently added to the sticky header as "Total Ability Cost" (sum of `data-cast`). | Before this change, only build cost was shown. |
| **Collapsible sections** | Docs do not specify UI layout. | Enactment blocks and the Enactment Type sub-section are collapsible; Interaction/Validation cards are collapsible. | Cosmetic convenience not addressed by docs. |
| **Perk selection** | Docs imply discrete perk choices with amounts. | Code uses direct inputs (dice tier, flat bonus, targets, range, etc.) rather than an explicit perk list. | The UX is different even when the resulting cost is similar. |

## Summary

The current implementation captures the *core loop* of the Ability Builder (Enactment → Interaction → Validation) and the cost math for the most common options, but it is a **simplified/subset version** of the documented system:

- Missing: **Preparation** ability type.
- Missing: **Enact State** enactment type.
- Missing: full **perk** modeling with `is_optional`, `amount`, and `total_add_cost`.
- Missing: **interaction inheritance** between Enactments.
- Missing: ally/enemy targeting and explicit engager inclusion/exclusion.
- Missing: **leveling / Ability Points** tracking.
- Missing: **nested effects** inside Persistent Effects.
- Different: many ability-type and interaction perks are represented as raw numeric inputs instead of discrete perks.
- Different: YAML export does not include perks and uses singular `interaction:` instead of the documented plural `interactions:`.

The code works as a cost calculator and YAML exporter for the subset it supports, but cannot yet represent every option described in the original rules.
