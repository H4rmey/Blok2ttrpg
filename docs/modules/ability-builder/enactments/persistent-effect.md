# persistent-effect
## Enact Persistent Effect

The Enact Persistent Effect applies a lingering effect to a target, such as fire, frost, or poison damage. By default, the effect lasts for {{fieldDefault (enactment "persistent_effect") "duration"}} rounds and triggers at either the start of the target's turn or the end of the engager's turn.

## Rules

*   **Duration**: Lasts {{fieldDefault (enactment "persistent_effect") "duration"}} rounds by default.
*   **Trigger Timing**: The effect triggers at either the start of the target's turn or the end of the engager's turn.
*   **Solutions**: Targets can spend one action to attempt to remove the effect using the provided solution. There must be two solutions, which can be any Trait Roll.
*   **Applies a Single Enactment**: The persistent effect applies a single other Enactment (e.g., Enact Damage, Enact Healing).

## Perks

{{perksTable (enactment "persistent_effect")}}

## Template

```yaml
enactment:
  - type: Enact Persistent Effect
    duration: {{fieldDefault (enactment "persistent_effect") "duration"}} rounds

    trigger_timing: Start of Target's Turn or The end of the engager's turn.
    solutions:
      - Dexterity
      - Constitution
    is_optional: True
    base_enactment_energy_cost: 2
    effects:
      - type: <Enactment here type>
    interactions:
      - type:
          validation:
```