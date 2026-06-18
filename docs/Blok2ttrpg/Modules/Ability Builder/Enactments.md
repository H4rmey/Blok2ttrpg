# Enactments

# Enactments

## Introduction

Enactments are the specific actions an Ability takes. Each Enactment serves a distinct purpose and follows specific rules. Enactments can be combined to create complex and powerful Abilities. The main types of Enactments include:

- **Enact Proficiency Shift**: Modifies an attribute or stat of the target (e.g., reducing perception, increasing agility).
- **Enact Persistent Effect**: Applies a lingering effect that remains for multiple turns (e.g., fire damage over time, regenerative healing).
- **Enact Damage**: Inflicts damage to a target (e.g., dealing 1d8 slashing damage).
- **Enact Healing**: Restores health to a target (e.g., healing for 1d6 HP).
- **Enact Movement**: Moves a target or the Ability user (e.g., pushing an enemy backward, teleporting the user).
- **Enact State (WIP)**: Applies a state or condition to a target (e.g., stunned, poisoned).

Each Enactment type has specific rules and perks that can be used to enhance their effects or introduce unique mechanics, providing flexibility and strategic depth in gameplay.

## Enact Damage

Enact Damage abilities allow characters to inflict harm on their enemies.

### Rules

- **Damage Dice**: The default damage dice is 1d4.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Shift Dice Tier of damage up | damage_dice | +1  | 2   |
| Change Damage Dice to one of your traits | damage_dice | 0   | 3   |
| Add a flat +1 bonus to the result | damage_dice | 0   | 2   |
| Add an Offensive Trait Dice to the Damage Dice | damage_dice | +2  | 4   |
| Will Always Resolve | n/a | +3  | 5   |
| Use the result of another roll for this entry | \*  | +1  | 3   |

### Template

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
        is_optional: <True/False> # Can the player drop this perk to save energy?
    interactions:
      - type:
          validation:
```

### Example

```yaml
enactments:
  - type: Enact Damage
    damage_dice: d8 + 2
    is_optional: False 
    base_enactment_energy_cost: 0 
    perks:
      - description: Upgrade Dice Tier of roll 
        add_cost: 2 
        amount: 2 
        total_add_cost: 4 
        energy_cost: 2 # Costs 1 energy per tier upgraded
        is_optional: True # Player can choose to just deal 1d4 to save 2 energy
      - description: Add a flat +1 bonus to the result 
        add_cost: 2 
        amount: 2 
        total_add_cost: 4 
        energy_cost: 0 # This perk doesn't cost extra energy to use
        is_optional: False
    interactions:
      - type: Ranged
          engager: Self
          target_amount: 1
          range: 10m + 4m = 14m
          visibility: Visible
          obstruction: Not obstructed
          perks:
          - description: Increase reach with 2m
            add_cost: 1
            amount: 2
            total_add_cost: 2
            energy_cost: 0
            is_optional: False
          validation:
            engagement_roll: Precision - 2
            counter_roll: 
              - Magic
              - Reflex
```

## Enact Healing

Enact Healing abilities allow characters to restore health to themselves or their allies. These abilities can be used to mend wounds, cure ailments, and provide vital support during combat.

### Rules

- **Healing Dice**: The default healing dice is 1d4.
- **Interaction Type**: If the interaction type is Self or Direct, no validation is required.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Shift Dice Tier of Heal up | healing_dice | +1  | 2   |
| Change Heal Dice to one of your traits | healing_dice | 0   | 3   |
| Add a flat +1 bonus to the result | healing_dice | 0   | 2   |
| Add Medicine Trait Dice to the heal effect (can only be used once) | healing_dice | +1  | 3   |
| Will Always Resolve | n/a | +2  | 4   |
| Use the result of another roll for this entry | \*  | +1  | 3   |

### Template

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

### Example

```yaml
enactments:
  - type: Enact Healing
    healing_dice: 1d4 + Medicine
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: Add Medicine trait dice to the heal effect 
        add_cost: 3 
        amount: 1 
        total_add_cost: 3 
        energy_cost: 1
        is_optional: True # Player can drop the medicine trait to save 1 energy
    interactions:
      - type: Direct
        engager: Self
        target_amount: 3
        range: 1m
        perks:
          - description: Add another target
            add_cost: 3
            amount: 2
            total_add_cost: 6
            energy_cost: 4 # Healing 2 extra targets costs heavy energy (2 per target)
            is_optional: True # Can heal fewer people to save energy
         validation: n/a
```

## Enact Movement

Enact Movement abilities allow characters to manipulate the position of themselves or their targets. These abilities can be used to push enemies away, pull allies closer, or reposition oneself strategically. Movement abilities add a dynamic element to gameplay, enabling tactical maneuvers and creative solutions to challenges. Additionally, the origin of the movement can be assigned to an object or another person, allowing for even more creative and strategic uses. For example, you could attach the origin to an arrow or a device, and then use a Ranged Interaction to throw it and pull the target towards it.

### Rules

- **Direction**: The target will move in one direction relative to an origin. Possible directions include Up, Down, Away, Towards, Forward, Left, or Right.
- **Distance**: The target will move 1 meter by default.
- **Origin**: The default origin is the engager or item/location from previous enactment.
- **Obstacle**: If the target moves into an obstacle, they take 1d4 damage.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Add 1m to the movement | minimal_distance | 0   | 1   |
| Add another option for the direction (when using ability, pick one) | direction_options | 0   | 1   |
| Change Origin to something else | origin | +1  | 2   |
| Change the total movement to any other trait | minimal_distance | 0   | 3   |
| Will Always Resolve | n/a | +1  | 3   |
| Use the result of another roll for this entry | \*  | +1  | 3   |

**Template**

```yaml
enactments:
  - type: Enact Movement
    minimal_distance: 1m # Distance moved (default is 1 meter)
    origin: engager
    direction_options: 
      - <Direction> # Direction of movement (Up, Down, Away, Towards, Left, or Right)
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

**Example**

```yaml
enactments:
  - type: Enact Movement
    origin: Engager 
    minimal_distance: 4m 
    direction_options: 
      - Up
      - Away
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: Add 1m to the movement 
        add_cost: 1 
        amount: 3 
        total_add_cost: 3 
        energy_cost: 0 # Moving further doesn't cost extra energy here
        is_optional: False
    interactions:
      - type: Area 
          engager: Self 
          radius: 1m 
          range: 0m 
          origin: Engager 
          validation:
            engagement_roll: Power 
            counter_roll:
              - Reflex 
              - Constitution
```

## Enact Proficiency Shift

Enact Proficiency Shift abilities allow characters to temporarily enhance or weaken Traits. These abilities can be used to boost a character's Proficiency in a specific area or to hinder an opponent's effectiveness. Proficiency shifts add a strategic layer to gameplay, enabling players to adapt to different situations by modifying their strengths and weaknesses.

### Rules

- **Shift Direction**: Shift a Proficiency Tier from a Trait either up or down. (yaml: `shift_direction`)
- **Single Use**: The shift only has one use.
- **Trait Check**: The next time a Trait check is made with the shifted Trait, you must use the shifted Proficiency.
- **Reset**: Using the shifted Proficiency resets the Trait back to its original Proficiency.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Add another use to the Shifted Trait | shift_uses | +1  | 3   |
| Shift the same Trait a second time | shift_amount | +1  | 3   |
| You may choose if you use the Shifted Proficiency or not. | n/a | 0   | 2   |
| Will Always Resolve | n/a | +2  | 4   |
| Use the result of another roll for this entry | \*  | +1  | 3   |

### Template

```yaml
enactments:
  - type: Enact Proficiency Shift
    shifted_trait: <trait here> # the trait you wish to shift.
    shift_direction: <UP DOWN or> # shift the trait up or down.
    shift_amount: 1 # the amount of Tiers the Trait is Shifted.
    shift_uses: 1 # the amount of times a the shifted perk stays until it is reset.
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

### Example

```yaml
# Lower perception of one Target
enactments:
  - type: Enact Proficiency Shift
    shifted_trait: Perception
    shift_direction: DOWN
    shift_amount: 2 # default of 1
    shift_uses: 2
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: Shift the same Trait a second time 
        add_cost: 3 
        amount: 1 
        total_add_cost: 3 
        energy_cost: 1
        is_optional: True # Can choose to only shift it down by 1 to save energy
      - description: Add another use to the shifted Trait 
        add_cost: 3 
        amount: 1 
        total_add_cost: 3 
        energy_cost: 1
        is_optional: True # Can choose to only have 1 use to save energy
    interactions:
      - type: Self
        engager: Self
        target_amount: 1
        range: 10m + 4m = 14m
        visibility: Visible
        obstruction: Not obstructed
        perks:
          - description: Increase reach with 2m
            add_cost: 1
            amount: 2
            total_add_cost: 2
            energy_cost: 0
            is_optional: False
        validation:
          engagement_roll: Mind - 2 
          counter_roll: Mind
```

## Enact Persistent Enactment

The Enact Persistent Effect applies a lingering effect to a target, such as fire, frost, or poison damage. By default, the effect lasts for 2 rounds and triggers at either the start of the target’s turn or the end of the engager's turn. Targets can spend one action to attempt to remove the effect using a provided solution, such as a Dexterity save.

### Rules

- **Duration**: Lasts 2 rounds by default.
- **Trigger Timing**: The effect triggers at either: (pick one)
- The start of the target's turn (before the target can use an action).
- The end of the engager's turn on each infected target.
- **Solutions**: Targets can spend one action to attempt to remove the effect using the provided solution. There must be two solutions, which can be any Trait Roll (e.g., Dexterity save).
- **Applies a Single Enactment**: The persistent effect applies a single enactment (e.g., Enact Damage, Enact Healing).

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Add another round to the Effect | duration | +1  | 2   |
| Remove one Solution option | solutions | +1  | 3   |
| Add another Effect | effects | +2  | 4   |
| Will Always Resolve | n/a | +3  | 5   |
| Use the result of another roll for this entry | \*  | +1  | 3   |

### Effects

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Enact Damage | effects | +1  | 2   |
| Enact Healing | effects | +1  | 2   |
| Enact Movement | effects | 0   | 1   |
| Enact Proficiency Shift | effects | +1  | 2   |

### Template

```yaml
enactment:
  - type: Enact Persistent Effect
    duration: 3 rounds # Duration of the effect (default is 2 rounds)
    trigger_timing: Start of Target's Turn or The end of the engager's turn.
    solutions:
      - Dexterity    # First solution to remove the effect
      - Constitution # Second solution to remove the effect
    is_optional: True # Persistent effects are often secondary and can be skipped
    base_enactment_energy_cost: 2
    effects: 
      # work out the enactments here
      - type: <Enactment here type> # Type of persistent effect (e.g., Fire, Frost, Poison)
    interactions:
      - type:
          validation:
    perks: Enact Persistent Effect
      - description: <Perk Description> 
        add_cost: <Cost> 
        amount: <Amount> 
        total_add_cost: <Total Cost> 
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
```

### Example

```yaml
enactment:
  - type: Enact Persistent Effect
    duration: 4 rounds # Duration of the effect (default is 2 rounds)
    trigger_timing: Start of Target's Turn
    solutions:
      - Dexterity    # First solution to remove the effect
      - Constitution # Second solution to remove the effect
    is_optional: True
    base_enactment_energy_cost: 2
    effects: 
      - type: Enact Damage # Type of persistent effect (e.g., Fire, Frost, Poison)
        damage: 1d4 + 1
        is_optional: False # Required if the persistent effect itself triggers
        base_enactment_energy_cost: 0
        perks: Enact Damage
          - description: Add a flat +1 bonus to the result
            add_cost: 2
            amount: 1
            total_add_cost: 2
            energy_cost: 0
            is_optional: False
    perks: Enact Persistent Effect
      - description: Add another round to the Effect
          add_cost: 2
          amount: 2
          total_add_cost: 4
          energy_cost: 2 # Adding 2 extra rounds costs 2 energy
          is_optional: True # Can choose to only let it burn 2 rounds to save energy
    interactions: 
      - type: Ranged
        engager: Self
        target_amount: 2
        range: 10m # Default range for Ranged interactions
        visibility: Target does not have to be visible (see perk)
        obstruction: Target must not be obstructed 
        perks: Ranged
          - description: Engagement does not have to be visible
            add_cost: 6 # (Value from original block, adjust as needed)
            amount: 1
            total_add_cost: 6
            energy_cost: 2
            is_optional: True # Can require line of sight to save energy
          - description: Add another character
            add_cost: 3
            amount: 1
            total_add_cost: 3
            energy_cost: 2
            is_optional: True # Can affect only 1 character to save energy
        validation:
            engagement_roll: Precision - 2
            counter_roll: 
              - Dexterity 
              - Constitution
```