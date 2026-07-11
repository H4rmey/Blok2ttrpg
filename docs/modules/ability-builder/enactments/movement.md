# movement
## Enact Movement

**Enact Movement** abilities allow characters to manipulate the position of themselves or their targets. These abilities can be used to push enemies away, pull allies closer, or reposition oneself strategically. Movement abilities add a dynamic element to gameplay, enabling tactical maneuvers and creative solutions to challenges. Additionally, the **Origin** of the movement can be assigned to an object or another person, allowing for even more creative and strategic uses. For example, you could attach the **Origin** to an arrow or a device, and then use a **Ranged Interaction** to throw it and pull the **Target** towards it.

## Rules

*   **Direction**: The target will move in one direction relative to an origin. Possible directions include Up, Down, Away, Towards, Forward, Left, Right, Free (extra cost).
*   **Distance**: The target will move {{.Movement.DefaultDistance}} meter by default.
*   **Origin**: The default **Origin** is the **Engager** or item/location from previous enactment.
*   **Obstacle**: If the Target moves into an obstacle, they take 1d4 damage.

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