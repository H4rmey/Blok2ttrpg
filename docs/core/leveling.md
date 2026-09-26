# leveling
## Leveling

## Introduction

As your character progresses through the world, they will gain levels. Leveling up represents your character's growth, allowing them to improve their Traits, increase their Vital stats, and become more capable in both combat and roleplay.

The maximum level a character can reach is Level {{ .Leveling.MaxLevel }}.

## Trait Points

Trait Points are used to upgrade your Proficiency Levels in various Traits (e.g., shifting a Trait from Untrained to Trained, or Expert to Master).

### Starting and Gaining Points

Both point pools follow the same rule: you start with a fixed amount at Level 1 and gain a fixed amount for every level after that.

$$Points = Start + PerLevel \times (Level - 1)$$

For Trait Points that is **{{ .Leveling.TraitPoints.Start }}** at Level 1 and **+{{ .Leveling.TraitPoints.PerLevel }}** per level thereafter. For Ability Points it is **{{ .Leveling.AbilityPoints.Start }}** at Level 1 and **+{{ .Leveling.AbilityPoints.PerLevel }}** per level.

Starting Trait Points are set so a new character can raise their Vital Traits off the bottom rung and still put a handful of Traits into their speciality. The per-level gain is one visible die step plus change, and it matches the Ability Point gain so both halves of your character sheet grow at the same rate.

### Refunding Points

You can dynamically gain Trait Points by lowering a Proficiency. For instance, if you are an Expert in a Trait but want to balance things out, you can lower it back toward the starting rung and recover the points you spent. Dropping a Trait *below* the starting rung to Inept even refunds an extra point. Spending points never locks you into your choices; you can always reallocate them.

## Proficiency Tiers

The dice ladder rises by a flat +1 average per rung. The die is capped at {{ highestDie }}; rungs above that add a flat bonus instead, which keeps high-end rolls from becoming wildly swingy. {{ defaultProficiencyName }} is the starting rung and is free. The rung below it refunds a point. HP and Energy climb by 3 per rung and Movement by 1.

| Tier | Cost | Die | HP | Movement | Energy |
| --- | --- | --- | --- | --- | --- |
{{range .Proficiencies}}| {{.Name}} | {{.Cost}} | {{.Die}} | {{index .Vitals "hp"}} | {{index .Vitals "movement"}} | {{index .Vitals "energy"}} |
{{end}}

## Leveling Table: Trait Points

{{ levelingTable "trait" }}

## Cost and Budget Rules

{{ rulesFlagsTable }}
