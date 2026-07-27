# passive
## Passive

Passives are Abilities that are always on. They work just like a Reaction, they trigger when something happens, but unlike a Reaction a Passive does not cost any Energy or Actions to use and is not bound to your action economy at all. Whenever the trigger happens, the linked Enactment is executed. For example, you could have a passive that triggers whenever someone damages you, Enacting a small healing effect on yourself.

Because a Passive is free to use and can trigger whenever, it is the most expensive Ability Type to build. This higher base build cost is the price you pay for never having to spend Energy or Actions on it.

## Rules

*   Does not cost an Action.
*   Does not cost any Energy to use.
*   Has a higher base build cost of {{(abilityType "passive").BaseCost.BuildCost}} build points.

*   Always has at least one Trigger. Each trigger has its own build cost (see the Perks table below); more powerful triggers cost more.

*   Has at least one Enactment (the first Enactment is free)
*   Only triggers when the triggering effect happens within {{(abilityType "passive").BaseRange}}m of you.

*   Target of Enactments is overwritten to the character that triggers the Passive.

## Perks

{{perksTable (abilityType "passive")}}

## Template

```yaml
ability:
  type: Passive
  range: {{(abilityType "passive").BaseRange}}
  uses: {{(abilityType "passive").BaseUses}}
  has_item_dependency: No # If yes, enter which item

  trigger: <trigger name here>
  enactments:
    - Type:
  perks:
```
