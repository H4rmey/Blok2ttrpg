# Enact Healing

Enact Healing abilities allow characters to restore health to themselves or their allies.

## Rules

*   **Healing Dice**: The default healing dice is {{.Healing.DefaultDice}}.
*   **Interaction Type**: If the interaction type is Self or Direct, no validation is required.

## Perks

{{.HealingPerksTable}}

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
