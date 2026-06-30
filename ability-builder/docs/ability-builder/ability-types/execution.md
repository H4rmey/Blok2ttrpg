# Execution

Execution is the most basic form for an Ability. It is simply the: "I want to do this now" Ability Type. Executions can be anything from casting a fireball to summoning a shield to block an attack or preparing a parry.

## Rules

*   **Enactments**: Has at least one Enactment (the first Enactment is free)
*   **Actions**: Costs {{.Execution.BaseAction}} Actions to use
*   **Energy**: Costs {{.Execution.BaseEnergy}} Energy to use

## Perks

{{.ExecutionPerksTable}}

## Compatible Enactments

{{.ExecutionEnactmentsTable}}

## Template

```yaml
ability:
  type: Execution
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{.Execution.BaseEnergy}}
  action_cost: {{.Execution.BaseAction}}
  enactments:
    - Type:
  perks:
```
