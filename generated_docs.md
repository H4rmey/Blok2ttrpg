- [ ] healing traits.general.medicine doesn't work
- [ ] cost for adding solution
- [ ] rename solution to multiselect 
- [ ] multiselect add small or/and to lef of items
- [ ] multiselect allow for non dropdowns
- [ ] add special case for nerf where you do not need validation or interaction because it is always applied to yourself
- [ ] allow for enactments to specify what validations and what interactions are allowed/blocked. when you allow something it will only show what is on the allowed list. when you block something it will only show everything except what is on the block list
```yaml
allowed_validations:
blocked_validations:
allowed_interactions:
blocked_interactions:
```

# blok2ttrpg
## Blok2ttrpg

This is a long running project of mine to create my own ttrpg. The primarily focus is on having a system that is flexible and can be applied to most scenarios. This means that this system does not offer any thematic input, but instead relies on the world building of the players. In theory any theme should work with this system without having to revamp or re-theme anything, as nothing has a theme yet :).

That being said. The rules for this system are sometimes very specific and sometimes very vague. This is on purpose and the documentation will (for the most part) explain in each section why that is.

I borrow a lot of input from other systems as being alone trying to think of every little mechanic myself is not doable :). So if you know other ttrpg systems, you already know a lot about things like attacking, initiative,  moving etc...

OH RIGHT! before i forget, THIS IS STILL HIGHLY and i mean HIGHLY WORK IN PROGRESS >:) So there will be issues, bugs and other stuff

All the docs were written by me by hand, letter for letter, space for space. The application surrounding this system is optional but very usefull when creating abilities (for the ability building system). That full application is vibe coded, i honestly do not have the time to do this all myself. I'd rather spend time with friends and family so when i do get the time to work on this, i focus on the docs and not the application, as nowadays when using AI correctly for code it can produce some awesome stuff.

So yes, i used AI for the coding part. Yes, i am a programmer by trade. So also yes, i highly steer the AI and manage it!

Oh right, i also never playtested this, so there is that.

uuuuh, anything else?
i'm probably still missing a bunch of mechanics

## subpages

The subpages are modules that can be used int he Blok2ttrpg. The following Modules exist:

*   Character Creation (core)
*   Ability Builder
*   Lore (WIP)

## Planned Modules:

### Character Presets (races/classes)

Basically preset traits that are usable by users to have a quick start. maybe even adding specific level up charts for thoses classes/races. Also maybe make it so there are abilities you get by selecting a specific race/class. But i never want to make abilities be tied to anything, anyone should be able to cast a fireball, that should not be a Wizard exclusive ability. 

### Predefined Abilities

A list of abilities that can be used by any character. These will costs ability points and are created using the ability creator, so they can also be upgraded. But i think i will also add abilities at different levels/tiers. So fireball level 1 just deals damage, level 2 will also apply burn effect and level 3 will explode after hitting for example.

### Items list

Same as predifined abilities, but they can not be upgraded. They might even have different rules from the specific ability creator. These will have a huge reduction on the energy cost and also do not need ability points to obtain.

### MOAR ENACTMENTS

The current list of enactments is great but it's missing a ton of features, currently i have the following enactments planned: 
*   Enact Teleportation
*   Enact Message
*   Enact Illusion

Luckely the configuration is flexible enough that adding these should not be so hard.

# character-attributes
## Character Attributes

## Attributes

Attributes define a character, situation, or environment by providing specific, factual details. They are not inherently positive or negative, and they can be permanent or temporary. Attributes can be invoked during gameplay to influence outcomes, either positively or negatively.

---

## Character Attributes

A character’s attributes describe their traits, background, and abilities. Not all attributes need to be filled in, but they should be unique and avoid duplication. Below are the attribute sections a character sheet is organised into:

**Identity**

*   Name
*   Player
*   Ancestry
*   Archetype

**Description**

*   Appearance
*   Backstory
*   Notes


---

## Temporary Attributes

Temporary Attributes are created by the GM to reflect short-term effects or conditions. They can be used to challenge or reward players and must include a duration and conditions for removal. For example:

*   A character who stayed up all night studying a case might gain the Temporary Attribute “Knowledgeable about the case” for the next day.
*   A character who touched a magic stone and lost part of their memory might gain the Temporary Attribute “Partial Memory Loss” until they defeat a great evil to regain it.

Temporary Attributes should always be discussed with the player to ensure agreement.

---

## Environment Attributes

Environment Attributes describe the setting where the player characters are located. These attributes can be used by players to enhance their chances of success or by the GM to introduce challenges. For example:

*   In a rainforest, the GM might assign the Attributes: “Wet,” “Hunting Animals,” and “Untouched by Civilization.”
*   Players can use these Attributes to their advantage (e.g., using “Wet” to create a slippery terrain for enemies) or face penalties (e.g., “Wet” making it harder to light a fire).

---

## Crafting Effective Attributes

Attributes should be specific, factual, and avoid vague language. Below are guidelines for creating effective attributes:

1.  **Be Specific**

*   **Bad:** “Smooth Talker” —> Too broad. This could apply to multiple scenario’s such as: romance, business, crime.
*   **Good:** “Flirt with the Ladies” —> Nice an specific, you are a player.

1.  **Make It Factual**

*   **Bad:** “Good at working with hands” —> Vague, when is something defined as good?
*   **Good:** “Carpenter by Trade” —> This specifies that you are a carpenter, so presumable you know how to handle wood

1.  **Avoid Vague Language**

*   **Bad:** “Tends to have a lot of luck” —> Ok so you sometimes have luck and sometimes you don’t?
*   **Better:** “Lucky” —> Better but needs more specifing
*   **Good:** “Lucky when it comes to money” —> Slotmachines, gambling, haggling you are just a lucky bastard when money is involved

1.  **Link to a Specific Trait**

*   **Bad:** “Great Investigator” —> Same as before, when is something defined as “Great”
*   **Good:** “Used to be a Detective” —> Now we specify that we have experience in being a detective

Note that these rules are only a guideline, if you want to specify “Lucky” as a character attribute and be done with it i’m not going to stop you.

---

## Using Attributes in Gameplay

Attributes can be invoked at any time to influence the outcome of a roll or event. Players or the GM can use them to modify the Dice Tier (e.g., increase or decrease the die size used for a roll). Each attribute can only grant one penalty or bonus per use.

#### Examples of Attribute Use

1.  **Positive Use**

*   Michael, a wooden puppet, wants to climb a tree. Instead of making an Athletics check, he uses his Attribute “Likes Making Furniture of Wood” to craft a makeshift ladder. The GM allows him to make a Crafting check with a higher Dice Tier.

1.  **Negative Use**

*   During combat, a barrel of alcohol catches fire. The GM rules that Michael, being made of wood, is more vulnerable to fire. His Saving Throw roll is reduced by one Dice Tier (e.g., from d8 to d6).

1.  **Balancing Penalties**

*   If Michael already has the Temporary Attribute “Stuck,” the GM should avoid multiple penalties. In the above example, the GM agrees that applying both penalties would be too harsh, so they reduce the penalty to one Dice Tier.

---

## Attribute economy

An optional way to manage how many attributes a player and the GM is has used. The GM and the players get 1 to 2 tokens. When players want to use an attribute to help themself, they have to give the GM one token. If a player chooses to hinder themself using an attribute, they gain a token from the GM.

The GM can also hinder the players by spending their tokens. If it hinders two or more players it goes to a global pool which each player can pick from. When the GM hinders a specific player, that player receives the token.

## Example 1 - Michael

**Name:** Michael  
**Type:** Wooden Puppet  
**Description:** A rebellious puppet who broke free from the Ylten Guild but still has lingering connections to them.

*   **Age:** 2 months
*   **Size:** 1.2m
*   **Alignment:** Rebel Chaotic
*   **Backstory:** Forced to work for the Ylten Guild
*   **Personality:** Overly Positive and Impulsive
*   **Traits:** Walking Database of Knowledge
*   **Appearance:** Wooden Puppet
*   **Hobbies:** Likes Making Furniture of Wood
*   **Inventory:** Hidden Compartments in His Body
*   **Quirks:** Likes Bullying Insecure People
*   **Temporary Attribute:** Lost His Left Arm (Must Rebuild It or Find It Back)

---

## Example 2 - Tavern

Michael is fighting in a tavern with the following Environment Attributes: “Alcohol,” “Wood,” “Tables,” and “Bar.”

1.  **Positive Use:**

*   Michael uses his Attribute “Likes Making Furniture of Wood” to craft a shield from the wooden tables. The GM allows him to make a Crafting check with a higher Dice Tier (e.g., d12).

1.  **Negative Use:**

*   A barrel of alcohol catches fire, and the GM rules that Michael, being made of wood, is more susceptible to fire. His Saving Throw is reduced by one Dice Tier (e.g., from d8 to d6).

1.  **Balancing Penalties:**

*   After failing the Saving Throw, Michael catches fire. The GM initially rules that the damage dice increases by one tier (e.g., d4 to d6). However, Michael’s player argues that applying both a penalty to the Saving Throw and increased damage is too harsh. The GM agrees and decides to apply only one penalty

---

# character-traits
## Character Traits

## **Traits**

Character Attributes form the core of your character, while Traits determine the success of your actions. Below are two lists of Traits your character might possess. Depending on the world setting, you may modify some of these Traits.

---

## **Trait Points**

To calculate the amount of trait points you need use the following formula:

$$Trait Points = (TraitAmount+2)/3$$

For example, if your setting uses 22 Traits, you would receive 8 Trait Points:

$$(22+2)/3=8$$

By the time you level up, you gain additional Trait Points. You can also gain Trait Points by lowering your Proficiency. For instance, if you are an Expert in Dexterity but want to balance out your Traits, you can lower the Proficiency to Trained or even Untrained to gain 1 or 2 points, respectively. This means that spending points does not lock you into your choices; you can always reallocate them as needed.

---

## Trait List

Each Trait is rated by a Proficiency tier. Dice-backed Traits roll the die shown for their tier; Vital Traits use the numeric value shown instead. The *Cost* row is the Trait Point cost to raise a Trait into that tier.

### General Traits

| Trait | Untrained | Novice | Proficient | Expert | Master | Legendary |
| --- | --- | --- | --- | --- | --- | --- |
| *Cost* | 1 | 1 | 1 | 1 | 1 | 0 |
| **Strength** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Dexterity** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Stealth** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Perception** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Nature** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Crafting** | d4 | d6 | d8 | d10 | d12 | d20 |
| **People Skill** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Performance** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Thievery** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Knowledge** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Magic** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Medicine** | d4 | d6 | d8 | d10 | d12 | d20 |

### Offense Traits

| Trait | Untrained | Novice | Proficient | Expert | Master | Legendary |
| --- | --- | --- | --- | --- | --- | --- |
| *Cost* | 1 | 1 | 1 | 1 | 1 | 0 |
| **Precision** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Power** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Mind** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Magic** | d4 | d6 | d8 | d10 | d12 | d20 |

### Defense Traits

| Trait | Untrained | Novice | Proficient | Expert | Master | Legendary |
| --- | --- | --- | --- | --- | --- | --- |
| *Cost* | 1 | 1 | 1 | 1 | 1 | 0 |
| **Reflex** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Constitution** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Mind** | d4 | d6 | d8 | d10 | d12 | d20 |
| **Magic** | d4 | d6 | d8 | d10 | d12 | d20 |

### Vital Traits

These traits use numeric values rather than dice.

| Trait | Untrained | Novice | Proficient | Expert | Master | Legendary |
| --- | --- | --- | --- | --- | --- | --- |
| **HP** | 8 | 12 | 16 | 20 | 24 | 28 |
| **Movement** | 3 | 4 | 5 | 6 | 7 | 8 |
| **Energy** | 5 | 8 | 12 | 16 | 20 | 25 |

# dice-rolling
## Dice Rolling

## Introduction

The dice system used in this system consists of six different dice: **d4, d6, d8, d10, d12, and d20**. These dice are categorized into **Dice Tiers** (1-6), each corresponding to a **Proficiency Level**:

| Proficiency | Dice |
| --- | --- |
| Clumsy | d4 |
| Untrained | d6 |
| Trained | d8 |
| Expert | d10 |
| Master | d12 |
| Legendary | d20 |

---

## Dice Tier Mechanics

Each Proficiency Level is directly tied to its Dice Tier. When referring to dice rolls:

*   A **Trained Roll** refers to rolling a **d8**.
*   Shifting up a **Dice Tier** means upgrading to the next die in the sequence (e.g., d8 → d10).
*   Shifting up a **Proficiency Level** means improving to the corresponding Dice Tier (e.g., Trained → Expert).
*   Shifting down a **Dice Tier** means downgrading to the previous die in the sequence (e.g., d6 → d4).
*   Shifting down a **Proficiency Level** means downgrading to the corresponding Dice Tier (e.g., Expert → Trained).
*   You can also state that you can have a **die shift** of -2 (Shifting to tiers down) or a **die shift** of +1 Shifting up one time.

---

### Engagement and Counter Rolls

When attempting an action where the outcome is uncertain, the acting character must make a **Trait Check**. Unlike systems that use a d20 and flat modifiers, this system relies entirely on variable Dice Tiers.

**1\. The Engagement Roll**

The character initiating the action is called the **Engager**. To determine their success, the Engager checks their Proficiency Level for the relevant Trait and rolls the corresponding die (ranging from d4 to d12). This is the **Engagement Roll**.

**2\. The Counter Roll**

The obstacle, creature, or entity the Engager is acting against is called the **Target**. The Target opposes the Engager with a **Counter Roll**, determined by the Game Master in one of two ways:

*   **Static Difficulty:** The GM sets a fixed difficulty number between 1 and 12.
*   **Opposed Die:** The GM selects a Dice Tier that represents the Target's resistance (e.g., a d10 for a sturdy vault door, or a d6 for an average guard) and rolls it.

**3\. Resolution**

Compare the **Engagement Roll** to the **Counter Roll** (or static difficulty). If the **Engager's** total is **equal to or higher** than the Target's total, the Trait Check is a Success. Ties always favor the Engager.

> **Example:** You attempt to hide in a bustling market. You are an Expert in Stealth, making you the **Engager** with an Engagement Roll of a d8.
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
**Critical Succes**: +2
**Succes**: +1
**Fail**: -1
**Critical Fail**: -2

The total of the roll must be **zero or above** in order to succeed. If the final result is **4 or higher**, the group **Critically Succeeds**.

The idea is that when someone in the group fails. the other PC's can still aid the PC that failed the roll. 

### Aid/Help

You can choose to help someone on a **Trait Check**. This uses the same rules as a **Group Roll**. But only the people aiding and the PC making the **Trait Check** have to roll. **Aiding/Helping** someone can also be done after noticing the PC making the **Trait Check** has failed.

# combat
## Combat

Combat works as most other ttrpg's, there is a type of grid where each square represents 1m. We have initiative rolls, actions, movement and all the other good stuff. Most of it is very basic so i won't go into to much detail. 

### Initiative

Rolling for initiative is done by rolling your perception + movement. PC's go before NPC's when equal values are rolled. PC's will discuess between themselfs when they roll an equal roll.

### Turns and Actions

On your turn you get three actions. By default you have 3 of actions. How much actions an ability costs can may differ. 

### Movement

Movement costs one action; it is fully allowed to just keep using actions just to move, however each subsequent movement action costs 1 energy extra (this stacks between turns).So moving 3 times in a row will cost 0 + 1 + 2 = 3 Energy.

### Attacking/Healing/Doing

Oppenents are not willing to get hit by your attacks/abilities. That is why when attacking an opponent you make an **Attack Roll** to a **Target**. In the chapter about [Dice Rolling](dice-rolling.md) We already dicussed Engegement Rolls and Counter Rolls. An **Attack Roll** is a type of **Engegment Roll**.

When Attacking/Healing/Prepping/Anythinging you always first roll the Engagement Roll to see if you hit, then you resolve the action/enactment/thing.

# Conditions

## Conditions

**Conditions** can either boost or limit your character. They can either be a collection of nerfs/buffs to your **Traits**, have some special properties, or be a combination of both.

We separate them into two groups: **Shifting Conditions** and **Fixed Conditions**.

**Shifting Conditions**: Conditions that are flexible in their use. They only **Shift** a collection of Traits up or down by x amount. For example, Blinded, Encumbered, Encouraged or Frightened.

**Fixed Conditions**: Conditions that impact something other than the Traits, like action economy or character behaviour. For example: When you are Stunned you lose one of your actions. When you are Taunted, you may only attack one preset Target.

When a **Condition** has impact on your **Traits** it always has a value representing **Die Shifts** in your **Traits**. So each Condition shifts x amount of traits in your character a y amount. This can either be temporary or permanent, depending on how the **Condition** was applied and what was discussed with the DM.

> **Example:** Your character gets **Blinded** by a flash of light because you failed a **Counter Roll**. The DM tells you that you are now **Blinded -2**. You now have **-2 Die Shift** on **Offensive Precision Rolls** and **Defensive Reflex Rolls**. The DM now rules that you will be blinded for 2 rounds.

> [!NOTE]
> As the number and type of **Traits** can differ between games, the DM or group may need to tweak what a Condition applies to. For example, in some games you may not have a **Crafting Trait**. Because it is highly encouraged to create your own list of **Traits**, we cannot make a clear definition of what does what. There is also the issue that **Conditions** are not always applicable to all scenarios.

> **Example:** Being **Frightened** can be either you are scared in the dark, or you are afraid of public speaking.

> **Example:** Being **Encumbered** might be a Condition that applies to all movement type traits but also strength. As you carry too much you also do not have any strength left to lift anything else, so it will also impact your Offensive Power stat. At least that is how I would maybe rule it, but another DM/group might not agree. It is also highly dependent on how you are Encumbered, so the removal method might change depending on the situation.

## Condition List

The following Conditions are available. For Shifting Conditions the affected Traits are up to the DM to decide and depend on the situation.

### Shifting Conditions

These conditions raise or lower a collection of traits. The value is a number of die shifts within the range shown; which traits are affected is decided at the table.

| Condition | Shift Range | Effect |
| --- | --- | --- |
| **Blinded** | -6 to 0 | Reduces certain trait rolls |
| **Encumbered** | -6 to 0 | Reduces movement and relevant traits |
| **Encouraged** | +1 to +6 | Positive trait shifts |
| **Frightened** | -6 to 0 | Negative trait shifts |

### Fixed Conditions

These conditions apply a set effect rather than a trait shift.

| Condition | Effect |
| --- | --- |
| **Taunted** | You can only target a preset Target. |
| **Swayed** | You cannot target a preset Target anymore. |
| **Untouchable** | You cannot be targeted. |
| **Ignored** | You cannot be targeted, but you can still get hit. |
| **Confused** | You indiscriminately target anyone. |
| **Vengeful** | You can only target the last entity that dealt damage to you. |
| **Distracted** | If you change your current target to a new one, your rolls will shift one down for that new target. |
| **Isolated** | You cannot target, heal, or buff your allies; you can only interact with yourself or your direct opponent. |
| **Charmed** | You must do what the Charmer says. |
| **Hypnotized** | You mimic the exact movement and action of the person who hypnotized you on their last turn. |
| **Stubborn** | You cannot repeat the same action or use the same ability two turns in a row. |
| **Paranoid** | You refuse help from anyone. |
| **Insane** | At the start of your turn, roll a die (d4) to determine random behavior. |
| **Stunned** | You are stunned; you lose one of your actions. |
| **Paralyzed** | Lose your turn. |
| **Pacified** | You can no longer attack any Target. |
| **Enraged** | You can no longer use weapons. |
| **Disarmed** | You no longer have a weapon. |
| **Silenced** | You can no longer talk. |
| **Deafened** | You can no longer listen. |
| **Stifled** | Your hands are bound; you cannot use items, potions, or consumables from your inventory. |
| **Staggered** | You can no longer use Reactions or your Preparation is cancelled. |
| **Prone** | You are knocked on the ground; use one of your actions to get up. |
| **Anchored** | You can no longer move. |
| **Restrained** | You can no longer use your hands. |
| **Slowed** | Movement speed shifted one down. |
| **Terrified** | You move 1m away from the target (or take 1d4 damage). |
| **Weakened** | Damaging rolls are shifted one die down. |
| **Fragile** | Engager that targets you, may shift their damage die by +1. |
| **Cursed** | Die shifts up are converted to die shifts down. |
| **Blessed** | Die shifts down are converted to die shifts up. |
| **Hesitant** | Your Engagement Roll must be rolled twice. |
| **Broken Gear** | Rolls for this gear are reduced by one die shift. |
| **Amplified Gear** | Rolls for this gear are upgraded by one die shift. |
| **Fatigued** | Energy cost of abilities is increased by 1. |
| **Energized** | Energy cost of abilities is reduced by 1. |
| **Delayed** | You move down one in the turn order. |
| **Hastened** | You move up one in the turn order. |
| **Echoed** | Whatever action you took last round will be executed automatically next round. |
| **Dying** | Your HP went below 0; you are out of combat until revived by an ally. |
| **Doomed** | You will move to Dying in x turns. |
| **Invincible** | You cannot be damaged. |
| **Zombified** | Healing damages instead; Damage heals instead. |
| **Linked** | You are linked to a Target; what happens to you happens to the linked Target. This links per Trait. |
| **Incorporeal** | Phase through walls. |
| **Marked** | Everything that happens to you is buffed or nerfed (pick one). |

# abilities
## Abilities

Abilities are the main source of interactions in combat. They can also be used as tools during other types of gameplay. This page will list some **predefined abilities** and some **Specialized Abilities**.

## Predefined Abilities

This page will list some predefined abilities that players and DM's can use to create their character. While there is an Ability Builder system that can be used. Sometimes people do not want to use it or they want a quick lookup for an ability. The abilitis listed here are all created in the Ability Builder.

### Ability List

## Specialized Abilities 

While the Ability Builder is perfect for creating fireballs, sword strikes, and healing spells using standard Enactments, some concepts are too abstract, vague, or narrative-driven to fit into the Ability Builder system.

Abilities like Message, Mind Reading, or Illusion often lack hard numbers. Predefined Abilities solve this by providing a conceptual base effect with hardcoded rules and a dedicated list of Perks to upgrade them.

### Specialized Abilities List

# items
## Items

Items are physical objects that characters can use to aid them in combat, exploration, or roleplay. They range from simple tools and weapons to Imbued artifacts that hold complex Abilities.

In the Ability Builder, you can often select the **Has item dependency** perk, which reduces the Ability's Add Cost by 1. This means the Ability is physically tied to the item—if the item is dropped, stolen, or broken, the character can no longer use the Ability.
Item Categories

## Item Categories

**Equipment (Weapons & Armor):** Items that grant a passive Die Shift to specific Traits while equipped. For example, a well-crafted sword might grant a +1 Die Shift to Offensive Power rolls, while heavy armor might grant a +1 Die Shift to Defensive Constitution rolls but a -1 Die Shift to Stealth.

**Consumables:** Single-use items like potions, bombs, or rations. These often trigger an Area Interaction or Direct Interaction with a predefined Enactment (like Enact Healing).

# leveling
## Leveling

## Introduction

As your character progresses through the world, they will gain levels. Leveling up represents your character's growth, allowing them to improve their Traits, increase their Vital stats, and become more capable in both combat and roleplay.

The maximum level a character can reach is Level 10.

## Trait Points

Trait Points are used to upgrade your Proficiency Levels in various Traits (e.g., shifting a Trait from Untrained to Trained, or Expert to Master).

### Starting Trait Points

At Level 1, your base Trait Points are calculated based on the total number of Traits used in your specific campaign setting. To calculate your starting Trait Points, use the following formula:

$$TraitPoints=(TraitAmount+2)/3$$

For example, if your setting uses the standard 22 Traits, you would receive 8 Trait Points at Level 1:

$$(22+2)/3=8$$

### Gaining and Refunding Points

By the time you level up, you gain additional Trait Points as outlined in the leveling table below.

You can also dynamically gain Trait Points by lowering your Proficiency. For instance, if you are an Expert in Dexterity but want to balance out your Traits, you can lower the Proficiency to Trained or even Untrained to gain 1 or 2 points, respectively. This means spending points does not lock you into your choices; you can always reallocate them as needed.

## Leveling Table: Trait Points

| Level | Points Gained | Total Trait Points (Standard 22-Trait Setting) |
| --- | --- | --- |
| **1** | +0 | 8 |
| **2** | +1 | 9 |
| **3** | +1 | 10 |
| **4** | +1 | 11 |
| **5** | +2 | 13 |
| **6** | +1 | 14 |
| **7** | +1 | 15 |
| **8** | +1 | 16 |
| **9** | +1 | 17 |
| **10** | +2 | 19 |

## Proficiency Tiers

| Tier | Cost | General Dice | Offense Dice | Defense Dice | HP | Movement | Energy |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Untrained | 1 | <no value> | <no value> | <no value> | 8 | 3 | 5 |
| Novice | 1 | <no value> | <no value> | <no value> | 12 | 4 | 8 |
| Proficient | 1 | <no value> | <no value> | <no value> | 16 | 5 | 12 |
| Expert | 1 | <no value> | <no value> | <no value> | 20 | 6 | 16 |
| Master | 1 | <no value> | <no value> | <no value> | 24 | 7 | 20 |
| Legendary | 0 | <no value> | <no value> | <no value> | 28 | 8 | 25 |

# cheat-sheet
## Cheat Sheet

## Ability Types

The when part of the ability

passive 		--> when the trigger condition is met, at any time 
reaction 		--> when the trigger condition is met, only once per turn, during combat 
preparation 	--> when the trigger condition is met, only once per preperation
execution 		--> right now
concentration 	--> after activation & at the start of your turn

## Enactments

The what part of the ability

damage 		    --> lose hp
heal 		    --> gain hp 
movement 	    --> move away/towards
reduction 	    --> reduce the effect of an enactment 
amplification 	--> reduce the effect of an enactment 

condition	    --> apply a condition to someone
negation 	    --> negate a condition completely

effect 	        --> lose/gain hp over time or move over time
shift           --> shift proficiencies now
phase 		    --> shift proficiencies now, reverse the proficiency later

stack           --> stack points now, use them later for other abilities

minion 	        --> create a minion to fight for you

## Interactions

The who part of the ability

Self    --> Target yourself (no validation needed)
Direct  --> right next to you
Ranged  --> within range
Area    --> Around a specific area triggers once
AoE     --> Around a specific area hangs around for a while

## Validation

Will it hit or miss? I guess they never miss huh?

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

*   **Rules** — define how the Component works by default.
*   **Perks** — modify the Rules to upgrade the Component.

Every **Ability** must contain **at least one Enactment**. Additional Enactments may be added to create more complex effects, which are resolved **in sequence**. Each Enactment is evaluated independently unless explicitly overridden by a Perk.

The Ability Builder is intentionally **system-agnostic** with regard to flavor. A fireball, a sword technique, a healing prayer, or a mechanical trap are all created using the same underlying rules. The narrative description of an Ability is left to the player and GM, while the mechanical behavior remains the same. So a shot from an arrow might be the same as a light beam in terms of Ability Components.

## Costs (WIP)

Applying perks has a cost. The first cost is the **Ability Cost** to add the Perk. Each level you gain **Ability Points** that can be spent to create abilities.

Then there is the **Energy Cost**. This cost is used to use your ability. Sometimes you do not have enough energy to use your ability. In this system it is allowed to still use your ability, but there is a catch: either you take damage equal to the amount of energy you are missing, or you only partially use your ability. The latter is done by not executing all enactments of the ability. The fireball you cast will still burn someone, but will not explode on impact anymore because you don't have the energy for that.

## Executing Abilities

So an **Ability** is made up from Enactments. Each of these Enactments describe what they do. The order in which you execute the Enactment is a bit odd compared to other systems. The order goes as follows:

1.  Resolve the **Interaction** → Which targets are going to be affected by this **Enactment.**
2.  Resolve the **Validation** → Are the targets going to be affected by this **Enactment.**
3.  Resolve the **Enactment** → Check what the will be if the **Validation** succeeds.
   

Let's say you want to hit someone with a an **Damage Enactment**. You first check, who am i going to target. You first make your **Enactment Roll**. You then make each of your **Targets** make a **Counter Roll**. Those who fail the **Counter Roll** get damaged by the **Enactment**. 

## Additional Enactments

| Build Cost | Energy Cost | Description |
| --- | --- | --- |
| 0 | 0 | Adding an additional Enactment beyond the first |

# Ability Builder Configuration

## Overview

The Ability Builder loads its default rules from the split YAML directory `config/ability-builder/`. Set `ABILITY_BUILDER_CONFIG` to point at another config directory or a legacy single YAML file.

The split directory is loaded into `AbilityBuilderConfig` from these section files:

- `general.yaml`
- `file_order.yaml`
- `ability_types.yaml`
- `enactments.yaml`
- `interactions.yaml`
- `proficiencies.yaml`
- `traits.yaml`
- `leveling.yaml`
- `states.yaml`

Split config loading is strict: unknown YAML keys are rejected. Keep examples and edits aligned with the schema names exactly.

Legacy single-file configs, such as `config/dnd.yaml` and `config/pathfinder2e.yaml`, are still supported. When a config section has no generic `fields` schema, the system falls back to older hardcoded cost paths for that section.

## File-by-file reference

### `general.yaml`

Controls root metadata and global defaults:

- `version`
- `profile_id`
- `combat.actions.amount`
- `additional_enactment`
- `dice.damage`
- `dice.generic`
- `validations`
- generic validation `fields`

Example:

```yaml
version: 1
profile_id: ability-builder
combat:
  actions:
    amount: 3
additional_enactment:
  add_cost: 1
  energy_cost: 1
  description: "Adding an additional Enactment beyond the first"
```

### `file_order.yaml`

Controls the order in which Markdown files under `./docs/` are appended to generated output.

```yaml
file_order:
  - ./docs/modules/ability-builder/introduction.md
  - ./docs/modules/ability-builder/guide.md
```

The list must be exhaustive: every `.md` file under `./docs/` must appear exactly once.

### `ability_types.yaml`

Defines ability type display names, descriptions, base energy/action values, legacy cost settings, compatible enactments, and generic `fields` for Execution, Reaction, Phase, Minion, Preparation, Concentration, and Passive.

Example:

```yaml
ability_types:
  execution:
    name: "Execution"
    description: "Performed instantly during a character's turn."
    base_energy: 3
    base_action: 2
    compatible_enactments:
      - Enact Damage
      - Enact Healing
    fields:
      - key: item_dep
        label: "Has Item Dependency"
        type: checkbox
        cost:
          add_cost: -1
          energy_cost: 0
```

### `enactments.yaml`

Defines enactment type names, descriptions, base costs, legacy cost settings, and generic `fields` for Enact Damage, Healing, Movement, Proficiency Shift, Persistent Effect, State, and Negation.

Example:

```yaml
enactments:
  damage:
    type: "Enact Damage"
    description: "Inflicts damage to a target."
    base_cost:
      add_cost: 2
      energy_cost: 1
```

### `interactions.yaml`

Defines interaction type names, descriptions, base costs, legacy cost settings, and generic `fields` for Self, Direct, Ranged, Area, and Area of Effect.

Example:

```yaml
interactions:
  direct:
    type: "Direct"
    description: "Affects a single target within 1m."
    default_range: 1
    default_targets: 1
    base_cost:
      add_cost: 0
      energy_cost: 0
```

### `traits.yaml`

Provides trait lists used by option sources.

```yaml
traits:
  general:
    - Strength
    - Dexterity
  offense:
    - Precision
  defense:
    - Reflex
  vital:
    - HP
```

### `proficiencies.yaml`

Defines proficiency tiers, point cost, dice per category, and vital values.

```yaml
proficiencies:
  - id: trained
    name: "Trained"
    cost: 1
    dice:
      general: "d8"
      offense: "d8"
      defense: "d8"
    vitals:
      hp: 16
      movement: 5
      energy: 12
```

### `leveling.yaml`

Defines trait point and ability point progression.

```yaml
leveling:
  max_level: 10
  trait_points:
    standard_trait_count: 22
    starting_formula: "(trait_count + 2) / 3"
    levels:
      - level: 1
        points_gained: 0
        total: 8
```

### `states.yaml`

Defines the Enact State data set:

- `additional_state`: surcharge for each selected state after the first.
- `general_states`: flexible shift states with min/max bounds and per-shift cost.
- `specific_states`: fixed-cost named states.

```yaml
additional_state:
  add_cost: 1
  energy_cost: 0

general_states:
  - id: encouraged
    name: "Encouraged"
    description: "Positive trait shifts"
    min_shift: 1
    max_shift: 6
    shift_cost:
      add_cost: 2
      energy_cost: 1
```

## Generic field schema

Generic fields appear under `fields` or `row_fields`. Supported `FieldConfig` keys are:

| Key | Purpose |
| --- | --- |
| `key` | Stable submitted field key. This is persisted in generic field maps. |
| `label` | UI label. |
| `type` | Field type: `checkbox`, `dropdown`, `free_text`, `free_number`, `solutions`, or `states`. |
| `cost` | Field-level `add_cost` / `energy_cost`. |
| `options` | Inline dropdown options. |
| `options_source` | Dynamic option source name resolved by the browser. |
| `default` | Default submitted/display value. |
| `min` | Minimum numeric value. |
| `max` | Maximum numeric value. |
| `step` | Numeric increment size. Defaults to `1` for cost calculation when omitted. |
| `rounding` | Optional step rounding: `ceil` or `floor`. |
| `per_step` | `increase` / `decrease` costs for `free_number` deltas. |
| `default_count` | Default row count for repeatable fields. |
| `per_item` | `increase` / `decrease` costs for repeatable row count deltas. |
| `export` | Export mapping with `key`, optional `suffix`, and `omit_when_default`. |
| `row_fields` | Sub-fields used by `solutions` and `states` rows. |
| `stores_to` | Maps generic values to typed model fields for export compatibility. |
| `visibility_when` | Controlling sibling field key. |
| `show_when` | Required controlling value for visibility. |

`FieldOption` supports `value`, `label`, optional `cost`, and optional child `fields`. `CostDefinition` supports `add_cost`, `energy_cost`, optional `description`, and optional `step`.

## Supported field types

### `checkbox`

Charges `cost` only when checked.

```yaml
- key: always
  label: "Will always resolve"
  type: checkbox
  cost:
    add_cost: 5
    energy_cost: 3
```

### `dropdown`

A dropdown can use inline `options` or an `options_source`. Do not mix both on the same field.

A field-level `cost` is charged when any non-empty option is selected:

```yaml
- key: offense
  label: "Offensive Trait (extra die)"
  type: dropdown
  options_source: traits_offense
  cost:
    add_cost: 4
    energy_cost: 2
```

Inline options can carry per-option cost:

```yaml
- key: shift_dir
  label: "Direction"
  type: dropdown
  options:
    - value: UP
      label: "UP"
      cost:
        add_cost: 0
        energy_cost: 0
    - value: DOWN
      label: "DOWN"
      cost:
        add_cost: 0
        energy_cost: 0
```

### `free_text`

Stores text and has no direct cost.

```yaml
- key: other
  label: "Other Roll Text"
  type: free_text
  visibility_when: source
  show_when: other
```

### `free_number`

Uses `default`, `min`, `max`, `step`, optional `rounding`, and optional `per_step`.

```yaml
- key: range
  label: "Range"
  type: free_number
  default: 0
  min: 0
  max: 10
  step: 2
  rounding: ceil
  per_step:
    increase:
      add_cost: 1
      energy_cost: 0
```

For values above `default`, `per_step.increase` is multiplied by the number of steps. For values below `default`, `per_step.decrease` is multiplied by the absolute number of steps. `rounding: ceil` rounds positive partial steps up; `rounding: floor` rounds down through integer division.

### `solutions`

Repeatable row field. The form submits parallel arrays named `<field>__<subfield>`, and the server validates/evaluates each row using `row_fields`.

Blank rows are ignored before `per_item` calculation. A `default` value can seed initial rows; when it is an array of objects with keys matching `row_fields`, each object populates one row:

```yaml
- key: counter_trait
  label: "Counter Trait"
  type: solutions
  default_count: 2
  default:
    - type: defense
      value: Reflex
    - type: defense
      value: Constitution
  options_source: traits_all
  row_fields:
    - key: type
      label: "Counter Type"
      type: dropdown
      default: defense
      options:
        - value: defense
          label: "Defensive Trait"
          cost:
            add_cost: 0
            energy_cost: 0
        - value: general
          label: "General Trait"
          cost:
            add_cost: 4
            energy_cost: 0
        - value: offense
          label: "Offensive Trait"
          cost:
            add_cost: 4
            energy_cost: 0
        - value: previous
          label: "Use result of previous"
          cost:
            add_cost: 3
            energy_cost: 1
    - key: value
      label: "Counter Trait"
      type: dropdown
      options_source: traits_all
  per_item:
    increase:
      add_cost: 0
      energy_cost: 0
    decrease:
      add_cost: 3
      energy_cost: 1
```

Add/remove buttons on a `solutions` field always add or remove exactly one row at a time. `default_count` controls how many rows are shown initially.

### `states`

Repeatable Enact State rows backed by `states.yaml`. Blank rows are ignored; partially filled rows are invalid.

```yaml
- key: states
  label: "States"
  type: states
  default_count: 1
  row_fields:
    - key: state_kind
      label: "State Type"
      type: dropdown
      default: ""
      options:
        - value: specific
          label: "Specific State"
          cost:
            add_cost: 0
            energy_cost: 0
        - value: general
          label: "General State (shift)"
          cost:
            add_cost: 0
            energy_cost: 0
    - key: specific_state
      label: "Specific State"
      type: dropdown
      options_source: states_specific
      visibility_when: state_kind
      show_when: specific
    - key: general_state
      label: "General State"
      type: dropdown
      options_source: states_general
      visibility_when: state_kind
      show_when: general
    - key: shift_amount
      label: "Shift Amount"
      type: free_number
      default: 1
      step: 1
      visibility_when: state_kind
      show_when: general
```

## Cost evaluation rules

Cost calculation is server-authoritative:

- The browser mirrors the calculation for live feedback only.
- The server recomputes costs on save.
- Schema-backed cards do not trust hidden build/cast values from the form.
- Legacy no-schema paths still use submitted build/cast fallback behavior.

Rules:

- Enactments and interactions start from `base_cost`.
- The first enactment added to an ability is free: its component `base_cost` is waived, so adding the first enactment costs no build or energy. Field-driven costs on that first enactment (checkboxes, dropdowns, numbers, states, etc.) still apply normally.
- Each enactment beyond the first pays its full `base_cost` plus the `additional_enactment` surcharge from `general.yaml`.
- Ability types start from `base_energy` and `base_action` where applicable.

- `checkbox` cost applies only when checked.
- `dropdown` field-level `cost` applies when the selected value is non-empty.
- Inline dropdown option `cost` also applies for the selected option.
- `free_number` applies `per_step.increase` or `per_step.decrease` from the configured `default`.
- `rounding: ceil` rounds positive partial steps up; `rounding: floor` rounds down.
- Enact State adds specific/general state row costs from `states.yaml` plus `additional_state` once per selected state after the first.

Example: this Ranged field charges `+1` build for every full 2m above 10m and floors partial steps.

```yaml
- key: range
  label: "Range"
  type: free_number
  default: 10
  min: 10
  max: 20
  step: 2
  rounding: floor
  per_step:
    increase:
      add_cost: 1
      energy_cost: 0
    decrease:
      add_cost: 0
      energy_cost: 0
```

## Visibility rules

`visibility_when` references another field's `key`. The field is active only when that controlling value equals `show_when`. Hidden/inactive fields do not contribute cost.

For checkboxes, the submitted checked value is the string `"true"`, not `"on"`.

```yaml
- key: item_dep
  label: "Has Item Dependency"
  type: checkbox
  cost:
    add_cost: -1
    energy_cost: 0
- key: item_name
  label: "Item Name"
  type: free_text
  visibility_when: item_dep
  show_when: "true"
```

Dropdown-controlled visibility:

```yaml
- key: source_trait
  label: "Trait"
  type: dropdown
  options_source: traits_offense
  visibility_when: source
  show_when: trait
```

## Default-driven visibility

When a controlling field has no submitted value, `visibility_when` falls back to that field's configured `default`. This ensures dependent fields become visible as soon as the default is active, without requiring an explicit user selection first.

```yaml
- key: engage_mode
  label: "Engage Roll Type"
  type: dropdown
  default: trait
  options:
    - value: trait
      label: "Trait Roll"
- key: engage_trait
  label: "Trait"
  type: dropdown
  options_source: traits_all
  visibility_when: engage_mode
  show_when: trait
```

With the default above, `engage_trait` is visible immediately when the card renders.

## Trait dropdown grouping and filtering

Dropdowns backed by `traits_general`, `traits_offense`, `traits_defense`, or `traits_all` are rendered as grouped `<optgroup>` lists by category.

For `solutions` rows that include a `type` field (for example `counter_trait`), the `value` dropdown is filtered to traits of the selected type. When the type is empty or `previous`, all traits are shown grouped.

## Option sources

Option sources are resolved in `static/js/builder.js`. Adding a new source name requires a JavaScript mapping.

| Source | Resolves to |
| --- | --- |
| `traits_general` | `D.generalTraits` from `traits.general` |
| `traits_offense` | `D.offenseTraits` from `traits.offense` |
| `traits_defense` | `D.defenseTraits` from `traits.defense` |
| `traits_all` | General + offense + defense traits |
| `dice_damage` | `D.damageDiceOptions` from `dice.damage` |
| `dice_generic` | `D.genericDieOptions` from `dice.generic` |
| `states_general` | `C.states.general_states` |
| `states_specific` | `C.states.specific_states` |
| `directions_all` | `D.directionOptions` |
| `directions` | `D.directionOptions` |
| `shift_directions` | `D.shiftDirectionOptions` |
| `trigger_timings` | `D.triggerTimings` |
| `aoe_trigger_timings` | `D.aoeTriggerTimings` |
| `knockout_options` | `D.knockoutOptions` |
| `reaction_triggers` | `D.reactionTriggers` |
| `ability_types` | `D.abilityTypes` |
| `enactment_types` | `D.allEnactmentTypes` |
| `interaction_types` | `D.interactionTypes` |

## States configuration

Specific states have fixed `add_cost` and `energy_cost`:

```yaml
specific_states:
  - id: taunted
    name: "Taunted"
    description: "You can only target a preset Target."
    add_cost: 2
    energy_cost: 0
```

General states use `min_shift`, `max_shift`, and `shift_cost` per absolute shift:

```yaml
general_states:
  - id: frightened
    name: "Frightened"
    description: "Negative trait shifts"
    min_shift: -6
    max_shift: 0
    shift_cost:
      add_cost: 1
      energy_cost: 0
```

Validation rules:

- `specific` rows require `specific_state`.
- `general` rows require `general_state` and `shift_amount` within that state's range.
- Unknown state IDs are rejected.
- Blank rows are ignored before surcharge calculation.
- `additional_state` is applied once per selected state after the first.

## How to add or change config

Safe workflows:

1. Change labels, costs, or existing option values directly in YAML.
2. Add an option to an existing field by appending to its `options` list.
3. Add a field to an existing type by appending a valid `FieldConfig` under `fields`.
4. Add a specific or general state in `states.yaml` and use an existing `states_*` option source.
5. Add a new ability, enactment, or interaction type by adding its config entry and updating compatibility lists such as `compatible_enactments` as needed.

After edits, validate with:

```bash
go test ./...
go vet ./...
go build ./...
node --check static/js/builder.js
go run .
```

For documentation changes, also run:

```bash
go run ./cmd/docs
```

## Known boundaries

- Generic YAML import/export is not fully schema-driven unless implemented separately.
- Existing saved abilities may not migrate cleanly when config keys change.
- Field keys are persisted in generic `Fields` maps, so renaming a field key changes saved-data compatibility.
- Config-defined type lists are used by the builder, but model/export compatibility may still depend on existing typed fields for some paths.

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
| Passive | Always on, free to use, triggers whenever |
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

## Execution

Execution is the most basic form for an Ability. It is simply the "I want to do this now" Ability Type. Executions can be anything from casting a fireball to summoning a shield to block an attack or preparing a parry.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Energy +/-** - Adjust the Energy cost of this ability. Lowering energy costs extra build points; raising it refunds some. Any whole number from **-2 to 2** (starts at 0). Cost: 2 build per step.
3. **Action +/-** - Adjust the amount of Actions it will cost to use this ability. Any whole number from **-1 to 1** (starts at 0). Cost: 2 build per step.

## Preparation

Just like a Reaction, a Preparation works outside the regular turn order. It follows the exact same rules as a Reaction, but instead of passively sitting in the background, a Preparation costs an action to prepare, and in turn costs far less Energy to use.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Energy +/-** - Any whole number from **-2 to 2** (starts at 0). Cost: 2 build per step.
3. **Action +/-** - Any whole number from **-1 to 1** (starts at 0). Cost: 2 build per step.
4. **Triggers** - You start with **one** trigger and may add or remove triggers. Cost: Free per trigger.

   For each trigger, choose one of:
   - You, An Ally, An Opponent, Someone Else

   For each trigger, choose one of:
   - moves away from you, moves towards you, moves past you, enters interaction range, leaves interaction range, ends their turn within range, is moved by an effect, gets hit by damage of a type, deals damage of a type, gets healed by an ability of a type, gets hit by a weapon of a type, starts casting an ability of a type, gets targeted by an ability of a type, gets hit with an enactment of a type, resolves an enactment of a type, makes a trait check of a type, fails a validation of a type, succeeds on a validation of a type, becomes affected by a condition of a type, recovers from a condition of a type, falls unconscious, dies, moves, takes any damage, deals any damage, gets healed, casts any ability, gets targeted by any ability, is hit by any enactment, makes any trait check, fails any validation, succeeds on any validation, becomes affected by any condition, recovers from any condition

   For each trigger, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

   For each trigger, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

   For each trigger, choose one of:
   - Unarmed, Sword, Axe, Mace, Spear, Dagger, Bow, Crossbow, Thrown, Firearm, Staff, Wand, Shield

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Enact Amplification, Enact Condition, Enact Damage, Enact Effect, Enact Healing, Enact Movement, Enact Negation, Enact Nerf, Enact Phase, Enact Reduction, Enact Shift

   For each trigger, choose one of:
   - Enact Amplification, Enact Condition, Enact Damage, Enact Effect, Enact Healing, Enact Movement, Enact Negation, Enact Nerf, Enact Phase, Enact Reduction, Enact Shift

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - 

   For each trigger, choose one of:
   - 
5. **Range** - Any whole number from **1 to 6** (starts at 1). Cost: Free per meter.
6. **Uses** - Any whole number from **1 to 3** (starts at 1). Cost: Free per step.

## Reaction

Reactions are Abilities that trigger outside your normal action economy. Reactions trigger when someone else does something. When the trigger happens, the linked Enactment is executed. For example, you could have a reaction that triggers whenever someone runs towards you, Enacting a healing effect on yourself.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Energy +/-** - Any whole number from **-2 to 2** (starts at 0). Cost: 2 build per step.
3. **Triggers** - You start with **one** trigger and may add or remove triggers. Cost: Free per trigger.

   For each trigger, choose one of:
   - You, An Ally, An Opponent, Someone Else

   For each trigger, choose one of:
   - moves away from you, moves towards you, moves past you, enters interaction range, leaves interaction range, ends their turn within range, is moved by an effect, gets hit by damage of a type, deals damage of a type, gets healed by an ability of a type, gets hit by a weapon of a type, starts casting an ability of a type, gets targeted by an ability of a type, gets hit with an enactment of a type, resolves an enactment of a type, makes a trait check of a type, fails a validation of a type, succeeds on a validation of a type, becomes affected by a condition of a type, recovers from a condition of a type, falls unconscious, dies, moves, takes any damage, deals any damage, gets healed, casts any ability, gets targeted by any ability, is hit by any enactment, makes any trait check, fails any validation, succeeds on any validation, becomes affected by any condition, recovers from any condition

   For each trigger, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

   For each trigger, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

   For each trigger, choose one of:
   - Unarmed, Sword, Axe, Mace, Spear, Dagger, Bow, Crossbow, Thrown, Firearm, Staff, Wand, Shield

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Enact Amplification, Enact Condition, Enact Damage, Enact Effect, Enact Healing, Enact Movement, Enact Negation, Enact Nerf, Enact Phase, Enact Reduction, Enact Shift

   For each trigger, choose one of:
   - Enact Amplification, Enact Condition, Enact Damage, Enact Effect, Enact Healing, Enact Movement, Enact Negation, Enact Nerf, Enact Phase, Enact Reduction, Enact Shift

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - 

   For each trigger, choose one of:
   - 
4. **Range** - Any whole number from **1 to 6** (starts at 1). Cost: Free per meter.
5. **Uses** - Any whole number from **1 to 3** (starts at 1). Cost: Free per step.

## Concentration

Concentration is an Ability Type that allows an effect to persist over multiple rounds, as long as the Engager actively maintains focus. It takes one action to start the Concentration and then at the start of each of your turns you have to spend energy for the upkeep.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Energy +/-** - Any whole number from **-2 to 2** (starts at 0). Cost: 2 build per step.
3. **Action +/-** - Any whole number from **-1 to 1** (starts at 0). Cost: 2 build per step.
4. **Upkeep Cost** - Choose one of:
   - 1 Action or 1 Energy, 1 Action, 1 Energy
5. **Effortless (upkeep is free)** - Enable Effortless (upkeep is free). Cost: Free.

## Passive

Passives are Abilities that are always on. They work just like a Reaction, triggering when something happens, but unlike a Reaction a Passive does not cost any Energy or Actions to use and is not bound to your action economy at all. Whenever the trigger happens, the linked Enactment is executed. For example, you could have a passive that triggers whenever someone damages you, Enacting a small healing effect on yourself. Because a Passive is free to use and can trigger whenever, it is the most expensive Ability Type to build. This higher base build cost is the price you pay for never having to spend Energy or Actions on it.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Energy +/-** - Any whole number from **-2 to 2** (starts at 0). Cost: 2 build per step.
3. **Triggers** - You start with **one** trigger and may add or remove triggers. Cost: Free per trigger.

   For each trigger, choose one of:
   - You, An Ally, An Opponent, Someone Else

   For each trigger, choose one of:
   - moves away from you, moves towards you, moves past you, enters interaction range, leaves interaction range, ends their turn within range, is moved by an effect, gets hit by damage of a type, deals damage of a type, gets healed by an ability of a type, gets hit by a weapon of a type, starts casting an ability of a type, gets targeted by an ability of a type, gets hit with an enactment of a type, resolves an enactment of a type, makes a trait check of a type, fails a validation of a type, succeeds on a validation of a type, becomes affected by a condition of a type, recovers from a condition of a type, falls unconscious, dies, moves, takes any damage, deals any damage, gets healed, casts any ability, gets targeted by any ability, is hit by any enactment, makes any trait check, fails any validation, succeeds on any validation, becomes affected by any condition, recovers from any condition

   For each trigger, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

   For each trigger, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

   For each trigger, choose one of:
   - Unarmed, Sword, Axe, Mace, Spear, Dagger, Bow, Crossbow, Thrown, Firearm, Staff, Wand, Shield

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Execution, Reaction, Minion (Deprecated), Preparation, Concentration, Passive

   For each trigger, choose one of:
   - Enact Amplification, Enact Condition, Enact Damage, Enact Effect, Enact Healing, Enact Movement, Enact Negation, Enact Nerf, Enact Phase, Enact Reduction, Enact Shift

   For each trigger, choose one of:
   - Enact Amplification, Enact Condition, Enact Damage, Enact Effect, Enact Healing, Enact Movement, Enact Negation, Enact Nerf, Enact Phase, Enact Reduction, Enact Shift

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

   For each trigger, choose one of:
   - 

   For each trigger, choose one of:
   - 
4. **Range** - Any whole number from **1 to 6** (starts at 1). Cost: Free per meter.
5. **Uses** - Any whole number from **1 to 3** (starts at 1). Cost: Free per step.

## Enact Amplification

Enact Amplification allows you to increase the effect of an enactment you or someone else are the target for. It always has an Amplification Die which determines the amplification of the effect.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Amplification Mode** - Choose how the amplification is determined: a fixed die-size shift up/down the die ladder, or a die roll. Choose one of:
   - Die Shift, Die Roll
3. **Flat Bonus** - Any whole number from **0 to 8** (starts at 0). Cost: Free per +1.

## Enact Condition

Enact Condition applies a condition to a target (e.g., prone, stunned, charmed). A Condition always has a value. See the Conditions chapter for the full list of conditions and their effects.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Condition** - Choose one of:
   - *Conditions:* Blinded, Encumbered, Encouraged, Frightened, Taunted, Swayed, Untouchable, Ignored, Confused, Vengeful, Distracted, Isolated, Charmed, Hypnotized, Stubborn, Paranoid, Insane, Stunned, Paralyzed, Pacified, Enraged, Disarmed, Silenced, Deafened, Stifled, Staggered, Prone, Anchored, Restrained, Slowed, Terrified, Weakened, Fragile, Cursed, Blessed, Hesitant, Broken Gear, Amplified Gear, Fatigued, Energized, Delayed, Hastened, Echoed, Dying, Doomed, Invincible, Zombified, Linked, Incorporeal, Marked
3. **Duration (turns)** - Choose one of:
   - 1 turn, 2 turns, 3 turns, 4 turns, 5 turns, 6 turns, Unlimited
4. **Solutions** - You start with **two** solutions and may add or remove solutions. Cost: Free per solution.

   For each solution, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Defense:* Reflex, Constitution, Mind, Magic

## Enact Damage

Enact Damage allows characters to inflict harm on their enemies. It always has a Source Die and can have added bonuses.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Source** - The offensive trait whose die is rolled for damage. Higher proficiency in the trait means a larger die. Choose one of:
   - Precision, Power, Mind, Magic
3. **Flat Bonus** - Any whole number from **0 to 20** (starts at 0). Cost: Free per +1.
4. **Damage Types** - You start with **one** damage type and may add or remove damage types. Cost: Free per damage type.

   For each damage type, choose one of:
   - Physical, Slashing, Piercing, Bludgeoning, Fire, Cold, Lightning, Thunder, Acid, Poison, Psychic, Necrotic, Radiant, Force, Arcane, Nature, Holy, Shadow, Chaos

## Enact Effect

The Enact Effect applies a lingering effect to a target, such as fire, frost, or poison damage. By default, the effect lasts for  rounds and triggers at the start of the target's turn. On the target's turn they can re-roll the solution to get rid of the effect or take an action to remove it.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Name** *(optional)* - A note you can write on the ability. No cost.
3. **Applies** - Choose one of:
   - Damage, Heal, Move
4. **Solutions** - You start with **two** solutions and may add or remove solutions. Cost: Free per solution.

   For each solution, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Defense:* Reflex, Constitution, Mind, Magic

## Enact Healing

Enact Healing abilities allow characters to restore health to themselves or others. It always has a Source Die and may contain other bonuses.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Source** - Choose one of:
   - *Generic:* 1d4, 1d6, 1d8, 1d10, 1d12, 1d20
   - *Medicine:* Medicine
3. **Flat Bonus** - Any whole number from **0 to 20** (starts at 0). Cost: Free per +1.

## Enact Minion

Creates a minion to fight for you. By default the minion does not have any traits, has 1hp and no movement or energy. This can be upgraded.

## Enact Movement

Enact Movement allows you to move a Target up to a preset amount of meters.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Distance** - Any whole number from **1 to 20** (starts at 1). Cost: Free per meter.
3. **Directions** - You start with **one** direction and may add or remove directions. Cost: Free per direction.

   For each direction, choose one of:
   - Towards, Away

## Enact Negation

Enact Negation allows characters to ignore or nullify the effects of an enactment you or someone else are the target for.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Ability hits Engager instead** - Enable Ability hits Engager instead. Cost: Free.

## Enact Nerf

Enact Nerf can only be applied to yourself. Enact Nerf will apply a state or proficiency shift to your character to gain ability points or energy.

Enact Nerf can only target yourself.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Trait** - Choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic
3. **Shift -/+** - Any whole number from **-6 to -1** (starts at -1). Cost: Free per step.

## Enact Phase

Enact Phase allows you to shift some traits now and then reverse the effects later. It always lasts for a preset amount of turns. So if you shift a trait up for 2 rounds, after those two rounds those traits are shifted down for 2 rounds.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Duration (rounds)** - Any whole number from **1 to 5** (starts at 1). Cost: Free per round.
3. **Shift -/+** - Any whole number from **-6 to 6** (starts at 0). Cost: Free per step.
4. **Affected Trait(s)** - You start with **two** affected traits (Precision and Power) and may add or remove affected traits. Cost: Free per affected trait.

   For each affected trait, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

## Enact Reduction

Enact Reduction allows you to reduce the effect of an enactment you or someone else are the target for. It always has a Reduction Die which determines the reduction of the effect.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Reduction Mode** - Choose how the reduction is determined: a fixed die-size shift up/down the die ladder, or a die roll. Choose one of:
   - Die Shift, Die Roll
3. **Flat Bonus** - Any whole number from **0 to 8** (starts at 0). Cost: Free per +1.

## Enact Shift

Enact Shift allows you to temporarily enhance or weaken Traits. It always has a shift value ranging from -6 to 6, which decides how much and in what direction the shift happens.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Trait** - Choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic
3. **Shift -/+** - Any whole number from **-6 to 6** (starts at 0). Cost: Free per step.
4. **Uses** - Any whole number from **1 to 5** (starts at 1). Cost: Free per step.

## Enact Stack

Enact Stack (WIP)

## Area

**Area Interactions** encompass actions like bombs, splash potions, and traps. These interactions always have a defined **Radius** and **Range**:

*   **Radius**: This determines the area where the Enactment will take effect.
*   **Range**: This specifies how far from the user the point of origin is set. By default, the point of origin is 0m from the user.

You can also assign the point of **Origin** to an object, but this must be discussed with the GM beforehand. So you could put the point of **Origin** to an arrow or a device you’ve made. Then use a **Ranged Interaction** to throw it.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Radius** - Any whole number from **1 to 6** (starts at 1). Cost: Free per step.
3. **Range** - Choose one of:
   - 5m (Close), 25m (Medium), 50m (Long)

## Area of Effect

An **Area of Effect (AoE)** Interaction functions similarly to an Area Interaction, but its effects persist for several rounds. While an **Area Interaction** might be like a single-use bomb, an **AoE** Interaction is akin to a bomb that detonates every round. Alternatively, it could represent a healing circle, where characters gain health each round they remain within the **AoE**. The possibilities are endless, so get creative!

The effect of the **AoE** does not trigger immediately. Instead, it activates either at the start of a character's turn within the **AoE** or at the end of the **Engager**'s turn.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Radius** - Any whole number from **1 to 6** (starts at 1). Cost: Free per step.
3. **Range** - Choose one of:
   - 5m (Close), 25m (Medium), 50m (Long)
4. **Duration** - Any whole number from **2 to 6** (starts at 2). Cost: Free per round.

## Direct

Direct interactions are done by targeting those who are near you. They have to be within 1 meter of you in order for your enactment to execute.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Targets** - Any whole number from **1 to 5** (starts at 1). Cost: Free per step.

## Ranged

**Ranged** **Interactions** include actions like using bows, guns, and boomerangs. These interactions offer an increased range compared to **Direct** Interactions but come with a lower success rate due to a penalty on the **Engagement Roll**. Additionally, the target must not be obstructed or invisible to the **Engager** by default.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Range** - Choose one of:
   - 5m (Close), 25m (Medium), 50m (Long)
3. **Targets** - Any whole number from **1 to 5** (starts at 1). Cost: Free per step.

## Self

Self Interactions apply to your own character. They do require a validation still. But the Counter roll is a Generic Die instead. This means that you are still the Enagager and the DM makes the Counter Roll.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.

## Validations

Here, you'll find the guidelines and options for customizing your **Engagement** and **Counter Rolls**.

*   **Engagement Roll**: This is an Offensive Trait used to initiate actions against a target.
*   **Counter Roll**: This involves two Defensive Traits, allowing the target to choose how they respond to the attack.

**How to build it**

1. **Comment** *(optional)* - A note you can write on the ability. No cost.
2. **Engage** - Choose one of:
   - *Generic:* 1d4, 1d6, 1d8, 1d10, 1d12, 1d20
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic
3. **Counter Trait** - You start with **two** counter traits (Reflex and Constitution) and may add or remove counter traits. Cost: Free per counter trait.

   For each counter trait, choose one of:
   - *General:* Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic, Medicine
   - *Offense:* Precision, Power, Mind, Magic
   - *Defense:* Reflex, Constitution, Mind, Magic

# leveling
## Ability Builder Leveling

## Introduction

As you level up, your character gains a deeper understanding of their powers, techniques, and spells. This growth is represented by **Ability Points**. Ability Points are spent to pay the **Add Cost** of Perks, Enactments, Interactions, and Validations when constructing or upgrading your Abilities.

---

## Ability Points

At Level 1, a character starts with a base pool of Ability Points. As they level up, they gain a steady stream of new points, with larger spikes at milestone levels (Level 5 and Level 10).

These points are permanently invested into your abilities during character creation or level-ups.

### Upgrading Abilities

You do not need to create a brand new Ability every time you level up. You can spend your newly gained Ability Points to upgrade an existing Ability by adding new Perks, extending its Range, or attaching additional Enactments.

### Refunding Ability Points

Some Perks in the Ability Builder apply drawbacks or restrictions to an Ability (such as giving it an Item Dependency or increasing its Action Cost). These Perks have a **negative Add Cost**. Taking these drawbacks refunds Ability Points, allowing you to spend them elsewhere on the same Ability to make it more powerful

## Example Progression

If you build a simple "Fireball" at **Level 1**, you might spend 4 of your 10 starting points on it, leaving 6 points for a defensive Reaction ability.

By **Level 5**, you will have earned 11 additional Ability Points. You could spend 6 of those new points to add an Area of Effect Interaction to your Fireball and increase its damage dice, transforming it from a basic projectile into a massive explosion.

## Leveling Table: Ability Points

| Level | Points Gained | Total Ability Points |
| --- | --- | --- |
| **1** | +0 | 10 |
| **2** | +2 | 12 |
| **3** | +3 | 15 |
| **4** | +2 | 17 |
| **5** | +4 | 21 |
| **6** | +2 | 23 |
| **7** | +3 | 26 |
| **8** | +2 | 28 |
| **9** | +3 | 31 |
| **10** | +5 | 36 |