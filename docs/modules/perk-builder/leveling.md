# leveling
## Perk Builder Leveling

As you level up, your character gains a deeper understanding of their powers, techniques, and spells. This growth is represented by **Perk Points**. Perk Points are spent to pay the **Build Cost** of Perks, Enactments, Interactions, and Validations when constructing or upgrading your Perks.

## Perk Points

At Level 1 a character starts with **{{ .Leveling.PerkPoints.Start }}** Perk Points and gains **{{ .Leveling.PerkPoints.PerLevel }}** more with every level after that, up to level {{ .Leveling.MaxLevel }}. The gain is flat: there are no bonus points at milestone levels, and the same **+{{ .Leveling.PerkPoints.PerLevel }}** per level applies to Skill Points, so both halves of your character sheet grow at the same rate.

These points are permanently invested into your perks during character creation or level-ups.

### Upgrading Perks

You do not need to create a brand new Perk every time you level up. You can spend your newly gained Perk Points to upgrade an existing Perk by adding new Perks, extending its Range, or attaching additional Enactments.

### Refunding Perk Points

Your character is evolving and sometimes that means that they no longer have a use for the perks you've created. So you can remove perks or reduce their power in order to refund the points spent on them.

### Cost and Budget Rules

{{ rulesFlagsTable }}

## Leveling Table: Perk Points

{{ levelingTable "perk" }}

\page