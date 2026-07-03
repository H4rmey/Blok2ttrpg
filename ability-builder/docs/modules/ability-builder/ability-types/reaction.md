# Reaction

Reactions are Abilities that trigger outside your normal action economy. Reactions trigger when someone else (or you) does something. When the trigger happens, the linked Enactment is executed.

## Rules

*   Can only be used once per round.
*   Does not cost an action.
*   Costs {{.Reaction.BaseEnergy}} Energy to Use.
*   Always has at least one Trigger (Pick one from the list below, first one is free).
*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within {{.Reaction.BaseRange}}m of you.
*   Target of Enactments is overwritten to the character that triggers the Reaction.

## Perks

{{.ReactionPerksTable}}

## Triggers

{{.ReactionTriggersTable}}

## Compatible Enactments

{{.ReactionEnactmentsTable}}

## Template

```yaml
ability:
  type: Reaction
  range: {{.Reaction.BaseRange}}
  uses: {{.Reaction.BaseUses}}
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{.Reaction.BaseEnergy}}
  trigger: <trigger name here>
  enactments:
    - Type:
  Perks:
```
