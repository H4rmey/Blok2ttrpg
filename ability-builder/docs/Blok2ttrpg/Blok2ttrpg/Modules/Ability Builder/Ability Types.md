# Ability Types
## Introduction

Different types of Abilities exist, each functioning differently:

*   **Execution**: Performed instantly during a character’s turn.
*   **Reaction**: Triggered in response to an event or enemy action.
*   **Phase**: Requires setup or charging time before execution.
*   **Minion**: Summons entities that act independently.

## Execution

Execution is the most basic form for an Ability. It is simply the: “I want to do this now” Ability Type. Executions can be anything from casting a fireball to summoning a shield to block an attack or preparing a parry.

### Rules

*   **Enactments**: Has at least one Enactment (the first Enactment is free)
*   **Actions**: Costs Two Actions to use
*   **Energy**: Costs 3 Energy to use

### Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Has Item Dependency | 0 | \-1 |
| Reduce Energy cost by 1 (up to a minimum of one) | \-1 | 3 |
| Increase Energy cost by 1 | +1 | \-2 |
| Reduce Action Cost by 1 | +1 | 4 |
| Increase Action Cost by 1 | 0 | \-2 |

### Compatible Enactments

_(Note: Adding additional enactments beyond the first free one costs extra)_

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Adjustment | 1 | 2 |
| Enact Persistent Effect | 2 | 3 |
| Enact Damage | 1 | 2 |
| Enact Healing | 1 | 2 |
| Enact Movement | 0 | 1 |

### Template

```yaml
ability:
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  enactments:
    - Type:
  perks:
```

### Example 1

```yaml
ability:
  type: Execution
  has_item_dependency: No
  energy_cost: 3
  action_cost: 2
  enactments:
    - type: Enact Damage
      damage_dice: d8 + Power
      perks:
        - description: Shift Dice Tier of damage up
          cost: 2
          amount: 2
          total_cost: 4
        - description: Add an Offensive Trait Dice to the Damage Dice
          cost: 3
          amount: 1
          total_cost: 3
      interactions:
        - type: Ranged
          engager: Self
          target_amount: 1
          range: 10m
          target_may_be_visible: False
          target_may_be_obstructed: False
          validation:
            engagement_roll: Power - 2
            counter_roll: Relfex or Constitution

    - type: Enact Healing
      healing_dice: enactment1.damage_dice.result
      perks:
        - description: Use the result of another roll for this entry
          cost: 4
          amount: 1
          total_cost: 4
      interactions:
        - type: Self
          engager: Self
          target: Self
          validation: n/a
```

## Reaction

Reactions are Abilities that trigger outside your normal action economy. Reactions trigger when someone else (or you) does something. When the trigger happens, the linked Enactment is executed. For example, you could have a reaction that triggers whenever someone runs towards you, Enacting a healing effect on yourself.

### Rules

*   Can only be used once per round.
*   Does not cost an action.
*   Costs 3 Energy to Use.
*   Always has at least one Trigger (Pick one from the list below, first one is free).
*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within 1m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

### Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Add 1 meter to reaction range | 0 | 1 |
| Add one more use per round | +1 | 4 |
| Reduce Energy cost by 1 (up to a minimum of one) | \-1 | 3 |
| Increase Energy cost by 1 | +1 | \-2 |
| Has Item Dependency | 0 | \-1 |

### Triggers

_(Note: First trigger is free. Below are the costs to add additional alternative triggers to the same ability)_

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Someone runs away from you | 0 | 2 |
| Someone runs towards you | 0 | 2 |
| Someone has to do a Defense roll | 0 | 2 |
| Someone takes damage | 0 | 2 |
| Someone gets healed | 0 | 2 |
| Someone gets an adjustment (Buff/Penalty) | 0 | 2 |
| Someone does a skill check | 0 | 2 |
| Someone summons a minion | 0 | 2 |
| Someone grabs your character | 0 | 2 |
| You walk towards someone | 0 | 2 |
| You walk away from someone | 0 | 2 |

### Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Adjustment | 1 | 2 |
| Enact Persistent Effect | 2 | 3 |
| Enact Damage | 1 | 2 |
| Enact Healing | 1 | 2 |
| Enact Movement | 0 | 1 |
| Enact other execution | 2 | 4 |

### Template

```yaml
ability:
  range: 1
  uses: 1
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  trigger: <trigger name here>
    skills:
      - <skill1> # this is required for the trigger "someone does a skill check"
      - <skill2> # each skill added is another option so the cost of adding this is stacked
  enactments:
    - Type:
  Perks:
```

### Example

```yaml
ability:
  type: Reaction
  range: 1
  uses: 1
  has_item_dependency: No
  energy_cost: 3
  trigger: Someone runs towards you
  enactments:
    - type: Enact Healing
      damage_dice: d10 + Medicine + 4
      perks:
        - description: Shift Dice Tier of Heal up
          cost: 2
          amount: 3
          total_cost: 6
        - description: Add Medicine Trait Dice to the heal effect
          cost: 1
          amount: 1
          total_cost: 1
        - description: Add a flat +1 bonus to the result
          cost: 3
          amount: 4
          total_cost: 12
      interactions:
        - type: Direct
          engager: Self
          target_amount: 1
          range: 1m
          perks: none
          validation: n/a
```

## Phase

Phases are a state or passive ability that lasts for a predefined amount of time. They exist to buff or nerf someone for a specific number of rounds. A Phase lasts for a few rounds, after which the Reverse Phase starts and lasts just as long as the original phase did.

### Rules

*   Costs 3 Energy to Use.
*   After activation, Phase is active for 2 rounds.
*   Phase ends at the start of the 2nd turn of the character.
*   When Phase ends, the Reverse Phase starts.
*   During the Reverse Phase, no new Phases can be started for the character.
*   Phase will have an Enactment assigned to it.
*   The Enactment can be triggered as a free action at the end of the character’s turn.
*   Reverse Phase will have a Bad Enactment assigned to it.
*   Bad Enactment will be applied to the character.
*   Bad Enactment must be used at the end of the character’s turn as a free action.
*   If no Bad Enactment is chosen, the Bad Enactment will be the reverse of the original Enactment (e.g., a dice increase will turn into a dice decrease).
*   Phase has a knockout requirement.
*   If any knockout requirement is met, the Phase ends (and the “Bad Enactment” starts).
*   The Reverse Phase cannot be cancelled by the knockout.

### Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Add another round to the Phase and the Reverse Phase | +1 | 2 |
| Remove one round from the Reverse Phase | 0 | 4 |
| All knockout requirements have to be met instead | 0 | 3 |
| Knockout can be used on the Reverse Phase | 0 | 3 |
| Reduce Energy cost by 1 (up to a minimum of one) | \-1 | 3 |
| Increase Energy cost by 1 | +1 | \-2 |
| Has Item Dependency | 0 | \-1 |

### Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Adjustment (free perk: Bonus stays for the amount of rounds the Phase lasts) | 0 | 0 |
| Enact Damage | +1 | 2 |
| Enact Healing | +1 | 2 |

### Compatible Bad Enactments

_(Note: Adding worse reverse enactments can refund points)_

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Adjustment (free perk: Penalty stays for the amount of rounds the Reverse Phase lasts) | 0 | 0 |
| Enact Damage | 0 | \-2 |
| Enact Healing (to enemies) | 0 | \-3 |

### Knockout Requirement

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| You take damage | 0 | 0 |
| You fall unconscious / die | 0 | 0 |
| You get grabbed/restrained | 0 | 0 |

## Preparation

Just like a Reaction a preparation works outside the regular turn order. They follow the exact same rules as a Reaction but instead of being passively on the background, a preparation must will cost an action to, well, prepare. 

### Rules

*   Can only be used once per round.
*   Costs 2 actions
*   Costs 3 Energy to Use.
*   Always has at least one Trigger (Pick one from the list below, first one is free).
*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within 1m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

### Perks

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Add 1 meter to reaction range | 0 | 1 |
| Add one more use per round | +1 | 4 |
| Reduce Energy cost by 1 (up to a minimum of one) | \-1 | 3 |
| Increase Energy cost by 1 | +1 | \-2 |
| Has Item Dependency | 0 | \-1 |
| Reduce action cost by 1 | 4 | 1 |
| Increase action cost by 1 | \-2 | 1 |

### Triggers

_(Note: First trigger is free. Below are the costs to add additional alternative triggers to the same ability)_

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Someone runs away from you | 0 | 2 |
| Someone runs towards you | 0 | 2 |
| Someone has to do a Defense roll | 0 | 2 |
| Someone takes damage | 0 | 2 |
| Someone gets healed | 0 | 2 |
| Someone gets an adjustment (Buff/Penalty) | 0 | 2 |
| Someone does a skill check | 0 | 2 |
| Someone summons a minion | 0 | 2 |
| Someone grabs your character | 0 | 2 |
| You walk towards someone | 0 | 2 |
| You walk away from someone | 0 | 2 |

### Compatible Enactments

| Description | Energy Cost | Add Cost |
| --- | --- | --- |
| Enact Adjustment | 1 | 2 |
| Enact Persistent Effect | 2 | 3 |
| Enact Damage | 1 | 2 |
| Enact Healing | 1 | 2 |
| Enact Movement | 0 | 1 |
| Enact other execution | 2 | 4 |

  
 

## Minion (W.I.P.)

Minions are entities that players can create, summon, and control. They have default stats and actions, and can use Enactments created by the user if the appropriate perk is selected. Minions follow specific rules within the action economy.

### Rules

*   Minions have their own turn in the action economy.
*   Minions can perform one action per turn.
*   Minions have default stats: Health, Attack, Defense, Speed, Lifetime.
*   Minions can be summoned once per encounter.
*   Minions require a summoning cost (e.g., energy, mana).
*   Minions can be controlled by the player during their turn.
*   Minions can be dismissed by the player as a free action.
*   Minions have a default lifetime of 3 rounds.

### Default Stats

|  |  |
| --- | --- |
| Stat | Value |
| Health | 10 |
| Attack | 2d6 |
| Defense | 1d6 |
| Speed | 5m |
| Lifetime | 3 rounds |

### Actions

|  |  |  |
| --- | --- | --- |
| Description | Energy Cost | Add Cost |
| Basic Attack | 0 | 0 |
| Defend | 0 | 0 |
| Move | 0 | 0 |

### Perks

|  |  |  |
| --- | --- | --- |
| Description | Energy Cost | Add Cost |
| Increase Health by 5 | 0 | 1 |
| Increase Attack by 1d6 | +1 | 2 |
| Increase Defense by 1d6 | 0 | 2 |
| Increase Speed by 2m | 0 | 1 |
| Add an additional action per turn | +2 | 5 |
| Minion can use an additional Ability | +1 | 4 |
| Increase Lifetime by 1 round | +1 | 1 |
| Minion has item dependency | 0 | \-1 |
| Reduce Energy cost by 1 (minimum of one) | \-1 | 3 |
| Increase Energy cost by 1 | +1 | \-2 |

### Compatible Enactments

Minions can use Enactments created by the user if the "Minion can use an additional Enactment" perk is selected. These Enactments follow the same rules as player Enactments but are executed by the minion.

|  |  |  |
| --- | --- | --- |
| Description | Energy Cost | Add Cost |
| Enact Adjustment | 1 | 2 |
| Enact Persistent Effect | 2 | 3 |
| Enact Damage | 1 | 2 |
| Enact Healing | 1 | 2 |
| Enact Movement | 0 | 1 |