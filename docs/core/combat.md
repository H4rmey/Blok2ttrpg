# combat
## Combat

Combat works as most other ttrpg's, there is a type of grid where each square represents 1m. We have initiative rolls, actions, movement and all the other good stuff. Most of it is very basic so i won't go into to much detail. 

### Initiative

Rolling for initiative is done by rolling your perception + movement. PC's go before NPC's when equal values are rolled. PC's will discuess between themselfs when they roll an equal roll.

### Turns and Actions

On your turn you get three actions. By default you have {{ .Combat.Actions.Amount }} of actions. How much actions an perk costs can may differ. 

### Reactions

Your actions are spent on your turn, but a **Reaction** lets you act out of turn. A reaction is freeform: it is not built in advance and not listed on your sheet, you simply describe what you do and pay for it.

{{ reactionRules }}

The invoke point cost is what makes a reaction a real decision rather than a free bonus turn; see the [Invoking](#invoking) chapter for where those points come from. The energy cost means a reaction competes with your own perks, so a character who reacts every round has less left to spend when their turn comes around.

### Movement


Movement costs one action; it is fully allowed to just keep using actions just to move, however each subsequent movement action costs 1 energy extra (this stacks between turns).So moving 3 times in a row will cost 0 + 1 + 2 = 3 Energy.

### Energy and Recovery

Energy is the resource that using Perks spends. Each Enactment in an Perk costs 1 Energy, so a full turn of three actions costs roughly 3 Energy, and a Concentration costs a further point of upkeep every round it stays up. Repeated movement in a single stretch also drains Energy as described above.

Your maximum Energy comes from the Proficiency tier you bought for the Energy Vital Skill, and it climbs by 3 for every rung of that ladder. Read the column as rounds of sustained play: the bottom rung supports not quite two full turns, while the top supports nine or ten.

Energy does not come back on its own. On a rest you regain **{{ .Combat.EnergyRecoveryPerRest }} Energy**. That is deliberately less than a single fight consumes, so Energy is a resource to be managed across a whole day rather than reset between encounters. The GM may grant more for an especially long or comfortable rest.

### Attacking/Healing/Doing

Oppenents are not willing to get hit by your attacks/perks. That is why when attacking an opponent you make an **Attack Roll** to a **Target**. In the chapter about [Dice Rolling](dice-rolling.md) We already dicussed Engegement Rolls and Counter Rolls. An **Attack Roll** is a type of **Engegment Roll**.

When Attacking/Healing/Prepping/Anythinging you always first roll the Engagement Roll to see if you hit, then you resolve the action/enactment/thing.

