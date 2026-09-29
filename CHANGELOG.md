# Changelog

All notable changes to Blok2 TTRPG are recorded here. The newest release is
listed first. Entries are written for players and game masters: they describe
what changed in the rules and in the app, not how the code was refactored.

## 2026-09-28

The largest update so far. Abilities became Perks, three new subsystems were
added (Passives, Invoking and Negotiation), the built-in content library was
rebuilt around the game's own setting, and printing a character finally produces
a sheet you can take to the table (all untested ofcourse).

### Renamed: Abilities are now Perks

- Everything that used to be called an *ability* is now a **perk**. The term is
  used consistently in the app, in the rulebook and in the content library.
- The perk builder replaces the ability builder, with the same workflow and a
  clearer layout.
- Perks no longer ship as one file per level. A perk is defined once and scales
  with your character, instead of existing as separate low-level and high-level
  versions.

### Added: Passives

- A new **passive** category for effects that are always on, as opposed to
  perks you spend resources to use.
- A passive library to browse what is available, and a configure step for
  passives that need a choice made when you take them.
- Passives have their own chapter in the rulebook.

### Added: Reactions and triggers

- Perks can now fire in response to something happening, rather than only on
  your own turn.
- A trigger describes the moment a perk reacts to. The builder exposes triggers
  as a first-class field, and generated perk instructions state the trigger
  explicitly so there is no ambiguity at the table.

### Added: Invoking

- **Invoking** is the new, unified answer to the question "how does this perk
  get put into play?". It covers what you spend, what you declare, and what the
  table checks before the effect resolves.
- Invoking is configurable per ruleset and documented as its own chapter.

### Added: Negotiation

- A **negotiation** subsystem for social conflict, giving talking its own
  structure in the same way combat has one.
- Configurable per ruleset, with its own chapter in the rulebook.

### Added: Skills

- **Skills** are now a distinct part of a character, separate from traits, with
  their own configuration and rules chapter.

### Changed: Traits and attributes merged

- The separate Attributes chapter is gone. Attributes and traits were two names
  for overlapping ideas, so they are now a single, reworked **Traits** system.
- Trait values, their effects and their presentation on the sheet were all
  revised.

### Changed: Packages rebuilt

- Packages now grant content more consistently: what a package gives you, and
  what it costs, follow the same rules across every category.

### Changed: Costs, energy and leveling

- The cost model was reworked so that what a perk costs reflects what it
  actually does, and comparable perks cost comparable amounts.
- Energy costs and vitals were rebalanced alongside it.
- The leveling curve and the budgets granted at each level were updated.

### Improved: Rulebook and documentation

- The generated rulebook was substantially rewritten and reorganised, with a
  configurable chapter order.
- New chapters: Invoking, Negotiation, Passives, Skills, Packages, Perks.
- The rulebook now includes generated reference tables and a schema reference,
  so the documentation cannot drift away from the configuration it describes.
- Documentation is linted as part of the build, which catches broken links and
  chapters that reference options that no longer exist.
- This changelog is available in the app under **Changelog**.

### Upgrade notes

- Existing saved characters deleted, they no longer work
