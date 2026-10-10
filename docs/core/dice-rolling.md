# Dice Rolling

## Introduction

Every roll in this system uses the die granted by your **Proficiency Level** in the relevant Skill. The ladder runs from {{ lowestDie }} at the bottom to {{ highestDie }} at the top; the die itself is capped at d12, and going beyond that will add a flat bonus instead.

{{ proficiencyDiceTable }}

{{ defaultProficiencyName }} is the default value every Skill starts at and is free; the one below it is a deliberate weakness.

## Dice Tier Mechanics

Each Proficiency Level is directly tied to its Dice Tier. When referring to dice rolls:

*   A **Trained Roll** refers to rolling a **d8**.
*   Shifting up a **Dice Tier** means upgrading to the next die in the sequence (e.g., d8 → d10).
*   Shifting up a **Proficiency Level** means improving to the corresponding Dice Tier (e.g., Trained → Expert).
*   Shifting down a **Dice Tier** means downgrading to the previous die in the sequence (e.g., d6 → d4).
*   Shifting down a **Proficiency Level** means downgrading to the corresponding Dice Tier (e.g., Expert → Trained).
*   You can also state that you can have a **die shift** of -2 (Shifting to tiers down) or a **die shift** of +1 Shifting up one time.

### Attempting a skill check

Skill checks work like most system. However it is defined by having an Engagement Roll or Counter Roll. 

**1\. The Engagement Roll**

The character initiating the action is called the **Engager**. To determine their success, the Engager checks their Proficiency Level for the relevant Skill and rolls the corresponding die (ranging from {{ lowestDie }} to {{ highestDie }}). This is the **Engagement Roll**.

**2\. The Counter Roll**

The obstacle, creature, or entity the Engager is acting against is called the **Target**. The Target opposes the Engager with a **Counter Roll**. The Die for the is determined by the GM, based on the difficulty of the roll.

**3\. Resolution**

Compare the **Engagement Roll** to the **Counter Roll** (or static difficulty). If the **Engager's** total is **equal to or higher** than the Target's total, the Skill Check is a Success. Ties always favor the Engager.

> **Example:** You attempt to hide in a crowded market. You are an Expert in Stealth, making you the **Engager** with an Engagement Roll of a d8.
> 
> The **Target** is the crowd's general awareness. Because the crowd is thick and distracted, the GM decides it will be an opposed roll using a d6.
> 
> You roll a 5. The GM rolls a 5 for the crowd. Because ties favor the Engager, your stealth check is successful.

### Die Overloading

When making an **Engagement Roll** or **Counter Roll**, if your die lands on its maximum possible value, you may choose to **Overload** the die.

To **Overload**, roll the die again, subtract **1** from the new result, and add it to your total.

If this new roll _also_ lands on its maximum value, you may choose to **Overload** the die a second time. However, the penalty increases with every subsequent roll: the second Overload takes a **\-2**, the third takes a **\-3**, and so on. (Formula: $Roll - (n - 1)$ where $n$ is the current roll number).

> **Example:** You roll a d8 and get an 8. You choose to Overload.
> 
> *   **Roll 2:** You roll a 5. Result: `8 + 5 - 1 = 12`.
> 
> But what if you roll an 8 on that second roll instead?
> 
> *   **Roll 2:** You roll an 8. Result: `8 + 8 - 1 = 15`. You can stop here, or Overload again.
> *   **Roll 3:** You roll a 4. Result: `15 + 4 - 2 = 17`.

### Critical Success/Fail

When attempting an action, compare your final total against the **Counter Roll** if the **Engagement Roll** is 4 or higher than the **Counter Roll** then it is a **Critcal Success**. If it is 4 or lower it is a **Critical Fail**. A **Critical Success/Fail** can mean multiple things. But that truly depends on what type of roll you get critical success or fail. This is described per category, and if it is not specified the DM will have to get creative :).

### Group Rolls

When doing a group roll such as Stealth. Everybody rolls a their die. 

Count the successes/fails:
| Result | Score |
|:--|:--:|
| Critical Success | +2 |
| Success | +1 |
| Failure | -1 |
| Critical Failure | -2 |

The total of the roll must be **zero or above** in order to succeed. If the final result is **4 or higher**, the group **Critically Succeeds**.

The idea is that when someone in the group fails. The other PC's can still aid the PC that failed a roll and not bring the group down as a whole.

### Aid/Help

You can choose to help someone on a **Skill Check**. This uses the same rules as a **Group Roll**. But only the people aiding and the Player making the **Skill Check** have to roll. **Aiding/Helping** someone can also be done after noticing the PC making the **Skill Check** has failed. 

\page