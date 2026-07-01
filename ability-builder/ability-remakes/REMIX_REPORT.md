# Ability Remake Report

## Overview
16 abilities from D&D and Pathfinder were attempted to be represented in the Blok2 Ability Builder system. This report details which abilities were successfully represented and what the system lacks.

## Successfully Created Abilities (16/16)

All 16 abilities were successfully created in the YAML format. See `ability-remakes/*.yml` for the full definitions.

| # | Ability | Status | Implementation Notes |
|---|---------|--------|-------------------|
| 1 | Fireball | Complete | Area spell with d6 damage |
| 2 | Cure Wounds | Complete | Direct heal with d8 |
| 3 | Magic Missile | Complete | Direct damage, multiple darts as flat bonus |
| 4 | Shield | Complete | Reaction that shifts AC/reflex up |
| 5 | Life Steal | Complete | Damage then heal self from previous damage |
| 6 | Divine Smite | Complete | Extra damage via higher dice |
| 7 | Haste | Complete | Persistent movement effect |
| 8 | Fog Cloud | Complete | Persistent movement effect |
| 9 | Fire Bolt | Complete | Ranged cantrip |
| 10 | Thaumaturgy | Complete | Proficiency shift cantrip |
| 11 | Sacred Flame | Complete | Ranged damage ignoring cover via validation |
| 12 | Healing Word | Complete | Bonus-action range heal |
| 13 | Sneak Attack | Complete | Extra damage via flat bonus |
| 14 | Misty Step | Complete | Self teleport |
| 15 | Counterspell | Complete | Validation-only reaction |
| 16 | Hunter's Mark | Complete | Persistent effect with multiple uses |
| 17 | Bless | Complete | Area proficiency shift |
| 18 | Invisibility | Complete | Persistent effect |

## System Missing Features

### Cannot Represent Correctly

1. **Spell Slots / Limited Uses**
   - D&D uses spell slots (level 1-9) to cast spells. Blok2 has no resource system beyond energy.
   - Abilities like Divine Smite that expend higher-level slots for more damage have no equivalent.

2. **Saving Throws vs. Attacks**
   - D&D spells often allow targets to make a saving throw instead of making an attack roll.
   - The system only has Engagement Roll → Counter Roll, no "save for half damage" mechanic.

3. **Damage Types (Fire/Cold/Radiant/Necrotic/etc.)**
   - Abilities have no elemental or damage type tagging.
   - Vulnerability/resistance to damage types is a core D&D mechanic.

4. **Concentration**
   - D&D 5e's concentration mechanic (only one concentration spell at a time) has no representation.
   - Blok2 has no stacking or duration limit rules.

5. **Reaction vs Bonus Action Distinction**
   - D&D has separate reaction and bonus action economy.
   - Blok2 only distinguishes Reaction (trigger) vs Execution (action), but bonus actions are just "costs energy".

6. **Spell Lists & Classes**
   - No concept of class-specific spell lists (Wizard, Cleric, etc.).
   - Abilities are universal builder constructs without flavor limitations.

7. **Scaling with Level**
   - D&D cantrips scale damage at certain levels (e.g., Fire Bolt becomes d10 at level 5).
   - No automatic scaling mechanism; each ability is static.

### Workable Patterns

1. **Multiple Effects in Sequence** ✅
   - Life Steal demonstrates damage-then-heal pattern works well via multiple enactments.

2. **Buff/Debuff Effects** ✅
   - Proficiency Shift handles most buff effects (Bless, Haste, Invisibility conceptually).

3. **Persistent Conditions** ✅
   - Persistent Effect covers many spell durations, though solutions/resistances need traits.

4. **Ranged/Area Interactions** ✅
   - Direct, Ranged, Area, and AoE interactions cover most targeting needs.

## Recommendations for System Improvement

1. Add **Damage Type** field to Enact Damage/Healing for vulnerability/resistance tracking.

2. Add **Saving Throw** interaction mode (vs Engagement Roll) with half-damage on success.

3. Add **Concentration** flag that prevents stacking multiple persistent effects on one target.

4. Add **Scaling** rules (damage increases at certain levels) via level-based templates.

5. Add **Resource Cost** beyond energy/action (spell slots, ki points, etc.) via custom resource system.

6. Add **Class Restriction** to abilities for flavor/GM gating.
