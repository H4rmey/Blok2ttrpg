# introduction
## Perk Builder

> [!NOTE]
> This document assumes you've read the core chapters, character-trait, character-skills, leveling and multi-dice-system.

## Introduction

The **Perk Builder** is the core system used to create actions, maneuvers, spells, techniques, and special effects. An Perk represents an **action** taken by a character. But rather than relying on predefined spell lists or class-locked perks, this system allows perks to be created from **Enactments**.  What that action does, who it affects, how it is resolved, and under which conditions it succeeds are all explicitly defined by the **Enactment** chosen during creation. Each **Enactment** has one **Interaction** and one **validation**.

Definitions: 

*   **Enactments** — define _what happens_ (damage, healing, movement, shifts, persistent effects, etc.).
*   **Interactions** — define _how and to whom_ the Enactments are applied (self, direct, ranged, area, or area of effect).
*   **Validations** — define _if and how_ the Enactments succeed or fail.

In turn each of these Components (Enactment, Validation, Interaction) has Rules and Perks:

*   **Rules** — define how the Component works by default.
*   **Perks** — modify the Rules to upgrade the Component.

Every **Perk** must contain **at least one Enactment**. Additional Enactments may be added to create more complex effects, which are resolved **in sequence**. Each Enactment is evaluated independently unless explicitly overridden by a Perk.

The Perk Builder is intentionally **system-agnostic** with regard to flavor. A fireball, a sword technique, a healing prayer, or a mechanical trap are all created using the same underlying rules. The narrative description of an Perk is left to the player and GM, while the mechanical behavior remains the same. So a shot from an arrow might be the same as a light beam in terms of Perk Components.

## Costs 

Applying perks has a cost. The first cost is the **Perk Cost** to add the Perk. Each level you gain **Perk Points** that can be spent to create perks.

Then there is the **Energy Cost**. This is what you pay every time you use the perk, as opposed to the Perk Cost which you pay once when you build it.

Running out of Energy does not stop you using a perk, but it costs you something else instead. That rule belongs to the resource economy rather than the builder, so it lives in [Combat](../../core/combat.md#running-out-of-energy) under "Running out of Energy".

## Executing Perks

So a **Perk** is made up from Enactments. Each of these Enactments describe what they do. The order in which you execute the Enactment is the following:

1.  Resolve the **Interaction** → Which targets are going to be affected by this **Enactment.**
2.  Resolve the **Validation** → Are the targets going to be affected by this **Enactment.**
3.  Resolve the **Enactment** → Check what the will be if the **Validation** succeeds.
   
Let's say you want to hit someone with a an **Damage Enactment**. You first check, who am i going to target. You first make your **Enactment Roll**. You then make each of your **Targets** make a **Counter Roll**. Those who fail the **Counter Roll** get damaged by the **Enactment**. 

## Additional Enactments

{{enactmentSurchargeTable}}

\page