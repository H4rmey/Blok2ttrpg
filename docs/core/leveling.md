# leveling
## Leveling

## Introduction

As your character progresses through the world, they will gain levels. Leveling up represents your character's growth, allowing them to improve their Skills, increase their Vital stats, and become more capable in both combat and roleplay.

The maximum level a character can reach is Level {{ .Leveling.MaxLevel }}.

## Skill Points

Skill Points are used to upgrade your Proficiency Levels in various Skills (e.g., shifting a Skill from Untrained to Trained, or Expert to Master).

### Starting and Gaining Points

Both point pools follow the same rule: you start with a fixed amount at Level 1 and gain a fixed amount for every level after that.

$$Points = Start + PerLevel \times (Level - 1)$$

For Skill Points that is **{{ .Leveling.SkillPoints.Start }}** at Level 1 and **+{{ .Leveling.SkillPoints.PerLevel }}** per level thereafter. For Perk Points it is **{{ .Leveling.PerkPoints.Start }}** at Level 1 and **+{{ .Leveling.PerkPoints.PerLevel }}** per level.

Starting Skill Points are set so a new character can raise their Vital Skills off the bottom rung and still put a handful of Skills into their speciality. The per-level gain is one visible die step plus change, and it matches the Perk Point gain so both halves of your character sheet grow at the same rate.

## Proficiency Tiers

The dice ladder rises by a flat +1 average per rung. The die is capped at {{ highestDie }}; rungs above that add a flat bonus instead, which keeps high-end rolls from becoming wildly swingy. {{ defaultProficiencyName }} is the starting rung and is free. The rung below it refunds a point. HP and Energy climb by 3 per rung and Movement by 1.

| Tier | Cost | Die | HP | Movement | Energy |
| --- | --- | --- | --- | --- | --- |
{{range .Proficiencies}}| {{.Name}} | {{.Cost}} | {{.Die}} | {{index .Vitals "hp"}} | {{index .Vitals "movement"}} | {{index .Vitals "energy"}} |
{{end}}

## Leveling Table: Skill Points

{{ levelingTable "skill" }}

## Cost and Budget Rules

{{ rulesFlagsTable }}

\page