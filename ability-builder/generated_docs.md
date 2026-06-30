# Ability Builder

## Introduction

The **Ability Builder** is the core system used to create actions, maneuvers, spells, techniques, and special effects. Rather than relying on predefined spell lists or class-locked abilities, this system allows abilities to be **constructed from modular components** with clearly defined rules and costs.

An Ability represents a **single intentional action** taken by a character. What that action does, who it affects, how it is resolved, and under which conditions it succeeds are all explicitly defined by the components chosen during creation.

Abilities are built from the following core elements:

*   **Enactments** — define _what happens_ (damage, healing, movement, shifts, persistent effects, etc.).
*   **Interactions** — define _how and to whom_ the Enactments are applied (self, direct, ranged, area, or area of effect).
*   **Validations** — define _if and how_ the Enactments succeed or fail.
*   **Rules and Perks** — modify default behavior, allowing abilities to break or bend standard limitations at a defined cost.

Every Ability must contain **at least one Enactment**. Additional Enactments may be added to create more complex effects, which are resolved **in sequence**. Each Enactment is evaluated independently unless explicitly overridden by a Perk.

The Ability Builder is intentionally **system-agnostic** with regard to flavor. A fireball, a sword technique, a healing prayer, or a mechanical trap are all created using the same underlying rules. The narrative description of an Ability is left to the player and GM, while the mechanical behavior remains explicit and predictable.

## Costs

Rules are free, but to apply perks there is a cost. The first cost is the **Add Cost** to add the Perk. Each level you gain **Ability Points** that can be spent to create abilities.

Then there is the **Energy Cost**. This cost is used to use your ability. Sometimes you do not have enough energy to use your ability. In this system it is allowed to still use your ability, but there is a catch: either you take damage equal to the amount of energy you are missing, or you only partially use your ability.

## Additional Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Adding an additional Enactment beyond the first | +1 | +1 |



# Execution

Execution is the most basic form for an Ability. It is simply the: "I want to do this now" Ability Type. Executions can be anything from casting a fireball to summoning a shield to block an attack or preparing a parry.

## Rules

*   **Enactments**: Has at least one Enactment (the first Enactment is free)
*   **Actions**: Costs 2 Actions to use
*   **Energy**: Costs 3 Energy to use

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Has item dependency | +0 | -1 |

## Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Damage | +1 | +2 |
| Enact Healing | +1 | +2 |
| Enact Movement | +0 | +1 |
| Enact Proficiency Shift | +1 | +2 |
| Enact Persistent Effect | +2 | +3 |

## Template

```yaml
ability:
  type: Execution
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  action_cost: 2
  enactments:
    - Type:
  perks:
```


# Minion

Minions are entities that players can create, summon, and control. They have default stats and actions, and can use Enactments created by the user if the appropriate perk is selected. Minions follow specific rules within the action economy.

## Rules

*   Minions have their own turn in the action economy.
*   Minions can perform one action per turn.
*   Minions have default stats: Health, Attack, Defense, Speed, Lifetime.
*   Minions can be summoned once per encounter.
*   Minions require a summoning cost (e.g., energy, mana).
*   Minions can be controlled by the player during their turn.
*   Minions can be dismissed by the player as a free action.
*   Minions have a default lifetime of 3 rounds.

## Default Stats

| Stat | Value |
| --- | --- |
| Health | 10 |
| Attack | 2d6 |
| Defense | 1d6 |
| Speed | 5m |
| Lifetime | 3 rounds |

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Has item dependency | +0 | -1 |

## Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Damage | +1 | +2 |
| Enact Healing | +1 | +2 |
| Enact Movement | +0 | +1 |
| Enact Proficiency Shift | +1 | +2 |
| Enact Persistent Effect | +2 | +3 |


# Phase

Phases are a state or passive ability that lasts for a predefined amount of time. They exist to buff or nerf someone for a specific number of rounds. A Phase lasts for a few rounds, after which the Reverse Phase starts and lasts just as long as the original phase did.

## Rules

*   Costs 3 Energy to Use.
*   After activation, Phase is active for 2 rounds.
*   Phase ends at the start of the 2nd turn of the character.
*   When Phase ends, the Reverse Phase starts.
*   During the Reverse Phase, no new Phases can be started for the character.
*   Phase will have an Enactment assigned to it.
*   The Enactment can be triggered as a free action at the end of the character's turn.
*   Reverse Phase will have a Bad Enactment assigned to it.
*   Bad Enactment will be applied to the character.
*   Bad Enactment must be used at the end of the character's turn as a free action.
*   If no Bad Enactment is chosen, the Bad Enactment will be the reverse of the original Enactment.
*   Phase has a knockout requirement.
*   If any knockout requirement is met, the Phase ends (and the "Bad Enactment" starts).
*   The Reverse Phase cannot be cancelled by the knockout.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| All knockout requirements have to be met | +0 | +3 |
| Knockout can be used on the reverse phase | +0 | +3 |
| No knockout possible | +0 | +5 |
| Has item dependency | +0 | -1 |

## Knockout Requirements

| Description | Add Cost |
| --- | --- |
| None | +0 |
| You take damage | +0 |
| You fall unconscious | +0 |
| You die | +0 |
| You get grabbed or restrained | +0 |
| You move voluntarily | +0 |
| You are moved by another effect | +0 |
| You fail a validation | +0 |
| You use another phase | +0 |
| You lose line of sight to target | +0 |
| Target moves out of range | +0 |
| Target falls unconscious | +0 |
| Target dies | +0 |
| Target succeeds on a counter roll | +0 |
| Phase duration expires | +0 |
| You run out of energy | +0 |

## Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Damage | +1 | +2 |
| Enact Healing | +1 | +2 |
| Enact Proficiency Shift | +1 | +2 |

## Template

```yaml
ability:
  type: Phase
  phase_duration: 2 rounds
  reverse_phase_duration: 2 rounds
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  enactments:
    - Type:
  Perks:
```


# Preparation

Just like a Reaction, a Preparation works outside the regular turn order. They follow the exact same rules as a Reaction but instead of being passively on the background, a Preparation must cost an action to prepare.

## Rules

*   Can only be used once per round.
*   Costs 2 actions.
*   Costs 3 Energy to Use.
*   Always has at least one Trigger (Pick one from the list below, first one is free).
*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within 1m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Has item dependency | +0 | -1 |

## Triggers

_No triggers available._

## Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Damage | +1 | +2 |
| Enact Healing | +1 | +2 |
| Enact Movement | +0 | +1 |
| Enact Proficiency Shift | +1 | +2 |
| Enact Persistent Effect | +2 | +3 |

## Template

```yaml
ability:
  type: Preparation
  range: 1
  uses: 1
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  action_cost: 2
  trigger: <trigger name here>
  enactments:
    - Type:
  Perks:
```


# Reaction

Reactions are Abilities that trigger outside your normal action economy. Reactions trigger when someone else (or you) does something. When the trigger happens, the linked Enactment is executed.

## Rules

*   Can only be used once per round.
*   Does not cost an action.
*   Costs 3 Energy to Use.
*   Always has at least one Trigger (Pick one from the list below, first one is free).
*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within 1m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Has item dependency | +0 | -1 |

## Triggers

| Description | Add Cost |
| --- | --- |
| Someone runs away from you | +2 |
| Someone runs towards you | +2 |
| Someone moves past you | +2 |
| Someone gets healed | +2 |
| Someone takes damage | +2 |
| Someone does a skill check | +2 |
| Someone starts casting an ability | +2 |
| A turn ends within range | +2 |
| Someone enters range | +2 |
| Someone leaves range | +2 |
| Someone fails a validation | +2 |
| Someone succeeds on a validation | +2 |
| Someone becomes affected by an enactment | +2 |
| You take damage | +2 |
| You are targeted by an ability | +2 |
| An ally within range takes damage | +2 |
| An ally within range gets healed | +2 |
| Someone is moved by an effect | +2 |
| A persistent effect triggers | +2 |
| A minion is summoned within range | +2 |

## Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Damage | +1 | +2 |
| Enact Healing | +1 | +2 |
| Enact Movement | +0 | +1 |
| Enact Proficiency Shift | +1 | +2 |
| Enact Persistent Effect | +2 | +3 |

## Template

```yaml
ability:
  type: Reaction
  range: 1
  uses: 1
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  trigger: <trigger name here>
  enactments:
    - Type:
  Perks:
```


# Enact Damage

Enact Damage abilities allow characters to inflict harm on their enemies.

## Rules

*   **Damage Dice**: The default damage dice is 1d4.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Change Damage Dice to one of your traits | +0 | +3 |
| Add a flat +1 bonus to the result | +0 | +2 |
| Add an Offensive Trait Dice to the Damage Dice | +2 | +4 |
| Will Always Resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
enactments:
  - type: Enact Damage
    damage_dice: <dice here>
    is_optional: False # First enactment is usually mandatory
    base_enactment_energy_cost: 0 # 0 if first enactment, otherwise 1
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Add Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
    interactions:
      - type:
          validation:
```


# Enact Healing

Enact Healing abilities allow characters to restore health to themselves or their allies.

## Rules

*   **Healing Dice**: The default healing dice is 1d4.
*   **Interaction Type**: If the interaction type is Self or Direct, no validation is required.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Change Heal Dice to one of your traits | +0 | +3 |
| Add a flat +1 bonus to the result | +0 | +2 |
| Add Medicine Trait Dice to the heal effect | +1 | +3 |
| Will Always Resolve | +2 | +4 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
enactments:
  - type: Enact Healing
    healing_dice: <dice here>
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
    interactions:
        validation:
```


# Enact Movement

Enact Movement abilities allow characters to manipulate the position of themselves or their targets. Movement abilities add a dynamic element to gameplay, enabling tactical maneuvers and creative solutions to challenges.

## Rules

*   **Direction**: The target will move in one direction relative to an origin. Possible directions include Up, Down, Away, Towards, Forward, Left, Right, Free (extra cost).
*   **Distance**: The target will move 1 meter by default.
*   **Origin**: The default origin is the engager or item/location from previous enactment.
*   **Obstacle**: If the target moves into an obstacle, they take 1d4 damage.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Add another option for the direction | +0 | +1 |
| Change Origin to something else | +1 | +2 |
| Change the total movement to any other trait | +0 | +3 |
| Will Always Resolve | +1 | +3 |
| Use the result of another roll for this entry | +1 | +3 |
| Free direction (extra cost) | +1 | +2 |

## Template

```yaml
enactments:
  - type: Enact Movement
    minimal_distance: 1m
    origin: engager
    direction_options:
      - <Direction>
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
    interactions:
      - type:
          validation:
```


# Enact Persistent Effect

The Enact Persistent Effect applies a lingering effect to a target, such as fire, frost, or poison damage. By default, the effect lasts for 2 rounds and triggers at either the start of the target's turn or the end of the engager's turn.

## Rules

*   **Duration**: Lasts 2 rounds by default.
*   **Trigger Timing**: The effect triggers at either the start of the target's turn or the end of the engager's turn.
*   **Solutions**: Targets can spend one action to attempt to remove the effect using the provided solution. There must be two solutions, which can be any Trait Roll.
*   **Applies a Single Enactment**: The persistent effect applies a single enactment (e.g., Enact Damage, Enact Healing).

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Remove one Solution option | +1 | +3 |
| Add another Effect | +2 | +4 |
| Will Always Resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Effects

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Damage | +1 | +2 |
| Enact Healing | +1 | +2 |
| Enact Movement | +0 | +1 |
| Enact Proficiency Shift | +1 | +2 |

## Template

```yaml
enactment:
  - type: Enact Persistent Effect
    duration: 2 rounds
    trigger_timing: Start of Target's Turn or The end of the engager's turn.
    solutions:
      - Dexterity
      - Constitution
    is_optional: True
    base_enactment_energy_cost: 2
    effects:
      - type: <Enactment here type>
    interactions:
      - type:
          validation:
```


# Enact Proficiency Shift

Enact Proficiency Shift abilities allow characters to temporarily enhance or weaken Traits.

## Rules

*   **Shift Direction**: Shift a Proficiency Tier from a Trait either up or down.
*   **Single Use**: The shift only has one use.
*   **Trait Check**: The next time a Trait check is made with the shifted Trait, you must use the shifted Proficiency.
*   **Reset**: Using the shifted Proficiency resets the Trait back to its original Proficiency.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| You may choose if you use the Shifted Proficiency or not | +0 | +2 |
| Will Always Resolve | +2 | +4 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
enactments:
  - type: Enact Proficiency Shift
    shifted_trait: <trait here>
    shift_direction: <UP DOWN or>
    shift_amount: 1
    shift_uses: 1
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
    interactions:
      - type:
          validation:
```


# Enact State

Enact State abilities apply a state or condition to a target (e.g., stunned, poisoned). This enactment type is currently a work in progress.

## Rules

*   **State**: Applies a condition to the target.

## Perks

_No perks available._

## Template

```yaml
enactments:
  - type: Enact State
    state: <state here>
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
```


# Area of Effect

An Area of Effect (AoE) Interaction functions similarly to an Area Interaction, but its effects persist for several rounds. While an Area Interaction might be like a single-use bomb, an AoE Interaction is akin to a bomb that detonates every round.

The effect of the AoE does not trigger immediately. Instead, it activates either at the start of a character's turn within the AoE or at the end of the Engager's turn.

## Rules

*   **Validation**: The interaction must have a validation.
*   **Radius**: The default radius is 1 meter.
*   **Range**: The default range is 0 meters.
*   **Origin**: The point of origin for the radius is the engager.
*   **Duration**: The effect lasts for 2 rounds.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Engager is immune to the effect | +0 | +2 |
| Change Origin to something else | +1 | +2 |
| Will always resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
interactions:
  - type: Area of Effect
    radius: 1m
    range: 0m
    origin: Engager
    duration: 2 rounds
    immunity: false
    trigger_conditions:
      - Entering the Area of Effect
      - Start of character's turn within the Area of Effect
    perks:
      - description: <insert description of perk here>
        add_cost: <cost of the perk>
        amount: <amount of times the perk is chosen>
        total_add_cost: <total add cost>
        energy_cost: <energy cost to use>
        is_optional: <True/False>
    validation:
      engagement_roll: <pick an Offensive Trait>
      counter_roll: <pick two Defensive Traits>
```


# Area

Area Interactions encompass actions like bombs, splash potions, and traps. These interactions always have a defined **Radius** and **Range**:

*   **Radius**: This determines the area where the Enactment will take effect.
*   **Range**: This specifies how far from the user the point of origin is set. By default, the point of origin is 0m from the user.

## Rules

*   **Validation**: The interaction must have a validation.
*   **Radius**: The default radius is 1 meter.
*   **Range**: The default range is 0 meters.
*   **Origin**: The point of origin for the radius is the engager or item/location from previous enactment.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Change Origin to something else | +1 | +2 |
| Will always resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
interactions:
  - type: Area
    radius: 1m # Default radius for Area interactions
    range: 0m # Default range for Area interactions
    origin: Engager # Point of origin is the Engager
    perks:
      - description: <insert description of perk here>
        add_cost: <cost of the perk>
        amount: <amount of times the perk is chosen>
        total_add_cost: <total add cost>
        energy_cost: <energy cost to use>
        is_optional: <True/False>
    validation:
      engagement_roll: <pick an Offensive Trait>
      counter_roll: <pick two Defensive Traits>
```


# Direct

## Rules

*   Has a Validation.
*   Target is a single character.
*   Target must be within 1m of your character.

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Will always resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
interactions:
  - type: Direct
    engager: Self
    target_amount: 1
    range: 1m
    perks:
      - description: <insert description of perk here>
        add_cost: <cost of the perk>
        amount: <amount of times the perk is chosen>
        total_add_cost: <total add cost of this perk = cost * amount>
        energy_cost: <energy cost to use>
        is_optional: <True/False>
    validation:
      engagement_roll: <pick an Offensive Trait>
      counter_roll: <pick two Defensive Traits>
```


# Ranged

Ranged Interactions include actions like using bows, guns, and boomerangs. These interactions offer an increased range compared to Direct Interactions but come with a lower success rate due to a penalty on the Engagement Roll.

## Rules

*   Has a Validation
*   Target is a single character.
*   Target must be within 10m of your character.
*   Target must be visible.
*   Target must not be obstructed.
*   Engagement roll result is lowered by 2

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Engagement does not have to be visible | +1 | +3 |
| Engagement may be obstructed | +1 | +3 |
| Remove the Engagement Roll Penalty | +1 | +3 |
| Will always resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
interactions:
  - type: Ranged
    engager: Self
    target_amount: 1
    range: 10m # Default range for Ranged interactions
    visibility: Visible # Target must be visible
    obstruction: Not obstructed # Target must not be obstructed
    perks:
      - description: <insert description of perk here>
        add_cost: <cost of the perk>
        amount: <amount of times the perk is chosen>
        total_add_cost: <total add cost>
        energy_cost: <energy cost to use>
        is_optional: <True/False>
    validation:
      engagement_roll: <pick an Offensive Trait> - 2
      counter_roll: <pick two Defensive Traits>
```


# Self

## Rules

*   Has a Validation
*   Engager is yourself
*   Target is yourself
*   Counter Roll is a d8

## Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Validation counter roll is replaced by a Generic Dice: d8 | +0 | +2 |
| Will always resolve | +3 | +5 |
| Use the result of another roll for this entry | +1 | +3 |

## Template

```yaml
interactions:
  - type: Self
    engager: Self
    target: Self
    validation:
      engagement_roll: Power
      counter_roll: d8 # d8 is default
      perks:
        - description: <Perk Description>
          add_cost: <Cost>
          amount: <Amount>
          total_add_cost: <Total Add Cost>
          energy_cost: <Total Cost Energy>
          is_optional: <True/False>
```


# Validations

## Introduction

Here, you'll find the guidelines and options for customizing your engagement and counter rolls.

## Rules

*   **Engagement Roll**: This is an Offensive Trait used to initiate actions against a target.
*   **Counter Roll**: This involves two Defensive Traits, allowing the target to choose how they respond to the attack.

## Engagement Roll Modes

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Offensive Trait by default | +0 | +0 |
| Engage Roll is replaced by a Generic Roll (default 1d6) | +0 | -2 |
| Replace the Engagement Roll to any other Trait | +1 | +3 |
| Use the result of another roll for this entry | +1 | +3 |

## Counter Roll Types

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Defensive Trait (default) | +0 | +0 |
| Replace one of the Counter Rolls to a General Trait | +0 | +4 |
| Replace one of the Counter Rolls to an Offensive Trait | +0 | +4 |
| Use the result of another roll as counter | +1 | +3 |

## Tier Shifts

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Shift Dice Tier of Generic Counter Roll UP | +1 | +3 |
| Shift Dice Tier of Generic Counter Roll DOWN | +0 | -2 |
| Shift Dice Tier of Generic Engagement Roll UP | +1 | +3 |
| Shift Dice Tier of Generic Engagement Roll DOWN | +0 | -2 |

## Template

```yaml
validation:
  engagement_roll: <pick an offensive trait>
  counter_roll: <pick two defensive traits>
  perks:
    - description: <insert description of perk here>
      add_cost: <cost of the perk>
      amount: <amount of times the perk is chosen>
      total_add_cost: <total add cost of this perk = cost * amount>
      energy_cost: <total energy cost>
      is_optional: <True/False>
```


# Leveling

## Introduction

As you level up, your character gains a deeper understanding of their powers, techniques, and spells. This growth is represented by **Ability Points**. Ability Points are spent to pay the **Add Cost** of Perks, Enactments, Interactions, and Validations when constructing or upgrading your Abilities.

## Ability Points

At Level 1, a character starts with a base pool of Ability Points. As they level up, they gain a steady stream of new points, with larger spikes at milestone levels (Level 5 and Level 10).

These points are permanently invested into your abilities during character creation or level-ups.

### Upgrading Abilities

You do not need to create a brand new Ability every time you level up. You can spend your newly gained Ability Points to upgrade an existing Ability by adding new Perks, extending its Range, or attaching additional Enactments.

### Refunding Ability Points

Some Perks in the Ability Builder apply drawbacks or restrictions to an Ability (such as giving it an Item Dependency or increasing its Action Cost). These Perks have a **negative Add Cost**. Taking these drawbacks refunds Ability Points, allowing you to spend them elsewhere on the same Ability to make it more powerful.

## Leveling Table: Ability Points

| Level | Ability Points Gained | Total Ability Points |
| --- | --- | --- |
| **1** | Base | 10 |
| **2** | +2 | 12 |
| **3** | +3 | 15 |
| **4** | +2 | 17 |
| **5** | +4 | 21 |
| **6** | +2 | 23 |
| **7** | +3 | 26 |
| **8** | +2 | 28 |
| **9** | +3 | 31 |
| **10** | +5 | 36 |