# damage
## Enact Damage

Enact Damage allows characters to inflict harm on their enemies.

## Rules

*   **Damage Dice**: Choose a source die when building this enactment.

## Perks

{{perksTable (enactment "damage")}}


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