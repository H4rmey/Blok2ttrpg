# ranged
## Ranged

**Ranged** **Interactions** include actions like using bows, guns, and boomerangs. These interactions offer an increased range compared to **Direct** Interactions but come with a lower success rate due to a penalty on the **Engagement Roll**. Additionally, the target must not be obstructed or invisible to the **Engager** by default.

## Rules

*   Has a Validation
*   Target is a single character.
*   Target must be within {{.Ranged.DefaultRange}}m of your character.
*   Target must be visible.
*   Target must not be obstructed.
*   Engagement roll result is lowered by 2

## Perks

{{.RangedPerksTable}}

## Template

```yaml
interactions:
  - type: Ranged
    engager: Self
    target_amount: {{.Ranged.DefaultTargets}}
    range: {{.Ranged.DefaultRange}}m # Default range for Ranged interactions
    visibility: Visible # Target must be visible
    obstruction: Not obstructed # Target must not be obstructed
    perks:
      - description: <insert description of perk here>
        add_cost: <cost of the perk>
        amount: <amount of times the perk is chosen>
        total_add_cost: <total add cost>
        energy_cost: <energy cost to use>
        is_optional: <True/False>
    validation:
      engagement_roll: <pick an Offensive Trait> - 2
      counter_roll: <pick two Defensive Traits>
```