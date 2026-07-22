# state
## Enact State

Enact State will apply a state to a target (e.g., prone, stunned, charmed). 

## Rules

*   **State**: Applies a condition to the target.

## Perks

{{perksTable (enactment "state")}}


## Template

Each state row picks either a Specific State or a General State (which shifts a
group of traits by an amount). Rows can add per-entry options such as an
Intensity or spreading to adjacent targets; see the Perks table above for the
cost of each choice.

```yaml
enactments:
  - type: Enact State
    states:
      - state_kind: specific        # or: general
        specific_state: <state id>  # when state_kind = specific
        # general_state: <state id> # when state_kind = general
        # shift_amount: <amount>    # when state_kind = general
        intensity: <minor|severe>   # optional per-entry option
        spreads: <true|false>       # optional per-entry option
```

