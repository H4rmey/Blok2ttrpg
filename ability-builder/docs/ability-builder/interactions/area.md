# Area

Area Interactions encompass actions like bombs, splash potions, and traps. These interactions always have a defined **Radius** and **Range**:

*   **Radius**: This determines the area where the Enactment will take effect.
*   **Range**: This specifies how far from the user the point of origin is set. By default, the point of origin is 0m from the user.

## Rules

*   **Validation**: The interaction must have a validation.
*   **Radius**: The default radius is {{.Area.DefaultRadius}} meter.
*   **Range**: The default range is {{.Area.DefaultRange}} meters.
*   **Origin**: The point of origin for the radius is the engager or item/location from previous enactment.

## Perks

{{.AreaPerksTable}}

## Template

```yaml
interactions:
  - type: Area
    radius: {{.Area.DefaultRadius}}m # Default radius for Area interactions
    range: {{.Area.DefaultRange}}m # Default range for Area interactions
    origin: Engager # Point of origin is the Engager
    perks:
      - description: <insert description of perk here>
        add_cost: <cost of the perk>
        amount: <amount of times the perk is chosen>
        total_add_cost: <total add cost>
        energy_cost: <energy cost to use>
        is_optional: <True/False>
    validation:
      engagement_roll: <pick an Offensive Trait>
      counter_roll: <pick two Defensive Traits>
```
