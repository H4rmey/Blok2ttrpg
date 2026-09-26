# character-traits
## Character Traits

## **Traits**

Character Attributes form the core of your character, while Traits determine the success of your actions. Below are two lists of Traits your character might possess. Depending on the world setting, you may modify some of these Traits.

---

## **Trait Points**

Your Trait Point budget is set by your level:

$$TraitPoints = {{ .Leveling.TraitPoints.Start }} + {{ .Leveling.TraitPoints.PerLevel }} \times (Level - 1)$$

So you begin with **{{ .Leveling.TraitPoints.Start }}** points at Level 1 and gain **{{ .Leveling.TraitPoints.PerLevel }}** more each level, up to level {{ .Leveling.MaxLevel }}. See [Leveling](leveling.md) for the reasoning behind those numbers.

You can also gain Trait Points back by lowering a Proficiency. For instance, if you are an Expert in a Trait but want to balance out your spread, you can lower it back toward the starting rung and recover what you spent; dropping below the starting rung refunds an extra point. Spending points does not lock you into your choices; you can always reallocate them as needed.

---

## Trait List

Each Trait is rated by a Proficiency tier. Dice-backed Traits roll the die shown for their tier; Vital Traits use the numeric value shown instead. The *Cost* row is the Trait Point cost to raise a Trait into that tier.

{{ traitsTable }}

