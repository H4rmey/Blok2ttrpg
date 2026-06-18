# Ability Builder

## Introduction

The **Ability Builder** is the core system used to create actions, maneuvers, spells, techniques, and special effects within Blok2ttrpg. Rather than relying on predefined spell lists or class‑locked abilities, this system allows abilities to be **constructed from modular components** with clearly defined rules and costs.

An Ability represents a **single intentional action** taken by a character. What that action does, who it affects, how it is resolved, and under which conditions it succeeds are all explicitly defined by the components chosen during creation.

Abilities are built from the following core elements:

- **Enactments** — define _what happens_ (damage, healing, movement, shifts, persistent effects, etc.).
- **Interactions** — define _how and to whom_ the Enactments are applied (self, direct, ranged, area, or area of effect).
- **Validations** — define _if and how_ the Enactments succeed or fail.
- **Rules and Perks** — modify default behavior, allowing abilities to break or bend standard limitations at a defined cost.

Every Ability must contain **at least one Enactment**. Additional Enactments may be added to create more complex effects, which are resolved **in sequence**. Each Enactment is evaluated independently unless explicitly overridden by a Perk.

The Ability Builder is intentionally **system‑agnostic** with regard to flavor. A fireball, a sword technique, a healing prayer, or a mechanical trap are all created using the same underlying rules. The narrative description of an Ability is left to the player and GM, while the mechanical behavior remains explicit and predictable.

The goal of the Ability Builder is to provide:

- **Clarity** — every effect has defined rules and resolution.
- **Flexibility** — abilities can be customized without special‑case logic.
- **Consistency** — similar effects behave the same way across the system.
- **Player Agency** — meaningful choices in how abilities are constructed and used.

This system assumes that Abilities are created **with GM oversight** and that both players and GM understand the intent behind the constructed effects. While the system allows for powerful and creative combinations, it is designed to remain readable, debuggable, and fair at the table.

---

## Abilities

An **Ability** is the overarching action a character takes, whether it be an attack, spell, or special move. Each Ability is built from smaller components:

- **Enactments** – The effects the Ability has (e.g., damage, healing, movement).
- **Interactions** – Defines the targets and range of the Enactments.
- **Validations** – Establish whether an Enactment succeeds or fails.

Every Ability must contain **at least one Enactment**. Additionally, Abilities can contain multiple Enactments that execute sequentially, adding depth and complexity to their effects.

---

### Enactments

An **Enactment** represents the specific action an Ability takes. Each Ability can have one or more Enactments, and each Enactment serves a distinct purpose. If an Ability contains multiple Enactments, they execute in a set order. However, if an Enactment fails its Validation against a specific target, subsequent Enactments will not apply to that target unless a **Perk** overrides this rule.

---

### Interactions

An **Interaction** determines how an Enactment is applied within the game world. It answers key questions such as:

- **Who is affected?** – Self, a single target, or multiple targets.
- **What is the range?** – The distance at which the Ability can affect a target.
- **How many targets can be affected?** – A single entity or an area-based effect.
- **Does it affect allies or only enemies?**
- **Is the Ability user included in the effect or excluded?**

If an Enactment does not define its own Interaction, it automatically inherits the Interaction of the previous Enactment.

**Validations are part of the Interaction component** and determine the success or failure of the Enactment for each target individually.

---

### Validations

A **Validation** determines whether an Enactment successfully affects its targets. It consists of the following components:

1. **Engagement Roll** – A roll made by the Ability user to determine effectiveness.
2. **Counter Roll** – A roll made by each target in the **Interaction** to resist the effect.
3. **Comparison** – If the Engagement Roll is **equal to or greater than** the Counter Roll, the Enactment succeeds for that target.

If the Engagement Roll fails against a target’s Counter Roll, the Enactment does not apply to that target. Furthermore, subsequent Enactments will **only fail for those specific targets** where the previous Enactment failed, rather than canceling the entire Ability.

---

## Rules & Perks

Rules and Perks further refine and balance Abilities.

- **Rules** establish the standard behavior of Abilities, Enactments, Interactions and Validations. They create a structured foundation for how these elements function.
- **Perks** override or modify these Rules to introduce unique effects or customization options.
  - Example: A healing Enactment might have a default effect of **1d4** healing, but selecting a Perk such as “Upgrade Dice Tier” can increase it to **1d6** healing.

Perks allow Abilities to break conventional rules, making them more powerful or versatile while maintaining game balance.

## Costs

Rules are free, but to apply perks there is a cost. There first cost is the cost to add the Perk, this is called the add cost. This is the cost to add the perk to your ability. Each level you gain points to create abilities, you can spend these on these abilities.

Then we also have Energy cost. This cost is used to use your ability.

There is also a fun extra to energy cost. Sometimes you do not have enough energy to use your ability. In this system it is allowed to still use your ability, but there is a catch. Either you take damage equal to the amount of energy you are missing to cast the ability or you only partially use your ability.

Let’s say you have a fire punch that explodes on hit and then lights the target on fire that costs 8 points to use. You only have 3 energy left. So you can choose to take $8 - 3=5$damage to still use the full ability or you choose to leave the “lighting on fire” part out of it so the cost of the ability is reduced to 4 points. Then you choose to take the remaining missing 1 energy as damage.

---

## Example Abilities