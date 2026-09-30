# perk-creation-guide
## Perk Creation Guide

So you've read the docs and now you're staring at the Perk Builder thinking:

> "Cool, but how do I actually make a good perk?"

Yeah, that's fair.

The Builder is deliberately mechanical and flavourless. It does not care whether you are casting a fireball, throwing a punch, firing a laser, or summoning a giant rubber duck. A sword slash and a laser beam can be the exact same perk mechanically.

What it does care about is four questions:

- **What happens?** - the Enactments
- **Who does it happen to?** - each Enactment's Interaction
- **How do we find out if it worked?** - each Enactment's Validation
- **When does it happen?** - the Perk Type

Everything else is flavour, and flavour is free.

---

## Step 1 - Pick a Perk Type

The Perk Type is *timing*, and nothing else. There are three.

| Perk Type | What it really means |
|---|---|
| Execution | I want this to happen now |
| Concentration | I want this to keep happening while I hold it |
| Reaction | I want this to happen when a specific thing occurs |

**Start with Execution.** It is the plain "thing happens now" type and it covers the large majority of perks: a strike, a bolt, a heal, a shove, a knockdown.

Reach for the other two only when you specifically want their timing:

- **Concentration** costs an action to start and then an upkeep every round to keep going. Use it for effects that should persist because you are actively maintaining them, and accept that you are paying for them every round.
- **Reaction** fires out of turn, on a trigger you choose when you build it. It costs more build points *and* more energy than an Execution, because acting out of turn is worth a premium.

One thing that is **not** a perk type, despite what you might expect:

- **Preparing an action** is not a perk type. It is a thing every character can do in play: spend an action to declare a trigger and hold a perk ready. See the [Preparing an Action](#preparing-an-action) chapter. It means you do not need to build a Reaction just to use something out of turn once.

**Passive** is now in the same Type dropdown as Execution/Concentration/Reaction, but picking it does not open the builder above - it opens the passive catalogue instead, because a passive is always on and is not made of enactments. You pick one off the list rather than building it, and it is priced from the passive catalogue.

---

## Step 2 - Pick your main Enactment

The Enactment is the actual effect. Ask yourself what the perk should *do*, then find it here:

| Goal | Enactment |
|---|---|
| Hurt someone | Enact Damage |
| Restore someone | Enact Healing |
| Move something or someone | Enact Motion |
| Apply a condition (prone, stunned, burning...) | Enact Condition |
| Leave something ticking on the target | Enact Effect |
| Make a skill better or worse for a while | Enact Modification |
| Borrow power now and pay it back later | Enact Phase |
| Cripple yourself on purpose to afford more | Enact Nerf |
| Stop an effect landing on you or an ally | Enact Negation |
| Add to somebody else's damage, healing or motion | Enact Adjustment |

Think of Enactments as LEGO blocks. Most perks are one or two of these chained together, and almost every iconic effect from any other system decomposes into this list.

**A basic strike**

```text
Execution
  Enact Damage
    Direct, 1m
```

Done. That is a whole perk.

**A knockdown**

```text
Execution
  Enact Damage
    Direct, 1m
  Enact Condition (Prone)
```

Damage resolves first, then the condition. Two enactments, so two energy.

---

## Step 3 - Chain Enactments together

This is where it gets fun. Every enactment after the first adds to both the build cost and the energy cost, so a chain is powerful and expensive in equal measure.

**Lingering burn** - hit them, then leave them burning.

```text
Execution
  Enact Damage
  Enact Effect (Damage, 3 rounds)
```

**Pull and pin** - drag them to you and hold them there.

```text
Execution
  Enact Motion (towards, 3m)
  Enact Condition (Restrained)
```

**Drain** - hurt them, mend yourself.

```text
Execution
  Enact Damage
    Direct
  Enact Healing
    Self
```

**Blessing** - make an ally better at something for a few rounds.

```text
Execution
  Enact Modification (+1, 2 rounds)
    Direct
```

**All-out swing** - weaken your own defence to hit harder.

```text
Execution
  Enact Nerf (own defence, -1)
  Enact Damage
```

Enact Nerf is worth understanding: it targets **only yourself** and it *gives you back* budget. It is how you build something that is genuinely reckless rather than merely expensive.

---

### How a chain resolves

Each enactment is resolved in order, and each one goes through the same three steps:

1. **Interaction** - who is affected?
2. **Validation** - does it land on them?
3. **Enactment** - apply the effect.

So a two-enactment perk makes its own roll for each part. The damage can land and the condition can still be shrugged off, because they were validated separately.

### Targets flow down the chain

By default, every enactment after the first hits **the same target as the one before it**. That is almost always what you want: the creature you hit is the creature you knock down.

When it is not, tick **"this enactment has a different target than the enactment before it"**. That enactment then gets its own Interaction and Validation, and costs accordingly. This is how you build "damage them, heal me" - the healing needs its own target, so it needs its own interaction.

The rule of thumb: **if the whole perk happens to one creature, leave the box alone.** Every extra target you introduce is another roll and another chunk of your budget.

---

## Step 4 - Choose your Interaction

The Interaction answers *who*. There are three, and they are priced by how much reach and how many bodies they cover.

| Interaction | Use it for |
|---|---|
| Self | Anything that only affects you |
| Direct | One or more specific creatures, at a chosen range |
| Zone | An area, with a radius and a range, optionally lasting rounds |

Two things that cost real points and are easy to overspend on:

- **Range.** 1m is free. 5m, 25m and 50m each cost more. Buy the range you will actually use - a melee perk does not need 25m "just in case".
- **Targets and radius.** Each extra target on a Direct interaction, and each extra metre of Zone radius, is a significant addition. A 6m-radius blast is a very expensive perk, and it should be.

A Zone can also have a **duration**, which is how you build a lingering hazard: a patch of ground that keeps affecting whoever stands in it, rather than a one-off explosion.

---

## Step 5 - Timing in practice

The effect does not determine the Perk Type. The *timing* does. The same Enact Damage is a different perk depending on when it goes off.

**A thrown bomb** - now, in an area.

```text
Execution
  Enact Damage
    Zone, radius 2, range 5
```

**A sustained beam** - every round, while you hold it.

```text
Concentration
  Enact Damage
    Direct
```

Remember the upkeep. A Concentration costs energy at the start of each of your turns on top of what it cost to start, so an expensive one you hold for four rounds can quietly empty your pool.

**A counter-strike** - when they come at you.

```text
Reaction
  Trigger: Someone within range is attacked
  Trigger Range: 1
  Enact Damage
    Direct
```

---

### Triggers, for Reactions

A Reaction's trigger is chosen from a fixed list when you build it. You cannot write your own - if you want a bespoke trigger described in your own words, that is what [Preparing an Action](#preparing-an-action) is for.

Two things matter about triggers:

**Broader triggers cost more.** "Someone within range takes damage" fires far more often than "someone within range is reduced to 0 HP", and the price reflects that. Pick the narrowest trigger that still catches the moment you care about.

**Triggers are target-neutral.** Every trigger is written without caring whether the creature involved is friend or foe. "Someone within range is attacked" means *anyone*. Which of them your reaction actually affects is decided by your enactments, not by the trigger. That is what lets one trigger power both a bodyguard and a counter-attacker.

**Trigger Range is separate from your enactment's range.** Trigger Range is how far away the trigger may happen - your watch radius. Your enactment's interaction range is how far your response reaches. They are deliberately independent, so you can watch a wide area and still only strike something adjacent, or the reverse. A trigger range of 0 means it must happen to you or in your own space.

The full trigger list with prices lives in the [Combat](#reaction-triggers) chapter.

---

## Worked example - build one from scratch

Let's make something deliberately over-engineered.

**Chain Prison.** You throw out barbed chains. They yank the target to you, pin them, and keep grinding while they struggle.

Work through the four questions:

- **When?** It should keep working while you hold it, so: **Concentration**.
- **What?** Pull, pin, grind: **Motion**, **Condition**, **Effect**.
- **Who?** One creature at a short distance: **Direct**, 5m.
- **Does it land?** Each part rolls separately, so a strong target might be pulled but not pinned.

```text
Concentration
  Upkeep: 1 Energy

  Enact Motion (towards, 3m)
    Direct, 5m

  Enact Condition (Restrained, 2 turns)

  Enact Effect (Damage, 2 rounds)
```

The second and third enactments do not repeat the interaction, because they inherit the target of the first. Three enactments means three energy to start, plus the upkeep every round you keep it.

Now look at what it costs and ask the honest question: is this better than a plain Enact Damage for a third of the price? Sometimes. That is the decision the builder exists to make you take.

---

## Common mistakes

**Buying range you never use.** Range is one of the easiest places to waste build points. If the perk is something you do while standing next to someone, leave it at 1m.

**A trigger that fires too often.** A Reaction on a broad trigger is expensive *and* will constantly go off at the wrong moment, burning your one out-of-turn act per round on something trivial. Narrow triggers are usually the better build.

**Forgetting Concentration upkeep.** The build cost is the cheap part. Look at the per-round energy and multiply by how long you actually intend to hold it.

**Building a Reaction when preparing would do.** If you only want to use something out of turn occasionally, prepare an action instead and spend the perk points elsewhere. Build a Reaction when it is a defining part of how the character fights, not to cover a one-off.

**Adding enactments because you can.** Every extra enactment costs build points *and* energy every single time you use the perk. A three-enactment perk you cannot afford to use twice in a fight is worse than a one-enactment perk you can use all day.
