# minion
## Minion

Minions are entities that players can create, summon, and control. They have default stats and actions, and can use Enactments created by the user if the appropriate perk is selected. Minions follow specific rules within the action economy.

## Rules

*   Minions have their own turn in the action economy.
*   Minions can perform one action per turn.
*   Minions have default stats: Health, Attack, Defense, Speed, Lifetime.
*   Minions can be summoned once per encounter.
*   Minions require a summoning cost (e.g., energy, mana).
*   Minions can be controlled by the player during their turn.
*   Minions can be dismissed by the player as a free action.
*   Minions have a default lifetime of {{.Minion.BaseLifetime}} rounds.

## Default Stats

| Stat | Value |
| --- | --- |
| Health | {{.Minion.BaseHealth}} |
| Attack | 2d6 |
| Defense | 1d6 |
| Speed | 5m |
| Lifetime | {{.Minion.BaseLifetime}} rounds |

## Perks

{{.MinionPerksTable}}

## Compatible Enactments

{{.MinionEnactmentsTable}}