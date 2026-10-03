# Changelog

All notable changes to Blok2 TTRPG are recorded here. The newest release is
listed first. Entries are written for players and game masters: they describe
what changed in the rules and in the app, not how the code was refactored.

## 2026-10-03

## Updated Conditions, Skills and Enact Modification

- Enact Modification renamed to Enact Shift
- Added Apply Shift button to apply shifts to your character
- Updated the Conditions Documentation

## 2026-10-02

## Updated the documentation

- Added QuickStartGuide.md
- Updated the documentation to be up to date with some of the new rulesets
- Remove documentation that is not required
- re-organised the chapters of the documentation
- Reworked the side bar when reading the documentation so it now supports 3 headers #, ## and ###
- Reworked the markdown renderer for the documentation for better readability

## 2026-09-30

### Changed: Perk creation now starts from one modal

- The Perks page toolbar's three separate buttons (**+ New Perk**, **+ Premade
  Perk**, **+ Add Passive**) are now one **+ New Perk** button. It opens a
  modal where you name your perk and pick its Type - Execution, Concentration,
  Reaction, or **Passive** - and Continue takes you to the right next step.
  Browsing the premade-perk library moved into the same modal as a secondary
  option below the form.
- Passives are now player-nameable, like every other perk. Picking one -
  configurable or fixed - opens a short naming step before it is added, and a
  configurable passive already on your sheet can still be renamed from its
  Configure button.

## 2026-09-29

A rules addition, a documentation repair, and the docs became browsable.

### Added: Wiki-style docs and changelog reader

- The Docs and Changelog pages now have a **contents sidebar** listing every
  chapter with its sections, so the rulebook can be navigated instead of
  scrolled. The entry for whatever you are reading is highlighted as you go.
- A **search box** filters the page as you type, matching both headings and body
  text and showing where each hit is with a snippet. It searches the page you
  are on, and Enter jumps to the first result.
- Every heading now has its own **anchor link**, so a specific rule can be
  linked directly.
- Because both pages are generated from the ruleset and the changelog file on
  every request, the contents and the search results are always current - there
  is nothing to rebuild or publish.

### Fixed: In-document links did nothing

- Cross-references between chapters ("see the Invoking chapter") were rendered
  without heading anchors, so clicking them did nothing at all. Every heading now
  carries an id and all of those links work.

### Changed: Docs page layout

- Download Markdown and Print / Save PDF moved into the sidebar footer.
- The sidebar collapses behind a Contents button on phones, and is left out of
  printed output entirely.

### Added: Preparing an Action

- You may now spend one of your actions on your turn to **prepare**: declare, in
  your own words, what will set you off and what you will do about it. When that
  circumstance happens, you act.
- Preparing costs **one action and nothing else**. No invoke point, no perk
  point, no energy reserved up front.
- If you prepared a perk, its cost is worked out at the moment it actually
  fires. Nothing is set aside in advance, so if something drains your Energy
  while you are waiting you may reach the trigger unable to afford what you
  planned. That is too bad.
- A prepared action resolves between actions, so it pre-empts rather than undoes,
  and it counts against the same once-per-round limit on acting out of turn that
  Reactions and Invoke Actions share.
- This means you no longer need to build a Reaction perk just to use something
  out of turn occasionally. Reactions are for circumstances worth watching for
  permanently; preparing covers the one-off.
- New rules chapter, cross-referenced from Combat and Invoking.

### Changed: Perk Creation Guide rewritten

The guide had drifted a long way from the actual builder and was teaching
mechanics that no longer exist. It has been rewritten against the live ruleset:

- It listed seven perk types; there are three (Execution, Concentration,
  Reaction). Passives are not a perk type, and Preparation, Phase and Minion are
  gone as types - Phase survives as an enactment.
- Its enactment and interaction names were all obsolete. The guide now uses the
  real ones throughout.
- The long section on "Will Always Resolve" described a rule the builder does not
  have. It is replaced by an explanation of how a chain actually resolves, how
  targets are inherited from the previous enactment, and when to give an
  enactment its own target.
- Reaction triggers are no longer described as free text. The guide explains the
  fixed trigger list, why broader triggers cost more, why triggers are
  target-neutral, and how Trigger Range differs from an enactment's own range.
- Examples no longer borrow another system's spell names, and there is a new
  "common mistakes" section covering the easiest ways to waste points.

---

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
