# Skills

Traits say who your character is; Skills determine whether what they attempt works. Below are the Skills your character might possess, grouped by what they are used for. Depending on the world setting, you may modify some of these Skills.

## Skill Points

Your Skill Point budget is set by your level:

$$SkillPoints = {{ .Leveling.SkillPoints.Start }} + {{ .Leveling.SkillPoints.PerLevel }} \times (Level - 1)$$

So you begin with **{{ .Leveling.SkillPoints.Start }}** points at Level 1 and gain **{{ .Leveling.SkillPoints.PerLevel }}** more each level, up to level {{ .Leveling.MaxLevel }}. See [Leveling](leveling.md) for the reasoning behind those numbers.

You can also gain Skill Points back by lowering a Proficiency. For instance, if you are an Expert in a Skill but want to balance out your spread, you can lower it back toward the starting rung and recover what you spent; dropping below the starting rung refunds an extra point. Spending points does not lock you into your choices; you can always reallocate them as needed.

## Skill List

Each Skill is rated by a Proficiency tier. Dice-backed Skills roll the die shown for their tier; Vital Skills use the numeric value shown instead. The *Cost* row is the Skill Point cost to raise a Skill into that tier.

{{ skillsTable }}

\page