# Enact Persistent Effect

The Enact Persistent Effect applies a lingering effect to a target, such as fire, frost, or poison damage. By default, the effect lasts for {{.PersistentEffect.DefaultDuration}} rounds and triggers at either the start of the target's turn or the end of the engager's turn.

## Rules

*   **Duration**: Lasts {{.PersistentEffect.DefaultDuration}} rounds by default.
*   **Trigger Timing**: The effect triggers at either the start of the target's turn or the end of the engager's turn.
*   **Solutions**: Targets can spend one action to attempt to remove the effect using the provided solution. There must be two solutions, which can be any Trait Roll.
*   **Applies a Single Enactment**: The persistent effect applies a single enactment (e.g., Enact Damage, Enact Healing).

## Perks

{{.PersistentEffectPerksTable}}

## Effects

{{.PersistentEffectEffectsTable}}

## Template

```yaml
enactment:
  - type: Enact Persistent Effect
    duration: {{.PersistentEffect.DefaultDuration}} rounds
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
