# proficiency-shift
## Enact Proficiency Shift

**Enact Proficiency Shift** abilities allow characters to temporarily enhance or weaken Traits. These abilities can be used to boost a character's **Proficiency** in a specific area or to hinder an opponent's effectiveness. 

## Rules

*   **Shift Direction**: Shift a Proficiency Tier from a Trait either up or down.
*   **Single Use**: The shift only has one use.
*   **Trait Check**: The next time a Trait check is made with the shifted Trait, you must use the shifted Proficiency.
*   **Reset**: Using the shifted Proficiency resets the Trait back to its original Proficiency.

## Perks

{{perksTable (enactment "proficiency_shift")}}

## Template

```yaml
enactments:
  - type: Enact Proficiency Shift
    shifted_trait: <trait here>
    shift_direction: <UP DOWN or>
    shift_amount: {{fieldDefault (enactment "proficiency_shift") "shift_amount"}}
    shift_uses: {{fieldDefault (enactment "proficiency_shift") "shift_uses"}}

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
