# combat
## Combat

Combat works as most other ttrpg's, there is a type of grid where each square represents 1m. We have initiative rolls, actions, movement and all the other good stuff. Most of it is very basic so i won't go into to much detail. 

### Initiative

Rolling for initiative is done by rolling your perception + movement. PC's go before NPC's when equal values are rolled. PC's will discuess between themselfs when they roll an equal roll.

### Turns and Actions

On your turn you get three actions. By default you have {{ .Combat.Actions.Amount }} of actions. How much actions an ability costs can may differ. 

### Movement

Movement costs one action; it is fully allowed to just keep using actions just to move, however each subsequent movement action costs 1 energy extra (this stacks between turns).So moving 3 times in a row will cost 0 + 1 + 2 = 3 Energy.

### Attacking/Healing/Doing

Oppenents are not willing to get hit by your attacks/abilities. That is why when attacking an opponent you make an **Attack Roll** to a **Target**. In the chapter about [Dice Rolling](dice-rolling.md) We already dicussed Engegement Rolls and Counter Rolls. An **Attack Roll** is a type of **Engegment Roll**.

When Attacking/Healing/Prepping/Anythinging you always first roll the Engagement Roll to see if you hit, then you resolve the action/enactment/thing.

