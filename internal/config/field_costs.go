package config

import (
	"fmt"
)

// FieldValueMap is a generic map of field key -> submitted value.
// Values can be string, bool, int, float64, []string, or []any.
type FieldValueMap map[string]any

// EvaluateFieldCosts walks a fields schema against the supplied values and
// returns the total build/cast cost. It supports checkbox, dropdown, free_text,
// free_number, solutions, and states fields. Cascade dropdowns only charge
// for child fields belonging to the active option. Fields with a
// visibility_when are skipped unless the controlling field's value matches
// show_when.
func EvaluateFieldCosts(fields []FieldConfig, values FieldValueMap) (build int, cast int, err error) {
	for _, f := range fields {
		if !isFieldVisible(f, values) {
			continue
		}
		raw, ok := values[f.Key]
		if !ok {
			continue
		}
		b, c, ferr := evalField(f, raw, values)
		if ferr != nil {
			return 0, 0, fmt.Errorf("field %q: %w", f.Key, ferr)
		}
		build += b
		cast += c
	}
	return build, cast, nil
}

// isFieldVisible reports whether a field should be considered given the
// current values map. visibility_when names a sibling field whose value must
// match show_when for the field to be active.
func isFieldVisible(f FieldConfig, values FieldValueMap) bool {
	if f.VisibilityWhen == "" {
		return true
	}
	v, ok := values[f.VisibilityWhen]
	if !ok {
		return false
	}
	vs, _ := toString(v)
	return vs == showWhenString(f.ShowWhen)
}

func showWhenString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func evalField(f FieldConfig, raw any, all FieldValueMap) (int, int, error) {
	switch f.Type {
	case "checkbox":
		return evalCheckbox(f, raw)
	case "dropdown":
		return evalDropdown(f, raw, all)
	case "free_text":
		return 0, 0, nil
	case "free_number":
		return evalFreeNumber(f, raw)
	case "solutions":
		return evalSolutions(f, raw)
	case "states":
		return evalStates(f, raw, all)
	}
	return 0, 0, nil
}

func evalCheckbox(f FieldConfig, raw any) (int, int, error) {
	checked, err := toBool(raw)
	if err != nil {
		return 0, 0, err
	}
	if !checked {
		return 0, 0, nil
	}
	if f.Cost == nil {
		return 0, 0, nil
	}
	return f.Cost.AddCost, f.Cost.EnergyCost, nil
}

func evalDropdown(f FieldConfig, raw any, all FieldValueMap) (int, int, error) {
	val, err := toString(raw)
	if err != nil {
		return 0, 0, err
	}
	if val == "" {
		return 0, 0, nil
	}
	build, cast := 0, 0
	if f.Cost != nil {
		build += f.Cost.AddCost
		cast += f.Cost.EnergyCost
	}
	for _, opt := range f.Options {
		if opt.Value != val {
			continue
		}
		if opt.Cost != nil {
			build += opt.Cost.AddCost
			cast += opt.Cost.EnergyCost
		}
		childVals := optionChildValues(all, f.Key, opt.Value)
		cb, cc, ferr := EvaluateFieldCosts(opt.Fields, childVals)
		if ferr != nil {
			return 0, 0, ferr
		}
		build += cb
		cast += cc
		break
	}
	return build, cast, nil
}

// optionChildValues extracts values that belong to inline option child
// fields. The server receives these as separate form keys prefixed with the
// parent field key (e.g. "source__source_trait") and the JS sends them in a
// map nested under the parent key. We support both shapes.
func optionChildValues(all FieldValueMap, parentKey, optionValue string) FieldValueMap {
	if v, ok := all[parentKey+"__option"]; ok {
		if m, ok := v.(map[string]any); ok {
			if mv, ok := m[optionValue].(map[string]any); ok {
				return FieldValueMap(mv)
			}
		}
	}
	if v, ok := all[parentKey]; ok {
		if m, ok := v.(FieldValueMap); ok {
			return m
		}
		if m, ok := v.(map[string]any); ok {
			return FieldValueMap(m)
		}
	}
	prefixed := FieldValueMap{}
	prefix := parentKey + "__"
	for k, v := range all {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			prefixed[k[len(prefix):]] = v
		}
	}
	if len(prefixed) == 0 {
		return all
	}
	return prefixed
}

func evalFreeNumber(f FieldConfig, raw any) (int, int, error) {
	n, err := toInt(raw)
	if err != nil {
		return 0, 0, err
	}
	defVal := 0
	switch d := f.Default.(type) {
	case int:
		defVal = d
	case int64:
		defVal = int(d)
	case float64:
		defVal = int(d)
	}
	step := f.Step
	if step == 0 {
		step = 1
	}
	delta := n - defVal
	if delta == 0 {
		return 0, 0, nil
	}
	if f.PerStep == nil {
		return 0, 0, nil
	}
	var steps int
	if f.Rounding == "ceil" {
		steps = delta / step
		if delta%step != 0 && delta > 0 {
			steps++
		}
	} else if f.Rounding == "floor" {
		steps = delta / step
	} else {
		steps = delta / step
	}
	if steps == 0 {
		return 0, 0, nil
	}
	if steps > 0 {
		return steps * f.PerStep.Increase.AddCost, steps * f.PerStep.Increase.EnergyCost, nil
	}
	neg := -steps
	return neg * f.PerStep.Decrease.AddCost, neg * f.PerStep.Decrease.EnergyCost, nil
}

func evalSolutions(f FieldConfig, raw any) (int, int, error) {
	rows, err := solutionsRows(raw)
	if err != nil {
		return 0, 0, err
	}
	build, cast := 0, 0
	for _, row := range rows {
		cb, cc, ferr := EvaluateFieldCosts(f.RowFields, row)
		if ferr != nil {
			return 0, 0, ferr
		}
		build += cb
		cast += cc
	}
	if f.PerItem != nil {
		defaultCount := f.DefaultCount
		diff := len(rows) - defaultCount
		if diff > 0 {
			build += diff * f.PerItem.Increase.AddCost
			cast += diff * f.PerItem.Increase.EnergyCost
		} else if diff < 0 {
			neg := -diff
			build += neg * f.PerItem.Decrease.AddCost
			cast += neg * f.PerItem.Decrease.EnergyCost
		}
	}
	return build, cast, nil
}

// solutionsRows converts a submitted solutions value into a slice of
// FieldValueMap rows. The JS submits each solution row as a list of
// {value, type} objects, but a legacy form may submit a flat []string.
func solutionsRows(raw any) ([]FieldValueMap, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []FieldValueMap:
		return v, nil
	case []any:
		out := make([]FieldValueMap, 0, len(v))
		for _, item := range v {
			switch m := item.(type) {
			case FieldValueMap:
				out = append(out, m)
			case map[string]any:
				out = append(out, FieldValueMap(m))
			case string:
				if m == "" {
					continue
				}
				out = append(out, FieldValueMap{"value": m})
			default:
				return nil, fmt.Errorf("expected solution row, got %T", item)
			}
		}
		return out, nil
	case []string:
		out := make([]FieldValueMap, 0, len(v))
		for _, s := range v {
			if s == "" {
				continue
			}
			out = append(out, FieldValueMap{"value": s})
		}
		return out, nil
	}
	return nil, fmt.Errorf("expected list of solution rows, got %T", raw)
}

func evalStates(f FieldConfig, raw any, all FieldValueMap) (int, int, error) {
	rows, err := statesRows(raw)
	if err != nil {
		return 0, 0, err
	}
	build, cast := 0, 0
	for _, row := range rows {
		if isBlankStateRow(row) {
			continue
		}
		cb, cc, ferr := EvaluateFieldCosts(f.RowFields, row)
		if ferr != nil {
			return 0, 0, ferr
		}
		build += cb
		cast += cc
	}
	return build, cast, nil
}

func isBlankStateRow(row FieldValueMap) bool {
	kind, _ := toString(row["state_kind"])
	id, _ := toString(row["specific_state"])
	gen, _ := toString(row["general_state"])
	return kind == "" && id == "" && gen == ""
}

// EvaluateStatesWithSurcharge adds the additional-state surcharge for rows
// beyond the first when the surcharge cost definition is provided.
func EvaluateStatesWithSurcharge(surcharge CostDefinition, rowCount int) (int, int) {
	if rowCount <= 1 {
		return 0, 0
	}
	extra := rowCount - 1
	return extra * surcharge.AddCost, extra * surcharge.EnergyCost
}

// EvaluateStateRow computes the cost of a single Enact State row using the
// supplied state definitions. The row should contain a `state_kind` key
// ("specific" or "general") plus the matching `specific_state`,
// `general_state`, and optional `shift_amount`. The generalStateDefs and
// specificStateDefs slices come from StatesConfig.
func EvaluateStateRow(row FieldValueMap, generalStateDefs []GeneralStateConfig, specificStateDefs []SpecificStateConfig) (int, int) {
	kind, _ := toString(row["state_kind"])
	if kind == "general" {
		gs, _ := toString(row["general_state"])
		shift, _ := toInt(row["shift_amount"])
		for _, g := range generalStateDefs {
			if g.ID == gs {
				absShift := shift
				if absShift < 0 {
					absShift = -absShift
				}
				return absShift * g.ShiftCost.AddCost, absShift * g.ShiftCost.EnergyCost
			}
		}
		return 0, 0
	}
	id, _ := toString(row["specific_state"])
	for _, s := range specificStateDefs {
		if s.ID == id {
			return s.AddCost, s.EnergyCost
		}
	}
	return 0, 0
}

// ValidateSubmittedField ensures a value submitted for a field matches the
// field's declared constraints. It is used by the server to refuse out-of-
// bounds numbers, invalid dropdown values, etc., when the form is processed
// from the config schema. It recurses into the active option's child fields
// and into each state row.
func ValidateSubmittedField(f FieldConfig, raw any) error {
	if raw == nil {
		return nil
	}
	switch f.Type {
	case "dropdown":
		val, err := toString(raw)
		if err != nil {
			return fmt.Errorf("expected string, got %T", raw)
		}
		if val == "" {
			return nil
		}
		if len(f.Options) == 0 {
			return nil
		}
		var match *FieldOption
		for i, o := range f.Options {
			if o.Value == val {
				match = &f.Options[i]
				break
			}
		}
		if match == nil {
			return fmt.Errorf("invalid option %q for field %q", val, f.Key)
		}
		if all, ok := rawToMap(raw); ok {
			_ = all
		}
		// Recurse into active option's child fields using the prefixed
		// child map if the caller supplies one. The full validator
		// (ValidateSubmittedFields) recurses with the proper child map.
		return nil
	case "free_number":
		n, err := toInt(raw)
		if err != nil {
			return fmt.Errorf("expected int, got %T", raw)
		}
		if f.Min != 0 || f.Max != 0 {
			if n < f.Min || n > f.Max {
				return fmt.Errorf("field %q value %d out of bounds [%d, %d]", f.Key, n, f.Min, f.Max)
			}
		}
	case "checkbox":
		if _, err := toBool(raw); err != nil {
			return fmt.Errorf("field %q: %w", f.Key, err)
		}
	}
	return nil
}

// ValidateSubmittedFields walks a fields schema validating each value
// present in values against its field config. It also recurses into active
// dropdown option child fields and into each state row.
func ValidateSubmittedFields(fields []FieldConfig, values FieldValueMap) error {
	for _, f := range fields {
		if !isFieldVisible(f, values) {
			continue
		}
		raw, ok := values[f.Key]
		if !ok {
			continue
		}
		if err := ValidateSubmittedField(f, raw); err != nil {
			return err
		}
		if f.Type == "dropdown" {
			val, _ := toString(raw)
			for _, opt := range f.Options {
				if opt.Value == val && len(opt.Fields) > 0 {
					childVals := optionChildValues(values, f.Key, opt.Value)
					if err := ValidateSubmittedFields(opt.Fields, childVals); err != nil {
						return err
					}
				}
			}
		}
		if f.Type == "states" {
			rows, _ := statesRows(raw)
			for _, row := range rows {
				if isBlankStateRow(row) {
					continue
				}
				if err := ValidateSubmittedFields(f.RowFields, row); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func rawToMap(raw any) (FieldValueMap, bool) {
	switch v := raw.(type) {
	case FieldValueMap:
		return v, true
	case map[string]any:
		return FieldValueMap(v), true
	}
	return nil, false
}

func statesRows(raw any) ([]FieldValueMap, error) {
	switch v := raw.(type) {
	case []FieldValueMap:
		return v, nil
	case []any:
		out := make([]FieldValueMap, 0, len(v))
		for _, item := range v {
			if m, ok := item.(FieldValueMap); ok {
				out = append(out, m)
			} else if m, ok := item.(map[string]any); ok {
				out = append(out, FieldValueMap(m))
			} else {
				return nil, fmt.Errorf("expected state row map, got %T", item)
			}
		}
		return out, nil
	case FieldValueMap:
		if v == nil {
			return nil, nil
		}
		return []FieldValueMap{v}, nil
	case map[string]any:
		if v == nil {
			return nil, nil
		}
		return []FieldValueMap{v}, nil
	case nil:
		return nil, nil
	}
	return nil, fmt.Errorf("expected list of state rows, got %T", raw)
}

func toBool(raw any) (bool, error) {
	switch v := raw.(type) {
	case bool:
		return v, nil
	case string:
		switch v {
		case "on", "true", "1", "yes":
			return true, nil
		case "off", "false", "0", "no", "":
			return false, nil
		}
	}
	return false, fmt.Errorf("expected bool, got %T", raw)
}

func toString(raw any) (string, error) {
	switch v := raw.(type) {
	case string:
		return v, nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("expected string, got %T", raw)
}

func toInt(raw any) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
			return 0, fmt.Errorf("expected int, got %q", v)
		}
		return n, nil
	}
	return 0, fmt.Errorf("expected int, got %T", raw)
}

func toListLen(raw any) (int, error) {
	switch v := raw.(type) {
	case nil:
		return 0, nil
	case []string:
		// Count non-empty entries to mirror DOM behaviour.
		count := 0
		for _, s := range v {
			if s != "" {
				count++
			}
		}
		return count, nil
	case []any:
		count := 0
		for _, x := range v {
			if x == nil {
				continue
			}
			if s, ok := x.(string); ok && s == "" {
				continue
			}
			count++
		}
		return count, nil
	case string:
		if v == "" {
			return 0, nil
		}
		return 1, nil
	case int:
		return v, nil
	}
	return 0, fmt.Errorf("expected list, got %T", raw)
}
