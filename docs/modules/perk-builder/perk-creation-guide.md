# Perk Creation Guide

Building a perk means defining its mechanics (timing, effect, target, and success condition). Flavor is entirely separate and free. You pay **Perk Points** to build the perk and **Energy** to use it.

## Step 0: Choose Flavor
A step that is not mechanical, but still important. First think about "What do i want my character to do?". Only then think about how you want to achieve it using this system. 

## Step 1: Choose Timing (Perk Type)
* **Execution:** Happens immediately (base cost).
* **Concentration:** Persists while maintained (initial cost + per-round energy upkeep).
* **Reaction:** Fires out of turn on a fixed trigger (premium build and energy cost).

## Step 2: Choose Effects (Enactments)
Pick modular blocks that define *what* happens (examples: Damage, Healing, Motion, Condition, Effect, Shift, Phase, Adjustment, Resource). 
* The first Enactment costs 0 build points and 1 Energy.
* Each additional Enactment costs 1 build point and adds 1 Energy to the cast cost.

## Step 3: Choose Targets (Interactions)
Determine *who* is affected: **Self**, **Direct** (specific targets at range), or **Zone** (area radius, range, duration).
* **Inheritance:** Chained enactments automatically target the same creature as the one before it. No need to revalidate.
* **Splitting:** You can specify a different target for subsequent enactments (e.g., "damage them, heal me"), but this requires a new Interaction and Validation roll, increasing the cost.
* **Cost Control:** 1m range is free. Extra range and zone radius cost points.

## Step 4: Resolution (Validations)
Perks resolve step-by-step:
1. **Interaction:** Declare who is targeted.
2. **Validation:** You make an Enactment Roll; targets make Counter Rolls. 
3. **Enactment:** The effect applies to those who fail their Counter Roll. 
