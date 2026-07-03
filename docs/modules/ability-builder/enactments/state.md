# Enact State

Enact State abilities apply a state or condition to a target (e.g., stunned, poisoned). This enactment type is currently a work in progress.

## Rules

*   **State**: Applies a condition to the target.

## Perks

{{.StatePerksTable}}

## Template

```yaml
enactments:
  - type: Enact State
    state: <state here>
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
```
