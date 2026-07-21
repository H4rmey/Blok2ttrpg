# self
## Self

## Rules

*   Has a Validation
*   Engager is yourself
*   Target is yourself
*   Counter Roll is replaced by a Generic Dice when the relevant perk is chosen.

## Perks

{{perksTable (interaction "self")}}


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