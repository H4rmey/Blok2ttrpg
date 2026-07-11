# area-of-effect
## Area of Effect

An **Area of Effect (AoE)** Interaction functions similarly to an Area Interaction, but its effects persist for several rounds. While an **Area Interaction** might be like a single-use bomb, an **AoE** Interaction is akin to a bomb that detonates every round. Alternatively, it could represent a healing circle, where characters gain health each round they remain within the **AoE**. The possibilities are endless, so get creative!

The effect of the **AoE** does not trigger immediately. Instead, it activates either at the start of a character's turn within the **AoE** or at the end of the **Engager**'s turn.

## Rules

*   **Validation**: The interaction must have a validation.
*   **Radius**: The default radius is {{.AreaOfEffect.DefaultRadius}} meter.
*   **Range**: The default range is {{.AreaOfEffect.DefaultRange}} meters.
*   **Origin**: The point of origin for the radius is the engager.
*   **Duration**: The effect lasts for {{.AreaOfEffect.DefaultDuration}} rounds.

## Perks

{{.AreaOfEffectPerksTable}}

## Template

```yaml
interactions:
  - type: Area of Effect
    radius: {{.AreaOfEffect.DefaultRadius}}m
    range: {{.AreaOfEffect.DefaultRange}}m
    origin: Engager
    duration: {{.AreaOfEffect.DefaultDuration}} rounds
    immunity: false
    trigger_conditions:
      - Entering the Area of Effect
      - Start of character's turn within the Area of Effect
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