# leveling
## Leveling

## Introduction

As your character progresses through the world, they will gain levels. Leveling up represents your character's growth, allowing them to improve their Traits, increase their Vital stats, and become more capable in both combat and roleplay.

The maximum level a character can reach is Level 10.

## Trait Points

Trait Points are used to upgrade your Proficiency Levels in various Traits (e.g., shifting a Trait from Untrained to Trained, or Expert to Master).

### Starting Trait Points

At Level 1, your base Trait Points are calculated based on the total number of Traits used in your specific campaign setting. To calculate your starting Trait Points, use the following formula:

$$TraitPoints=(TraitAmount+2)/3$$

For example, if your setting uses the standard 22 Traits, you would receive 8 Trait Points at Level 1:

$$(22+2)/3=8$$

### Gaining and Refunding Points

By the time you level up, you gain additional Trait Points as outlined in the leveling table below.

You can also dynamically gain Trait Points by lowering your Proficiency. For instance, if you are an Expert in Dexterity but want to balance out your Traits, you can lower the Proficiency to Trained or even Untrained to gain 1 or 2 points, respectively. This means spending points does not lock you into your choices; you can always reallocate them as needed.

## Leveling Table: Trait Points

| Level | Points Gained | Total Trait Points (Standard 22-Trait Setting) |
| --- | --- | --- |
| {{range .Leveling.TraitPoints.Levels}} | **{{.Level}}** | +{{.PointsGained}} |
| {{end}}--- |  |  |

## Proficiency Tiers

| Tier | Cost | General Dice | Offense Dice | Defense Dice | HP | Movement | Energy |
| --- | --- | --- | --- | --- | --- | --- | --- |
| {{range .Proficiencies}} | {{.Name}} | {{.Cost}} | {{.Dice.General}} | {{.Dice.Offense}} | {{.Dice.Defense}} | {{index .Vitals "hp"}} | {{index .Vitals "movement"}} |
| {{end}} |  |  |  |  |  |  |  |