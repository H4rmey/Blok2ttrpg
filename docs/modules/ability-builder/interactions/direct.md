# direct
## Direct

## Rules

*   Has a Validation.
*   Target is a single character.
*   Target must be within {{fieldDefault (interaction "direct") "range"}}m of your character.

## Perks

{{perksTable (interaction "direct")}}

## Template

```yaml
interactions:
  - type: Direct
    engager: Self
    target_amount: {{fieldDefault (interaction "direct") "targets"}}
    range: {{fieldDefault (interaction "direct") "range"}}m

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