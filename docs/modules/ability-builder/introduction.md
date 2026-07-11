# introduction
## Ability Builder

> [!NOTE]
> This document assumes you've read the core chapters, character-attribute, character-traits, leveling and multi-dice-system.

## Introduction

The **Ability Builder** is the core system used to create actions, maneuvers, spells, techniques, and special effects. An Ability represents an **action** taken by a character. But rather than relying on predefined spell lists or class-locked abilities, this system allows abilities to be created from **Enactments**.  What that action does, who it affects, how it is resolved, and under which conditions it succeeds are all explicitly defined by the **Enactment** chosen during creation. Each **Enactment** has one **Interaction** and one **validation**.

Definitions: 

*   **Enactments** — define _what happens_ (damage, healing, movement, shifts, persistent effects, etc.).
    *   **Interactions** — define _how and to whom_ the Enactments are applied (self, direct, ranged, area, or area of effect).
    *   **Validations** — define _if and how_ the Enactments succeed or fail.

In turn each of these Components (Enactment, Validation, Interaction) has Rules and Perks:

*   **Rules**  — define how the Component works by default.
*   **Perks** — modify the Rules to upgrade the Component.

Every **Ability** must contain **at least one Enactment**. Additional Enactments may be added to create more complex effects, which are resolved **in sequence**. Each Enactment is evaluated independently unless explicitly overridden by a Perk.

The Ability Builder is intentionally **system-agnostic** with regard to flavor. A fireball, a sword technique, a healing prayer, or a mechanical trap are all created using the same underlying rules. The narrative description of an Ability is left to the player and GM, while the mechanical behavior remains the same. So a shot from an arrow might be the same as a light beam in terms of Ability Components.

## Costs

Applying perks has a cost. The first cost is the **Ability Cost** to add the Perk. Each level you gain **Ability Points** that can be spent to create abilities.

Then there is the **Energy Cost**. This cost is used to use your ability. Sometimes you do not have enough energy to use your ability. In this system it is allowed to still use your ability, but there is a catch: either you take damage equal to the amount of energy you are missing, or you only partially use your ability. The latter is done by not executing all enactments of the ability. The fireball you cast will still burn someone, but will not explode on impact anymore because you don't have the energy for that.

## Execution

So an **Ability** is made up from Enactments. Each of these Enactments describe what they do. The order in which you execute the Enactment is a bit odd compared to other systems. The order goes as follows:

1.  Resolve the **Enactment** → Check what the will be if the **Validation** succeeds.
2.  Resolve the **Interaction** → Which targets are going to be affected by this **Enactment.**
3.  Resolve the **Validation** → Are the targets going to be affected by this **Enactment.**

Let's say you want to hit someone with a ice fist. You first roll your damage 1d8 for example. This is the damage you are going to deal if you hit. Then you check, who are you going to hit, you choose the person right in front of you. Then you check if you are going to hit. So you make a **Enagement Roll** and the Target makes a **Counter Roll** of their choice.

## Additional Enactments

{{.AdditionalEnactmentTable}}