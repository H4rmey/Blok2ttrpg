# Leveling

## Introduction

As your character progresses through the world, they will gain levels. Leveling up represents your character's growth, allowing them to improve their Traits, increase their Vital stats, and become more capable in both combat and roleplay.

The maximum level a character can reach is Level 10.

---

## Trait Points

Trait Points are used to upgrade your Proficiency Levels in various Traits.

### Starting Trait Points

At Level 1, your base Trait Points are calculated based on the total number of Traits used in your specific campaign setting.

For the standard 22-Trait setting, you would receive 8 Trait Points at Level 1.

### Gaining and Refunding Points

By the time you level up, you gain additional Trait Points as outlined in the leveling table below.

You can also dynamically gain Trait Points by lowering your Proficiency.

---

## Leveling Table: Trait Points

| Level | Points Gained | Total Trait Points (Standard 22-Trait Setting) |
| --- | --- | --- |
{{range .Leveling.TraitPoints.Levels}}| **{{.Level}}** | +{{.PointsGained}} | {{.Total}} |
{{end}}---

## Leveling Table: Ability Points

| Level | Points Gained | Total Ability Points |
| --- | --- | --- |
{{range .Leveling.AbilityPoints.Levels}}| **{{.Level}}** | +{{.PointsGained}} | {{.Total}} |
{{end}}---

## Proficiency Tiers

| Tier | Cost | General Dice | Offense Dice | Defense Dice | HP | Movement | Energy |
| --- | --- | --- | --- | --- | --- | --- | --- |
{{range .Proficiencies}}| {{.Name}} | {{.Cost}} | {{.Dice.General}} | {{.Dice.Offense}} | {{.Dice.Defense}} | {{index .Vitals "hp"}} | {{index .Vitals "movement"}} | {{index .Vitals "energy"}} |
{{end}}
