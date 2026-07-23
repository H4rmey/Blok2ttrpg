# reaction
## Reaction

Reactions are Abilities that trigger outside your normal action economy. Reactions trigger when someone else does something. When the trigger happens, the linked Enactment is executed. For example, you could have a reaction that triggers whenever someone runs towards you, Enacting a healing effect on yourself.

## Rules

*   Can only be used once per round.
*   Does not cost an action.
*   Costs {{(abilityType "reaction").BaseEnergy}} Energy to Use.

*   Always has at least one Trigger. Each trigger has its own build cost (see the Perks table below); more powerful triggers cost more.

*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within {{(abilityType "reaction").BaseRange}}m of you.

*   Target of Enactments is overwritten to the character that triggers the Reaction.

## Perks

{{perksTable (abilityType "reaction")}}

## Template

```yaml
ability:
  type: Reaction
  range: {{(abilityType "reaction").BaseRange}}
  uses: {{(abilityType "reaction").BaseUses}}
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{(abilityType "reaction").BaseEnergy}}

  trigger: <trigger name here>
  enactments:
    - Type:
  Perks:
```
