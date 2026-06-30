# Interactions
## Interactions

## Introduction

Interactions determine how an Enactment is applied within the game world. They define the targets, range, and conditions under which an Enactment takes effect. Interactions provide the framework for how Abilities impact the game environment and characters, ensuring that Enactments are applied accurately and effectively.

There are different types of interactions. Which all have their own Rules:

*   Self
*   Direct
*   Ranged
*   Area
*   Area of effect (AoE)

## Self

### Rules

*   Has a Validation
*   Engager is yourself
*   Target is yourself
*   Counter Roll is a d8

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Validation counter roll is replaced by a Generic Dice: d8 | counter\_roll | 0 | 2 |
| Will always resolve | n/a | +3 | 5 |
| Use the result of another roll for this entry | \* | +1 | 3 |

### Template

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

## Direct

### Rules

*   Has a Validation.
*   Target is a single character.
*   Target must be within 1m of your character.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Increase range with 1m | range | 0 | 1 |
| Add another target | target\_amount | +2 | 3 |
| Will always resolve | n/a | +3 | 5 |
| Use the result of another roll for this entry | \* | +1 | 3 |

### Template

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

## Ranged

Ranged Interactions include actions like using bows, guns, and boomerangs. These interactions offer an increased range compared to Direct Interactions but come with a lower success rate due to a penalty on the Engagement Roll. Additionally, the target must not be obstructed or invisible to the Engager by default.

### Rules

*   Has a Validation
*   Target is a single character.
*   Target must be within 10m of your character.
*   Target must be visible.
*   Target must not be obstructed.
*   Engagement roll result is lowered by 2

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Increase range with 2m | range | 0 | 1 |
| Decrease range with 2m | range | 0 | \-1 |
| Add another character | target\_amount | +2 | 3 |
| Engagement does not have to be visible | target\_may\_be\_visible | +1 | 3 |
| Engagement may be obstructed | target\_may\_be\_obstructed | +1 | 3 |
| Remove the Engagement Roll Penalty | engagement\_roll | +1 | 3 |
| Will always resolve | n/a | +3 | 5 |
| Use the result of another roll for this entry | \* | +1 | 3 |

### Template

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

### Example

```yaml
interactions:
  - type: Ranged
    engager: Self
    target_amount: 1
    range: 10m + 4m =  14m
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

## Area

Area Interactions encompass actions like bombs, splash potions, and traps. These interactions always have a defined **Radius** and **Range**:

*   **Radius**: This determines the area where the Enactment will take effect.
*   **Range**: This specifies how far from the user the point of origin is set. By default, the point of origin is 0m from the user.

You can also assign the point of origin to an object, but this must be discussed with the GM beforehand. So you could put the point of Origin to an arrow or a device you’ve made. Then use a Ranged Interaction to throw it.

### Rules

*   **Validation**: The interaction must have a validation.
*   **Radius**: The default radius is 1 meter.
*   **Range**: The default range is 0 meters.
*   **Origin**: The point of origin for the radius is the engager or item/location from previous enactment.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Increase radius with 1m | radius | +1 | 2 |
| Increase Range by 2m | range | 0 | 1 |
| Change Origin to something else (Discuss with GM) | origin | +1 | 2 |
| Will always resolve | n/a | +3 | 5 |
| Use the result of another roll for this entry | \* | +1 | 3 |

### Template

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

### Example

```yaml
interactions:
  - type: Area
    radius: 2m # Increased radius by 1m using a perk
    range: 4m # Increased range by 2m using a perk
    origin: Engager
    perks:
      - description: Increase radius with 1m
        add_cost: 2
        amount: 1
        total_add_cost: 2
        energy_cost: 1
        is_optional: True # Player can keep radius to 1m to save 1 Energy
      - description: Increase Range by 2m
        add_cost: 1
        amount: 2
        total_add_cost: 2
        energy_cost: 0
        is_optional: False
    validation:
      engagement_roll: Strength
      counter_roll: 
        - Dexterity 
        - Constitution
```

## Area of Effect

An Area of Effect (AoE) Interaction functions similarly to an Area Interaction, but its effects persist for several rounds. While an Area Interaction might be like a single-use bomb, an AoE Interaction is akin to a bomb that detonates every round. Alternatively, it could represent a healing circle, where characters gain health each round they remain within the AoE. The possibilities are endless, so get creative!

The effect of the AoE does not trigger immediately. Instead, it activates either at the start of a character's turn within the AoE or at the end of the Engager's turn.

### Rules

*   **Validation**: The interaction must have a validation.
*   **Radius**: The default radius is 1 meter.
*   **Range**: The default range is 0 meters.
*   **Origin**: The point of origin for the radius is the engager.
*   **Trigger Conditions**:
    *   The Enactm  ent will be triggered when a character enters the Area of Effect.
    *   The Enactment will be triggered at either (choose one):
        *   The start of a character's turn while in the Area of Effect.
        *   The end of the engager’s turn on each character that is in the Area of Effect.
*   **Duration**: The effect lasts for 2 rounds.

### Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Increase radius with 1m | radius | +1 | 2 |
| Increase Range by 2m | range | 0 | 1 |
| Change Origin to something else (Discuss with GM) | origin | +1 | 2 |
| Increase the amount of rounds by 1 | duration | +1 | 2 |
| Engager is immune to the effect | immunity | 0 | 2 |
| Will always resolve | n/a | +3 | 5 |
| Use the result of another roll for this entry | \* | +1 | 3 |

### Template

```yaml
interactions: 
  - type: Area of Effect
    radius: 1m # Default radius for AoE interactions
    range: 0m # Default range for AoE interactions
    origin: Engager # Point of origin is the Engager
    duration: 2 # Default duration for AoE interactions
    immunity: false
    trigger_conditions:
      - Entering the Area of Effect # Triggered when a character enters the area
      - Start of character's turn within the Area of Effect # Triggered at the start of a character's turn within the area
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

### Example

```yaml
interactions:
  - type: Area of Effect
    radius: 2m # Increased radius by 1m using a perk
    range: 2m # Increased range by 2m using a perk
    origin: Engager
    duration: 3 rounds # Increased duration by 1 round using a perk
    trigger_conditions:
      - Entering the Area of Effect
      - Start of character's turn within the Area of Effect
    perks:
      - description: Increase radius with 1m
        add_cost: 2
        amount: 1
        total_add_cost: 2
        energy_cost: 1
        is_optional: True
      - description: Increase Range by 2m
        add_cost: 1
        amount: 1
        total_add_cost: 1
        energy_cost: 0
        is_optional: False
      - description: Increase the amount of rounds by 1
        add_cost: 2
        amount: 1
        total_add_cost: 2
        energy_cost: 1
        is_optional: True # Engager can end the AoE a round early to save 1 Energy
    validation:
      engagement_roll: Intelligence
      counter_roll: 
        - Wisdom
        - Constitution
```