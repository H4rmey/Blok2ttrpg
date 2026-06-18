# Validations

# Validations

## Introduction

Here, you'll find the guidelines and options for customizing your engagement and counter rolls, along with various perks to enhance your gameplay experience.

## Rules

- **Engagement Roll**: This is an Offensive Trait used to initiate actions against a target.
- **Counter Roll**: This involves two Defensive Traits, allowing the target to choose how they respond to the attack.

## Perks

| Description | yaml value to update | Energy Cost | Add Cost |
| --- | --- | --- | --- |
| Engage Roll is replaced by a Generic Roll (default 1d6) | engage_roll | 0   | \-2 |
| Counter Roll is replaced by a Generic Roll (default 1d12) | counter_roll | 0   | \-2 |
| Replace one of the Counter Rolls to any other Trait | counter_roll | 0   | 2   |
| Remove one of the Counter Roll options | counter_roll | +1  | 3   |
| Replace the Engagement Roll to any other Trait | engage_roll | 0   | 3   |
| Add another Engagement Roll option (you cannot roll both, but choose before using the ability) | engage_roll | 0   | 2   |
| Shift Dice Tier of Generic Counter Roll UP | counter_roll | 0   | \-2 |
| Shift Dice Tier of Generic Counter Roll DOWN | counter_roll | +1  | 3   |
| Shift Dice Tier of Generic Engagement Roll UP | engage_roll | +1  | 3   |
| Shift Dice Tier of Generic Engagement Roll DOWN | engage_roll | 0   | \-2 |
| Use the result of another roll for this entry | \*  | +1  | 3   |

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

## Example

```yaml
# Quick example of some perks chosen.
validation:
  engagement_roll: d10 # dice is replaced by a generic d6 and then upgraded twice
  counter_roll: Reflex or Medicine # normally I would select something like Reflex and Constitution, but due to my perks I replace Constitution with a Medicine check.
  perks:
    - description: Engage Roll is replaced by a Generic Roll (default 1d6)
      add_cost: -2
      amount: 1
      total_add_cost: -2
      energy_cost: 0
      is_optional: False
    - description: Shift Dice Tier of Generic Engagement Roll UP
      add_cost: 3
      amount: 2
      total_add_cost: 6
      energy_cost: 2 # 1 extra energy per tier shift up
      is_optional: True # Player can choose to roll lower to save 2 Energy
    - description: Replace one of the Counter Rolls to any other Trait
      add_cost: 2
      amount: 1
      total_add_cost: 2
      energy_cost: 0
      is_optional: False
```