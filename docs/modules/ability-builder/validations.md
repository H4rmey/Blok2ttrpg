# validations
## Validations

## Introduction

Here, you'll find the guidelines and options for customizing your **Engagement and Counter Rolls**.

## Rules

*   **Engagement Roll**: This is an Offensive Trait used to initiate actions against a target.
*   **Counter Roll**: This involves two Defensive Traits, allowing the target to choose how they respond to the attack.

## Options

{{perksFields .Validations.Fields}}


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