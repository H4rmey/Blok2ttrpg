# AI Prompts

These prompts help you use an AI assistant (ChatGPT, Claude, Gemini, etc.) to generate valid YAML files you can import directly into Blok2. Before using any prompt, provide the documentation to the AI so it understands the ruleset mechanics and point budgets.

\page

## Prompt 1 — Generate a Character YAML

Use this prompt to create a complete, mechanically balanced character import file. The result can be uploaded on the **Import Character** screen.

Copy and paste the following prompt into your AI chat alongside the documentation:

```text
You are a character generator for the Blok2 TTRPG system. Using the provided Blok2 documentation, generate a complete, mechanically valid, and thematic character import YAML file.

CRITICAL OUTPUT RULE:
Output ONLY valid, raw YAML. Do NOT include any introductory text, explanations, math breakdowns, or closing remarks. Your entire response must be parseable directly as a YAML file.

CHARACTER CONCEPT:
[Insert your character concept, theme, and desired Level (1-10) here. If left blank, invent a compelling, grounded character at Level 1.]

MECHANICAL & BUDGET CONSTRAINTS (STRICT):
1. Level & Point Budgets:
   - Level must be between 1 and 10.
   - Skill Point Budget = 13 + 3 * (Level - 1). You MUST NOT exceed this budget.
   - Perk Point Budget = 20 + 4 * (Level - 1). Total build cost of all perks and packages MUST NOT exceed this budget.

2. Traits:
   - Follow the "Creating Effective Traits" guidelines: make traits specific, factual, non-vague, and actionable for invocations.
   - Populate Identity (`name`, `age`, `description`), Permanent Traits (`motivation`, `problem`, `trait_1`, `trait_2`, `trait_3`), and optionally Temporary Traits (`temp_trait_1`, etc.).

3. Skills & Proficiency IDs:
   - Every skill starts at `untrained` for free (0 points).
   - Cumulative Skill Point cost from `untrained`:
     - `inept`: -1 point (refunds 1 point)
     - `novice`: 1 point
     - `proficient`: 2 points
     - `expert`: 3 points
     - `master`: 4 points
     - `grandmaster`: 6 points
     - `champion`: 8 points
     - `legendary`: 11 points
   - Only include skills in the YAML map that deviate from `untrained`.
   - Remember to upgrade Vital Skills (`vital.HP`, `vital.Movement`, `vital.Energy`) so the character is survivable and has enough Energy to use their perks!
   - Valid Skill Keys (`<group>.<Skill>`):
     - `general.Stealth`, `general.Perception`, `general.Nature`, `general.Crafting`, `general.People Skill`, `general.Performance`, `general.Thievery`, `general.Knowledge`, `general.Magic`, `general.Medicine`, `general.Provoke`, `general.Intimidate`, `general.Athletics`, `general.Deceive`, `general.Resources`
     - `offense.Strength`, `offense.Precision`, `offense.Wisdom`, `offense.Magic`
     - `defense.Constitution`, `defense.Reflex`, `defense.Wisdom`, `defense.Magic`
     - `vital.HP`, `vital.Movement`, `vital.Energy`

4. Perks:
   - Create 2 to 4 cohesive perks that fit the character's traits and highest skills.
   - Valid Perk `type` values: `execution`, `concentration`, `reaction`, `passive`.
   - Valid Enactment `type` values: `damage`, `healing`, `condition`, `effect`, `motion`, `phase`, `shift`, `adjustment`, `resource`.
   - Valid Interaction `type` values: `self`, `direct`, `zone`.
   - Ensure the `engage` skill in `validation_data` and `source` in enactments match the skills the character is actually proficient in.
   - Write evocative, concise flavor text for all `description` fields and clear mechanical summaries for `comment` fields.

SCHEMA:
id: ""
level: <integer 1-10>
traits:
  name: <string>
  age: <integer or string>
  description: <string>
  motivation: <string>
  problem: <string>
  trait_1: <string>
  trait_2: <string>
  trait_3: <string>
  temp_trait_1: <string, optional>
skills:
  "<group>.<Skill>": <proficiency_id>
perks:
  - name: <string>
    description: <string>
    type: "execution" | "concentration" | "reaction" | "passive"
    tags: [<string>, ...]
    fields:
      comment: <string>
      energy_steps: <integer, default 0>
      action_steps: <integer, default 0>
      trigger: <string, only if type is reaction>
    enactments: # Omit if type is passive
      - type: "damage" | "healing" | "condition" | "effect" | "motion" | "phase" | "shift" | "adjustment" | "resource"
        fields:
          comment: <string>
          source: "<group>.<Skill>" | "1d4" | "1d6" | "1d8" | "1d10" | "1d12" | "1d12+1" | "1d12+2" | "1d12+3"
          # Include relevant enactment-specific fields from the documentation (e.g., condition, duration, distance, directions, skill, shift)
        interaction: "self" | "direct" | "zone"
        interaction_data:
          comment: <string>
          range: "1m" | "5m" | "25m" | "50m" # Omit for self
          targets: <integer 1-8> # For direct
          radius: <integer 1-6> # For zone
          duration: <integer 0-6> # For zone
        validation_data:
          comment: <string>
          engage: "<group>.<Skill>" | "1d4" | "1d6" | "1d8" | "1d10" | "1d12"
          counter_skill:
            - value: "<group>.<Skill>"
            - value: "<group>.<Skill>"
    passives: # Only if type is passive
      - type: "Resistance" | "Brutal Critical" | "Silver Tongue" | "Grudge Keeper" | "Practiced Hand" | "Sure Footed" | "Opportunist" | "Deep Pockets" | "Prepared Mind" | "Lingering Touch" | "Mobile Attacker" | "Conditional Enhancement"
        fields:
          # Relevant configuration fields for the chosen passive from the documentation
packages: []
```

\page

## Prompt 2 — Generate a Perk YAML

Use this prompt to create a single perk file you can import from the Perk Library screen or drop into library/perks/<perk-name>/<perk-name>.yaml.

```text
You are a perk designer for the Blok2 TTRPG system. Using the provided Blok2 documentation, generate a valid, mechanically sound perk import YAML file.

CRITICAL OUTPUT RULE:
Output ONLY valid, raw YAML. Do NOT wrap the output in markdown code blocks (no ```yaml or ```). Do NOT include any introductory text, explanations, cost calculations, or closing remarks. Your entire response must be parseable directly as a YAML file.

PERK CONCEPT:
[Insert your desired weapon, spell, ability, reaction, or passive concept here, along with a target Perk Point budget if any. If left blank, invent a creative tactical ability.]

MECHANICAL CONSTRAINTS (STRICT):
1. Perk Type: Must be one of `execution`, `concentration`, `reaction`, or `passive`.
   - If `reaction`, include a valid `trigger` in `fields` from the Reaction Triggers table (e.g., "A validation near you succeeds", "Someone moves away from you.", "Someone within range takes damage").
2. Enactments (for non-passive perks):
   - First enactment has 0 base build cost (plus option costs) and costs 1 Energy. Each additional enactment adds +1 build cost and +1 Energy.
   - Valid Enactment `type` values: `damage`, `healing`, `condition`, `effect`, `motion`, `phase`, `shift`, `adjustment`, `resource`.
   - Use only valid conditions from the Condition List (`Taunted`, `Swayed`, `Ignored`, `Confused`, `Vengeful`, `Distracted`, `Isolated`, `Charmed`, `Hypnotized`, `Stubborn`, `Paranoid`, `Insane`, `Stunned`, `Paralyzed`, `Pacified`, `Enraged`, `Disarmed`, `Silenced`, `Deafened`, `Stifled`, `Staggered`, `Prone`, `Anchored`, `Restrained`, `Terrified`, `Fragile`, `Cursed`, `Blessed`, `Hesitant`, `Fatigued`, `Energized`, `Delayed`, `Hastened`, `Echoed`, `Zombified`, `Linked`, `Incorporeal`, `Marked`, `Stealthed`).
3. Interactions:
   - Valid `interaction` values: `self`, `direct`, `zone`.
   - Valid ranges: `1m`, `5m`, `25m`, `50m`.
4. Validations:
   - Must use valid `<group>.<Skill>` keys (`general.*`, `offense.*`, `defense.*`) or generic dice (`1d4` through `1d12+3`).
5. Descriptions:
   - Write a clear, flavorful `description` and concise mechanical `comment` fields.

SCHEMA:
name: <string>
description: <string>
type: "execution" | "concentration" | "reaction" | "passive"
tags:
  - <string>
fields:
  comment: <string>
  energy_steps: <integer, default 0>
  action_steps: <integer, default 0>
  trigger: <string, only if type is reaction>
enactments: # Omit if type is passive
  - type: "damage" | "healing" | "condition" | "effect" | "motion" | "phase" | "shift" | "adjustment" | "resource"
    fields:
      comment: <string>
      source: "<group>.<Skill>" | "1d4" | "1d6" | "1d8" | "1d10" | "1d12" | "1d12+1" | "1d12+2" | "1d12+3"
      # Include specific fields required by the enactment type (e.g., condition, duration, solution_1, solution_2, solution_dc, distance, directions, skill, shift)
    interaction: "self" | "direct" | "zone"
    interaction_data:
      comment: <string>
      range: "1m" | "5m" | "25m" | "50m"
      targets: <integer 1-8, if direct>
      radius: <integer 1-6, if zone>
      duration: <integer 0-6, if zone>
    validation_data:
      comment: <string>
      engage: "<group>.<Skill>" | "1d4" | "1d6" | "1d8" | "1d10" | "1d12"
      counter_skill:
        - value: "<group>.<Skill>"
        - value: "<group>.<Skill>"
passives: # Include ONLY if type is passive
  - type: "Resistance" | "Brutal Critical" | "Silver Tongue" | "Grudge Keeper" | "Practiced Hand" | "Sure Footed" | "Opportunist" | "Deep Pockets" | "Prepared Mind" | "Lingering Touch" | "Mobile Attacker" | "Conditional Enhancement"
    fields:
      # Configure according to the Passive list in the documentation
```

\page

## Prompt 3 — Generate a Package YAML

Use this prompt to create a package — a named bundle of related perks (like an equipment loadout, martial style, or spell school) that can be installed on a character together. Import it from the Package Library screen.

```text
You are a content designer for the Blok2 TTRPG system. Using the provided Blok2 documentation, generate a valid package import YAML file containing a cohesive set of perks.

CRITICAL OUTPUT RULE:
Output ONLY valid, raw YAML. Do NOT wrap the output in markdown code blocks (no ```yaml or ```). Do NOT include any introductory text, explanations, or closing remarks. Your entire response must be parseable directly as a YAML file.

PACKAGE THEME:
[Insert your desired package theme (e.g., "Pyromancer Starter Kit", "Heavy Shield Bearer", "Shadow Blade") and target Perk Point budget here. If left blank, invent a cohesive 3-perk starter kit.]

MECHANICAL CONSTRAINTS (STRICT):
1. Include 3 to 5 perks that synergize mechanically and thematically (e.g., a resource generator + a resource spender, or a condition applier + a conditional payoff).
2. Mix perk types (`execution`, `concentration`, `reaction`, `passive`) where appropriate for the theme.
3. Use ONLY valid Blok2 enums from the documentation:
   - Enactment types: `damage`, `healing`, `condition`, `effect`, `motion`, `phase`, `shift`, `adjustment`, `resource`.
   - Interaction types: `self`, `direct`, `zone` (Ranges: `1m`, `5m`, `25m`, `50m`).
   - Skill keys: `general.<Skill>`, `offense.<Skill>`, `defense.<Skill>`, `vital.<Skill>`.
4. Write evocative flavor text for `description` fields and clear mechanical notes for `comment` fields.

SCHEMA:
name: <string>
category: "combat" | "magic" | "utility" | "starter" | "equipment"
description: <string>
perks:
  - name: <string>
    description: <string>
    type: "execution" | "concentration" | "reaction" | "passive"
    tags:
      - <string>
    fields:
      comment: <string>
      energy_steps: <integer, default 0>
      action_steps: <integer, default 0>
      trigger: <string, only if type is reaction>
    enactments: # Omit if type is passive
      - type: "damage" | "healing" | "condition" | "effect" | "motion" | "phase" | "shift" | "adjustment" | "resource"
        fields:
          comment: <string>
          source: "<group>.<Skill>" | "1d4" | "1d6" | "1d8" | "1d10" | "1d12" | "1d12+1" | "1d12+2" | "1d12+3"
        interaction: "self" | "direct" | "zone"
        interaction_data:
          comment: <string>
          range: "1m" | "5m" | "25m" | "50m"
          targets: <integer 1-8, if direct>
          radius: <integer 1-6, if zone>
          duration: <integer 0-6, if zone>
        validation_data:
          comment: <string>
          engage: "<group>.<Skill>" | "1d4" | "1d6" | "1d8" | "1d10" | "1d12"
          counter_skill:
            - value: "<group>.<Skill>"
            - value: "<group>.<Skill>"
    passives: # Include ONLY if type is passive
      - type: <Passive Name from documentation>
        fields: {}
```