# Enact Movement

Enact Movement abilities allow characters to manipulate the position of themselves or their targets. Movement abilities add a dynamic element to gameplay, enabling tactical maneuvers and creative solutions to challenges.

## Rules

*   **Direction**: The target will move in one direction relative to an origin. Possible directions include Up, Down, Away, Towards, Forward, Left, Right, Free (extra cost).
*   **Distance**: The target will move {{.Movement.DefaultDistance}} meter by default.
*   **Origin**: The default origin is the engager or item/location from previous enactment.
*   **Obstacle**: If the target moves into an obstacle, they take 1d4 damage.

## Perks

{{.MovementPerksTable}}

## Template

```yaml
enactments:
  - type: Enact Movement
    minimal_distance: {{.Movement.DefaultDistance}}m
    origin: engager
    direction_options:
      - <Direction>
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
    interactions:
      - type:
          validation:
```
