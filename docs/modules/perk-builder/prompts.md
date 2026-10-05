# AI Prompts

These prompts help you use an AI assistant (ChatGPT, Claude, Gemini, etc.) to generate valid YAML files you can import directly into Blok2. Before using any prompt, paste the contents of the **QuickStartGuide** into the same conversation so the AI understands the system.

> [!NOTE]
> The AI will scaffold the mechanical structure for you. You still need to review and write the flavour text — descriptions, backstory, and comments — yourself. Every placeholder is marked `TODO`.

\page

## Prompt 1 — Generate a Character YAML

Use this prompt to create a complete character import file. The result can be uploaded on the **Import Character** screen.

---

Copy and paste the following prompt into your AI chat after sharing the QuickStartGuide:

```
You are helping me create a Blok2 TTRPG character import file.

Using the QuickStartGuide I provided as context, generate a valid character YAML
that I can import into Blok2. Follow the schema below exactly.

Schema:
  id:        string  (optional – leave blank so the system assigns one)
  level:     integer (1 or higher)
  traits:    map of trait field keys to values  (strings, numbers, or booleans)
  skills:    map of "<group_id>.<Skill Name>" to proficiency id
             e.g. "offense.Strength: novice"
  perks:     list of perk objects (see perk schema below)
  packages:  list of installed-package objects (see package schema below)

Perk object schema:
  name:         string
  description:  "TODO: add description"
  type:         "execution" | "passive" | "reaction"
  tags:         list of strings
  fields:
    comment:       string
    energy_steps:  integer
    action_steps:  integer
  enactments:   list of enactment objects
    Each enactment:
      type:             "damage" | "effect" | "adjustment" | "nerf" | "buff"
      fields:           map of field keys to values (source, flat, damage_type, etc.)
      interaction:      "direct" | "self" | "aoe" | "line"
      interaction_data: map (range, targets, comment)
      validation_data:  map (engage die, counter_skill list, comment)

Installed-package object schema:
  package_id:  string
  name:        string
  description: "TODO: add description"
  perks:       list of perk objects (same schema as above)

Instructions:
1. Create a character concept of your choice that fits a fantasy setting.
2. Give the character a name, a level between 1 and 5, and fill in at least
   name, age, and a traits.backstory field set to "TODO: add backstory".
3. Assign 3-6 skills with appropriate proficiency ids (novice, apprentice, adept,
   expert, master).
4. Create 2-3 perks that match the character concept. For every description field
   write "TODO: add description". Build at least one enactment per perk.
5. Optionally include one package.
6. Output ONLY the raw YAML, no prose, no markdown fences.
```

> ##### After you receive the output
> - Replace every `TODO` field with your own flavour text.
> - Check that skill keys match the skill groups defined in your config.
> - Import the file on the **Characters** screen using the **Import** button.

\page

## Prompt 2 — Generate a Perk YAML

Use this prompt to create a single perk file you can import from the **Perk Library** screen or drop into `library/perks/<perk-name>/<perk-name>.yaml`.

---

```
You are helping me create a Blok2 TTRPG perk import file.

Using the QuickStartGuide I provided as context, generate a valid perk YAML
that I can import into Blok2. Follow the schema below exactly.

Top-level schema:
  name:        string  (human-readable, e.g. "Flame Strike")
  description: "TODO: add description"
  type:        "execution" | "passive" | "reaction"
  tags:        list of strings that categorise the perk (e.g. [damage, spell, fire])
  fields:
    comment:       string  (brief mechanical note, not flavour)
    energy_steps:  integer (0 = use config default)
    action_steps:  integer (0 = use config default)
  enactments:  list of enactment objects

Enactment schema:
  type:   "damage" | "effect" | "adjustment" | "nerf" | "buff" | "heal"
  fields:
    comment:     string
    source:      "<group>.<Skill>" or a die code like "d8"
    flat:        integer              (for damage/heal enactments)
    damage_type:                      (for damage enactments)
      - value: "Piercing" | "Bludgeoning" | "Slashing" | "Fire" | "Cold" | ...
  interaction:   "direct" | "self" | "aoe" | "line"
  interaction_data:
    comment: ""
    range:   string  (number of squares, e.g. "3"; omit for melee)
    targets: integer
  validation_data:
    comment: ""
    engage:       "d4" | "d6" | "d8" | "d10" | "d12"
    counter_skill:
      - value: "<group>.<Skill>"

For passive perks, replace enactments with:
  passives:
    - type:   "buff" | "nerf" | "adjustment"
      fields: map of passive field keys to values

For reaction perks, include a trigger field in fields:
  fields:
    trigger: string  (what event fires this reaction)

Instructions:
1. Invent a perk concept (weapon, spell, ability, or technique).
2. Name it and set type appropriately.
3. Add 1-3 tags that describe it.
4. Build 1-2 enactments (or passives for a passive perk).
5. Leave every description field as "TODO: add description".
6. Output ONLY the raw YAML, no prose, no markdown fences.
```

> ##### After you receive the output
> - Fill in the `description` field with your own flavour text.
> - Verify skill keys against your config's skill groups.
> - Import the file on the **Perk Library** screen using **Import Custom Perk**,
>   or save it to `library/perks/<perk-name>/<perk-name>.yaml` for a permanent entry.

\page

## Prompt 3 — Generate a Package YAML

Use this prompt to create a package — a named bundle of related perks that can be installed on a character together. Import it from the **Package Library** screen.

---

```
You are helping me create a Blok2 TTRPG package import file.

Using the QuickStartGuide I provided as context, generate a valid package YAML
that I can import into Blok2. Follow the schema below exactly.

Top-level schema:
  name:        string  (e.g. "Rogue Starter Kit")
  category:    string  (e.g. "combat", "magic", "utility", "starter")
  description: "TODO: add description"
  imports:     list of relative perk file paths OR inline perk objects

If you inline the perks (recommended for a single-file import), use this structure:

  perks:
    - name:        string
      description: "TODO: add description"
      type:        "execution" | "passive" | "reaction"
      tags:        list of strings
      fields:
        comment:       string
        energy_steps:  integer
        action_steps:  integer
      enactments:
        - type:   "damage" | "effect" | "adjustment" | "nerf" | "buff" | "heal"
          fields:
            comment:     string
            source:      "<group>.<Skill>" or die code
            flat:        integer
            damage_type:
              - value: string
          interaction:      "direct" | "self" | "aoe" | "line"
          interaction_data:
            comment: ""
            range:   string
            targets: integer
          validation_data:
            comment: ""
            engage:       die code
            counter_skill:
              - value: "<group>.<Skill>"

Instructions:
1. Choose a theme for the package (e.g. assassin, fire mage, paladin, ranger).
2. Name the package and pick an appropriate category.
3. Include 3-5 perks that fit the theme. Mix types (execution, passive, reaction)
   for variety.
4. Leave every description field as "TODO: add description".
5. Output ONLY the raw YAML, no prose, no markdown fences.
```

> ##### After you receive the output
> - Fill in `description` on the package and on each perk.
> - Import the file on the **Package Library** screen using **Import Custom Package**,
>   or save it under `library/packages/<category>/<id>/<id>.yaml` for a permanent entry.

\page
