# concentration
## Concentration

Concentration is an Ability Type that allows an effect to persist over multiple rounds, as long as the Engager actively maintains focus. It acts like a continuous Execution. You can use it to maintain a beam of fire, hold an enemy in a telekinetic grip, or keep a protective shield active.
Rules

    Single Focus: You can only have one Concentration Ability active at a time. If you cast another Ability with the Concentration type, the first one immediately ends.

    Initial Cost: Costs 2 Actions and 3 Energy to initiate.

    Upkeep Cost: At the start of your turn, you must spend either 1 Action or 1 Energy to maintain the Concentration. If you cannot (or choose not to) pay this upkeep, the Ability ends immediately.

    Voluntary End: You can drop Concentration at any time as a free action.

    Breaking Focus: If you take damage or are hit by an Enact State that restricts your mind or movement (like Stunned or Paralyzed), you must make a Validation check to keep focus.

        Make a Counter Roll (using your Mind or Constitution Trait).

        Compare it to the attacker's original Engagement Roll.

        If your roll is equal to or higher, you maintain Concentration. If lower, the Ability ends.

    Persistent Enactments: Any Enactments attached to this Ability re-trigger automatically on your Target(s) at the start of your turn, right after you pay the upkeep cost.

Perks
Description	Energy Cost	Add Cost
Effortless: Upkeep no longer costs an Action or Energy; it just requires focus.	+0	+3
Iron Will: When rolling to maintain Concentration after taking damage, you may shift your Counter Roll up one Dice Tier.	+0	+2
Dual Focus: You can have a second Concentration Ability active, but the upkeep for both doubles.	+0	+5
Has item dependency	+0	-1
Compatible Enactments
Description	Energy Cost	Add Cost
Enact Damage	+1	+2
Enact Healing	+1	+2
Enact Movement	+0	+1
Enact Proficiency Shift	+1	+2
Enact Persistent Effect	+2	+3
Enact State	+1	+3
Template
YAML

ability:
  type: Concentration
  has_item_dependency: No # If yes, enter which item
  energy_cost: 3
  action_cost: 2
  upkeep_cost: 1 Action or 1 Energy
  enactments:
    - Type:
  perks:

