# preparation
## Preparation

Just like a Reaction, a Preparation works outside the regular turn order. They follow the exact same rules as a Reaction but instead of being passively on the background, a Preparation will cost an action to prepare, but in turn cost far less to use.

## Rules

*   Can only be used once per round.
*   Costs {{(abilityType "preparation").BaseAction}} actions.
*   Costs {{(abilityType "preparation").BaseEnergy}} Energy to Use.
*   Always has at least one Trigger. Each trigger has its own build cost (see the Perks table below); more powerful triggers cost more.

*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within {{(abilityType "preparation").BaseRange}}m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

## Perks

{{perksTable (abilityType "preparation")}}

## Template

```yaml
ability:
  type: Preparation
  range: {{(abilityType "preparation").BaseRange}}
  uses: {{(abilityType "preparation").BaseUses}}
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{(abilityType "preparation").BaseEnergy}}
  action_cost: {{(abilityType "preparation").BaseAction}}
  trigger: <trigger name here>
  enactments:
    - Type:
  Perks:
```
