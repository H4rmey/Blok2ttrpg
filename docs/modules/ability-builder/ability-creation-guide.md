# Ability Creation Guide

## Ability Creation Guide

So you've read the docs and now you're staring at the Ability Builder thinking:

> "Cool, but how do I actually make a good ability?"

Yeah, that's fair.

The Builder is intentionally mechanical and flavorless. It doesn't care if you're casting a fireball, performing a monk punch, firing a laser cannon, throwing an angry goose, or summoning a giant rubber duck.

What matters is:

- What happens? (**Enactments**)
- Who does it happen to? (**Interactions**)
- How do we determine success? (**Validations**)
- When does it happen? (**Ability Type**)

Everything else is flavor.

A sword slash and a laser beam can easily be the exact same Ability mechanically.

---

# Step 1 - Pick an Ability Type

Most people should start with **Execution**.

Execution simply means:

> I want thing happen now.

Examples:

- Fireball
- Sword Slash
- Healing Touch
- Stunning Strike
- Dash Attack
- Throw Rock

Only use the other Ability Types when you specifically want special timing or behavior.

| Ability Type | What It Really Means |
|-------------|----------------------|
| Execution | Do thing now |
| Reaction | Do thing when something happens |
| Preparation | Spend actions now, trigger later |
| Concentration | Keep doing thing every round |
| Phase | Gain something now, pay for it later |
| Minion | Create another dude |

---

# Step 2 - Pick the Main Enactment

This is the actual effect.

Ask yourself:

> What should my ability do?

Usually the answer is one of these:

| Goal | Enactment |
|--------|--------|
| Hurt someone | Damage |
| Heal someone | Healing |
| Move something | Movement |
| Apply a condition | State |
| Buff/Nerf a roll | Proficiency Shift |
| Create an ongoing effect | Persistent Effect |
| Block or reduce something | Negation |

Think of Enactments as LEGO blocks.

Most abilities are simply multiple Enactments chained together.

## Example - Acid Splash

### D&D

Throw acid at somebody.

### Builder Version

```text
Execution
  Damage
    Ranged Interaction
```

Done.

---

## Example - Stunning Strike

### D&D

Punch someone and potentially stun them.

### Builder Version

```text
Execution
  Damage
  State(Stunned)
```

Damage happens first.

State happens second.

Simple.

---

# Step 3 - Combine Enactments

This is where the fun starts.

Most iconic abilities are just multiple Enactments chained together.

## Ice Lance

Deals damage and slows.

```text
Execution
  Damage
  State(Slowed)
```

---

## Explosive Arrow

Deals damage and pushes people away.

```text
Execution
  Damage
  Movement(Away)
```

---

## Vampiric Touch

Deals damage and heals the caster.

```text
Execution
  Damage
  Healing(Self)
```

---

## Hook Shot

Pulls an enemy towards you.

```text
Execution
  Damage
  Movement(Towards)
```

---

## Divine Blessing

Buff an ally's next roll.

```text
Execution
  Proficiency Shift(UP)
```

---

## Poison Blade

Deals damage and applies poison.

```text
Execution
  Damage
  Persistent Effect
    Damage
```

---

# Understanding Enactment Chains

By default, Enactments are executed in order.

If an Enactment fails its Validation, the chain stops.

## Example

```text
Execution
  Damage
  State(Stunned)
  Movement(Away)
```

Suppose the Damage Enactment fails.

Result:

```text
Damage    -> Failed
State     -> Not Executed
Movement  -> Not Executed
```

The chain ends.

---

# Understanding "Will Always Resolve"

A common misunderstanding is:

> Will Always Resolve = Automatically Hits

That is **not** how it works.

Validation still happens normally.

Counter Rolls still happen normally.

The target can still resist the effect.

The only thing this perk changes is:

> The Enactment is processed even if previous Enactments failed.

## Example

```text
Execution
  Damage
  State(Stunned)
    Will Always Resolve
```

Suppose Damage fails.

Normally the chain would end.

Instead:

```text
Damage -> Failed
State  -> Still Executed
```

The State still attempts to resolve.

Its own Validation still happens.

The target can still resist it.

The perk only ignores failures from earlier Enactments.

---

## Example - Stunning Strike

```text
Execution
  Damage
  State(Stunned)
    Will Always Resolve
```

The punch can fail.

The stun attempt still occurs.

---

## Example - Lingering Acid

```text
Execution
  Damage

  Persistent Effect
    Damage
    Will Always Resolve
```

Even if the direct acid splash doesn't land, the acid pool may still be created.

---

# Design Philosophy

## Without Always Resolve

```text
Damage
  ↓
State
  ↓
Movement
```

Failure stops the chain.

---

## With Always Resolve

```text
Damage -> Failed

State -> Still Executed

Movement -> Still Executed
```

This allows utility effects to continue even when earlier effects fail.

---

# Examples From Other Systems

## Magic Missile

### D&D

Automatically damages a target.

### Builder Version

```text
Execution
  Damage
    Reliable Validation
```

---

## Fireball

### D&D

Explosion at range.

### Builder Version

```text
Execution
  Damage
    Area Interaction
```

---

## Thunderwave

### D&D

Deals damage and pushes.

### Builder Version

```text
Execution
  Damage
  Movement(Away)
```

---

## Guiding Bolt

### D&D

Damage and easier to hit afterwards.

### Builder Version

```text
Execution
  Damage
  State(Marked)
```

---

## Hold Person

### D&D

Prevents movement.

### Builder Version

```text
Execution
  State(Paralyzed)
```

---

## Haste

### D&D

Moves faster and acts faster.

### Builder Version

```text
Phase
  State(Hastened)

Reverse
  State(Fatigued)
```

---

## Hunter's Mark

### D&D

Extra damage against one target.

### Builder Version

```text
Concentration
  State(Marked)
```

---

## Shield

### D&D

Protects when attacked.

### Builder Version

```text
Reaction
  Negation
```

---

# Step 4 - Choose Timing

The effect itself does **not** determine the Ability Type.

The timing does.

---

## Opportunity Attack

```text
Reaction
  Damage
```

Trigger:

```text
Target moves away
```

---

## Trap

```text
Preparation
  Damage
```

Trigger:

```text
Target enters area
```

---

## Flame Beam

```text
Concentration
  Damage
```

Maintains continuous damage.

---

## Rage

```text
Phase
  Proficiency Shift UP

Reverse Phase
  Proficiency Shift DOWN
```

Gain power now.

Pay for it later.

---

# Example For Every Enactment

## Damage

```text
Execution
  Damage
```

*Sword Slash*

---

## Healing

```text
Execution
  Healing
```

*Healing Word*

---

## Movement

```text
Execution
  Movement(Away)
```

*Force Push*

---

## State

```text
Execution
  State(Anchored)
```

*Root*

---

## Persistent Effect

```text
Execution
  Persistent Effect
    Damage
```

*Poison*

---

## Proficiency Shift

```text
Execution
  Proficiency Shift UP
```

*Bless*

---

## Negation

```text
Reaction
  Negation
```

*Shield*

---

# Example For Every Ability Type

## Execution

### Fireball

```text
Execution
  Damage
```

---

## Reaction

### Riposte

```text
Reaction
  Damage
```

Trigger:

```text
Target damages engager
```

---

## Preparation

### Land Mine

```text
Preparation
  Damage
  Movement(Away)
```

---

## Concentration

### Mind Prison

```text
Concentration
  State(Anchored)
```

Reapplies every round.

---

## Phase

### Battle Trance

```text
Phase
  Proficiency Shift UP

Reverse
  Proficiency Shift DOWN
```

---

## Minion

### Wolf Companion

```text
Minion

Bite:
  Damage

Howl:
  State(Frightened)
```

---

# Full Example Using Almost Everything

Let's make something stupid.

## Thunder Chain Prison

You throw magical chains.

If they hit:

- Deal damage
- Pull target closer
- Restrain them
- Continuously shock them

### Builder Version

```text
Concentration
  Damage

  Movement
    Direction: Towards

  State(Restrained)

  Persistent Effect
    Damage
```

This combines:

- ✅ Damage
- ✅ Movement
- ✅ State
- ✅ Persistent Effect
- ✅ Concentration

All in a single ability.

---

