# leveling
## Ability Builder Leveling

## Introduction

As you level up, your character gains a deeper understanding of their powers, techniques, and spells. This growth is represented by **Ability Points**. Ability Points are spent to pay the **Add Cost** of Perks, Enactments, Interactions, and Validations when constructing or upgrading your Abilities.

---

## Ability Points

At Level 1, a character starts with a base pool of Ability Points. As they level up, they gain a steady stream of new points, with larger spikes at milestone levels (Level 5 and Level 10).

These points are permanently invested into your abilities during character creation or level-ups.

### Upgrading Abilities

You do not need to create a brand new Ability every time you level up. You can spend your newly gained Ability Points to upgrade an existing Ability by adding new Perks, extending its Range, or attaching additional Enactments.

### Refunding Ability Points

Some Perks in the Ability Builder apply drawbacks or restrictions to an Ability (such as giving it an Item Dependency or increasing its Action Cost). These Perks have a **negative Add Cost**. Taking these drawbacks refunds Ability Points, allowing you to spend them elsewhere on the same Ability to make it more powerful

## Example Progression

If you build a simple "Fireball" at **Level 1**, you might spend 4 of your 10 starting points on it, leaving 6 points for a defensive Reaction ability.

By **Level 5**, you will have earned 11 additional Ability Points. You could spend 6 of those new points to add an Area of Effect Interaction to your Fireball and increase its damage dice, transforming it from a basic projectile into a massive explosion.

## Leveling Table: Ability Points

| Level | Points Gained | Total Ability Points |
| --- | --- | --- |
{{range .Leveling.AbilityPoints.Levels}}| **{{.Level}}** | +{{.PointsGained}} | {{.Total}} |
{{end}}