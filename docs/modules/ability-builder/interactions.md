# Interactions

## Area

**Area Interactions** encompass actions like bombs, splash potions, and traps. These interactions always have a defined **Radius** and **Range**:

*   **Radius**: This determines the area where the Enactment will take effect.
*   **Range**: This specifies how far from the user the point of origin is set. By default, the point of origin is 0m from the user.

You can also assign the point of **Origin** to an object, but this must be discussed with the GM beforehand. So you could put the point of **Origin** to an arrow or a device you’ve made. Then use a **Ranged Interaction** to throw it.

{{buildGuide (interaction "area")}}

## Ranged

**Ranged** **Interactions** include actions like using bows, guns, and boomerangs. These interactions offer an increased range compared to **Direct** Interactions but come with a lower success rate due to a penalty on the **Engagement Roll**. Additionally, the target must not be obstructed or invisible to the **Engager** by default.

{{buildGuide (interaction "ranged")}}

## Direct

Direct interactions are done by targeting those who are near you. They have to be within 1 meter of you in order for your enactment to execute.

{{buildGuide (interaction "direct")}}

## Area of Effect

An **Area of Effect (AoE)** Interaction functions similarly to an Area Interaction, but its effects persist for several rounds. While an **Area Interaction** might be like a single-use bomb, an **AoE** Interaction is akin to a bomb that detonates every round. Alternatively, it could represent a healing circle, where characters gain health each round they remain within the **AoE**. The possibilities are endless, so get creative!

The effect of the **AoE** does not trigger immediately. Instead, it activates either at the start of a character's turn within the **AoE** or at the end of the **Engager**'s turn.

{{buildGuide (interaction "zone")}}

## Self

Self Interactions apply to your own character. They do require a validation still. But the Counter roll is a Generic Die instead. This means that you are still the Enagager and the DM makes the Counter Roll.

{{buildGuide (interaction "self")}}
