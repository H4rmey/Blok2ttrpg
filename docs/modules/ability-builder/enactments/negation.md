# negation
## Enact Negation

Enact Negation allows characters to reduce incomming damage by choosing a source die. This can also be used to remove a Persistant Effect.

## Rules

*   **Negate Dice**: Choose a source die when building this enactment.
*   **Remove Persistant Effect**: Make the Negation roll, the damage of the persistant effect is reduced by this value. If the Negate roll is higher or equal then the Persistant Effect is removed.

## Perks

{{perksTable (enactment "negation")}}


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
