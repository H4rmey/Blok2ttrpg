# execution
## Execution

Execution is the most basic form for an Ability. It is simply the: "I want to do this now" Ability Type. Executions can be anything from casting a fireball to summoning a shield to block an attack or preparing a parry.

## Rules

*   **Enactments**: Has at least one Enactment (the first Enactment is free)
*   **Actions**: Costs {{(abilityType "execution").BaseAction}} Actions to use
*   **Energy**: Costs {{(abilityType "execution").BaseEnergy}} Energy to use


## Perks

{{perksTable (abilityType "execution")}}

## Template

```yaml
ability:
  type: Execution
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{(abilityType "execution").BaseEnergy}}
  action_cost: {{(abilityType "execution").BaseAction}}

  enactments:
    - Type:
  perks:
```
