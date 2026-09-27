# leveling
## Perk Builder Leveling

## Introduction

As you level up, your character gains a deeper understanding of their powers, techniques, and spells. This growth is represented by **Perk Points**. Perk Points are spent to pay the **Build Cost** of Perks, Enactments, Interactions, and Validations when constructing or upgrading your Perks.

---

## Perk Points

At Level 1 a character starts with **{{ .Leveling.PerkPoints.Start }}** Perk Points and gains **{{ .Leveling.PerkPoints.PerLevel }}** more with every level after that, up to level {{ .Leveling.MaxLevel }}. The gain is flat: there are no bonus points at milestone levels, and the same **+{{ .Leveling.PerkPoints.PerLevel }}** per level applies to Skill Points, so both halves of your character sheet grow at the same rate.

These points are permanently invested into your perks during character creation or level-ups.

### Upgrading Perks

You do not need to create a brand new Perk every time you level up. You can spend your newly gained Perk Points to upgrade an existing Perk by adding new Perks, extending its Range, or attaching additional Enactments.

### Refunding Perk Points

Some Perks in the Perk Builder apply drawbacks or restrictions to an Perk (such as giving it an Item Dependency or increasing its Action Cost). These Perks have a **negative Build Cost**. Taking these drawbacks refunds Perk Points, allowing you to spend them elsewhere on the same Perk to make it more powerful.

There is a floor on this, however. See the table of cost rules below.

### Cost and Budget Rules

{{ rulesFlagsTable }}

## Leveling Table: Perk Points

{{ levelingTable "perk" }}
