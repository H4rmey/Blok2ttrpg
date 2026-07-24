# concentration
## Concentration

Concentration is an Ability Type that allows an effect to persist over multiple rounds, as long as the Engager actively maintains focus. It acts like a continuous Execution. You can use it to maintain a beam of fire, hold an enemy in a telekinetic grip, or keep a protective shield active.

## Rules

*   **Single Focus**: You can only have one Concentration Ability active at a time. If you cast another Ability with the Concentration type, the first one immediately ends.
*   **Initial Cost**: Costs {{(abilityType "concentration").BaseAction}} Actions and {{(abilityType "concentration").BaseEnergy}} Energy to initiate. Both the Action and Energy cost can be adjusted with the Action +/- and Energy +/- perks (the Action cost is always at least 1).
*   **Re-trigger Timing**: When building the Ability you choose whether its Enactments re-trigger at the **Start of your Turn** or the **End of your Turn**. This fires every round while Concentration is maintained.
*   **Upkeep Cost**: Each round you must pay the chosen upkeep cost ({{(abilityType "concentration").BaseUpkeepAction}} Action or {{(abilityType "concentration").BaseUpkeepEnergy}} Energy by default) to maintain the Concentration. If you cannot (or choose not to) pay this upkeep, the Ability ends immediately.
*   **Voluntary End**: You can drop Concentration at any time as a free action.
*   **Breaking Focus**: If you take damage or are hit by an Enact State that restricts your mind or movement (like Stunned or Paralyzed), you must make a Validation check to keep focus.
    *   Make a Counter Roll (using your Mind or Constitution Trait).
    *   Compare it to the attacker's original Engagement Roll.
    *   If your roll is equal to or higher, you maintain Concentration. If lower, the Ability ends.
*   **Persistent Enactments**: Any Enactments attached to this Ability re-trigger automatically on your Target(s) every round at the chosen timing (start or end of your turn), right after you pay the upkeep cost.


## Perks

{{perksTable (abilityType "concentration")}}

## Template

```yaml
ability:
  type: Concentration
  has_item_dependency: No # If yes, enter which item
  energy_cost: {{(abilityType "concentration").BaseEnergy}}
  action_cost: {{(abilityType "concentration").BaseAction}}
  upkeep_cost: {{(abilityType "concentration").BaseUpkeepAction}} Action or {{(abilityType "concentration").BaseUpkeepEnergy}} Energy
  enactments:
    - Type:
  perks:
```
