# phase
## Phase

Phases are a state or passive ability that lasts for a predefined amount of time. They exist to buff or nerf someone for a specific number of rounds. A Phase lasts for a few rounds, after which the Reverse Phase starts and lasts just as long as the original phase did.

## Rules

*   Costs {{(abilityType "phase").BaseEnergy}} Energy to Use.
*   After activation, Phase is active for {{(abilityType "phase").BaseDuration}} rounds.
*   Phase ends at the start of the {{(abilityType "phase").BaseDuration}}nd turn of the character.
*   When Phase ends, the Reverse Phase starts.
*   During the Reverse Phase, no new Phases can be started for the character.
*   Phase will have an Enactment assigned to it.
*   The Enactment can be triggered as a free action at the end of the character's turn.
*   Reverse Phase will have a Bad Enactment assigned to it.
*   Bad Enactment will be applied to the character.
*   Bad Enactment must be used at the end of the character's turn as a free action.
*   If no Bad Enactment is chosen, the Bad Enactment will be the reverse of the original Enactment.
*   Phase has a knockout requirement.
*   If any knockout requirement is met, the Phase ends (and the "Bad Enactment" starts).
*   The Reverse Phase cannot be cancelled by the knockout.

## Perks

{{perksTable (abilityType "phase")}}

## Template

```yaml
ability:
  type: Phase
  phase_duration: {{(abilityType "phase").BaseDuration}} rounds
  reverse_phase_duration: {{(abilityType "phase").BaseReverseDuration}} rounds
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{(abilityType "phase").BaseEnergy}}
  enactments:
    - Type:
  Perks:
```
