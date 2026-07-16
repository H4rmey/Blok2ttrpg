# state
## Enact State

Enact State abilities apply a state or condition to a target (e.g., prone, stunned, poisoned). This enactment type is currently a work in progress.

## Conditions

**Conditions** can either boost or limit your character. They are a collection of nerf/buffs to your **Traits**. They always have a value ranging from -6 to +6 representing **Die Shifts** in your **Traits**. So each Conditions shifts x amount of traits in your character, this can either be temporary or permanent.

> **Example:** Your character gets **Blinded** by a flash of light because you failed a **Counter Roll**. The DM tells you that you are now **Blinded** -2. You now have -2 **Die Shift** on **Offensive Presicion Rolls** and **Defensive Reflex Rolls**

> [!NOTE]
> As the number and type of **Traits** can differ between games, the DM or group may need to tweak what a condition applies to. 

I won't list what **Conditions** do exactly what, because that highly depends on the scenario. 

> **Example:** You are Frightened can be either you are scared in the dark or you have are afraid of public speaking.

> **Example:** Being Encumbered might is a condition that applies to all movement type skills but also strength. As you carry to much you also do not have any strength left to lift anything else, it will therefor also impact your Offensive power stat.

Some conditions do have a specific definition, rulesets or traits that they apply to. Those will be listed seperatly below

### List of General Conditions

*   Blinded
*   Deafened
*   Broken
*   Clumsy
*   Concealed
*   Confused
*   Controlled
*   Dazzled
*   Doomed
*   Drained
*   Dying
*   Encumbered
*   Enfeebled
*   Fascinated
*   Fatigued
*   Flat-Footed
*   Fleeing
*   Friendly
*   Frightened
*   Grabbed
*   Helpful
*   Hidden
*   Hostile
*   Immobilized
*   Indifferent
*   Invisible
*   Observed
*   Paralyzed
*   Persistent Damage
*   Petrified
*   Quickened
*   Restrained
*   Sickened
*   Slowed
*   Stunned
*   Stupefied
*   Unconscious
*   Undetected
*   Unfriendly
*   Unnoticed
*   Wounded

## States

A **State** is not something that gets applied directly. Your character is constantly switching between differents **States**. Generally we use **States** to know what a character is doing. 

> **Example:** Your character is running towards a Target. Because you are running you are automatically harder to hit. The DM now rules that because of that anyone that is trying to hit you now get a **Shift** of -1 to their **precision** checks. 

Below are listed some states with their general rules. But as ttrpg's are flexible systems make sure you rule it however you see fit.

### Movement States

*   Standing
*   Walking
*   Running
*   Sitting
*   Prone
*   Jumping
*   Airborne/Flying
*   Falling
*   Spinning
*   Dancing

### Activity States

*   Building/Crafting
*   Blocking
*   Healing
*   Damaging
*   Singing
*   Acting
*   Stealthing
and more to list

## Rules

*   **State**: Applies a condition to the target.

## Perks

{{.StatePerksTable}}

## Template

```yaml
enactments:
  - type: Enact State
    state: <state here>
    is_optional: False
    base_enactment_energy_cost: 0
    perks:
      - description: <Perk Description>
        add_cost: <Cost>
        amount: <Amount>
        total_add_cost: <Total Cost>
        energy_cost: <Total Cost Energy>
        is_optional: <True/False>
```
