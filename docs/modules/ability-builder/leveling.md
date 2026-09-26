# leveling
## Ability Builder Leveling

## Introduction

As you level up, your character gains a deeper understanding of their powers, techniques, and spells. This growth is represented by **Ability Points**. Ability Points are spent to pay the **Build Cost** of Perks, Enactments, Interactions, and Validations when constructing or upgrading your Abilities.

---

## Ability Points

At Level 1 a character starts with **{{ .Leveling.AbilityPoints.Start }}** Ability Points and gains **{{ .Leveling.AbilityPoints.PerLevel }}** more with every level after that, up to level {{ .Leveling.MaxLevel }}. The gain is flat: there are no bonus points at milestone levels, and the same **+{{ .Leveling.AbilityPoints.PerLevel }}** per level applies to Trait Points, so both halves of your character sheet grow at the same rate.

These points are permanently invested into your abilities during character creation or level-ups.

### Upgrading Abilities

You do not need to create a brand new Ability every time you level up. You can spend your newly gained Ability Points to upgrade an existing Ability by adding new Perks, extending its Range, or attaching additional Enactments.

### Refunding Ability Points

Some Perks in the Ability Builder apply drawbacks or restrictions to an Ability (such as giving it an Item Dependency or increasing its Action Cost). These Perks have a **negative Build Cost**. Taking these drawbacks refunds Ability Points, allowing you to spend them elsewhere on the same Ability to make it more powerful.

There is a floor on this, however. See the table of cost rules below.

### Cost and Budget Rules

{{ rulesFlagsTable }}

## Leveling Table: Ability Points

{{ levelingTable "ability" }}
