# preparation
## Preparation

Just like a Reaction, a Preparation works outside the regular turn order. They follow the exact same rules as a Reaction but instead of being passively on the background, a Preparation will cost an action to prepare, but in turn cost far less to use.

## Rules

*   Can only be used once per round.
*   Costs {{.Preparation.BaseAction}} actions.
*   Costs {{.Preparation.BaseEnergy}} Energy to Use.
*   Always has at least one Trigger (Pick one from the list below, first one is free).
*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within {{.Preparation.BaseRange}}m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

## Perks

{{.PreparationPerksTable}}

## Triggers

{{.PreparationTriggersTable}}

## Compatible Enactments

{{.PreparationEnactmentsTable}}

## Template

```yaml
ability:
  type: Preparation
  range: {{.Preparation.BaseRange}}
  uses: {{.Preparation.BaseUses}}
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{.Preparation.BaseEnergy}}
  action_cost: {{.Preparation.BaseAction}}
  trigger: <trigger name here>
  enactments:
    - Type:
  Perks:
```
