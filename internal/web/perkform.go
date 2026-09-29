// This file holds the form decoding for the perk builder: turning a posted
// request into the loosely-typed field values the cost engine consumes. It is
// deliberately separate from the handlers so request parsing can be read and
// tested without the surrounding HTTP plumbing.

package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// readFieldValues extracts values for a set of fields from a form using a key
// prefix. It handles each field type generically.
func readFieldValues(cfg *config.Config, fields []config.Field, prefix string, r *http.Request) map[string]any {
	out := map[string]any{}
	for _, f := range fields {
		name := prefix + f.Key
		switch f.Type {
		case "checkbox":
			out[f.Key] = r.FormValue(name) == "on" || r.FormValue(name) == "true"
		case "free_number":
			n, _ := strconv.Atoi(r.FormValue(name))
			out[f.Key] = n
		case "multiselect", "conditions":
			out[f.Key] = readRowValues(f, name, r)

		case "condition_select":
			out[f.Key] = r.FormValue(name)
			// The per-condition shift dropdown is injected by HTMX only for general
			// conditions, so it is not a standalone config field. Capture its posted
			// value under the sibling shift key so the cost engine can read it.
			shiftKey := f.ShiftKey
			if shiftKey == "" {
				shiftKey = "shift_amount"
			}
			if v := r.FormValue(prefix + shiftKey); v != "" {
				n, _ := strconv.Atoi(v)
				out[shiftKey] = n
			}
		case "dropdown":

			val := r.FormValue(name)
			// A dropdown must always carry a real value. If the post is missing
			// one (e.g. a region that was not rendered), fall back to the
			// configured default and then to the first option, matching what
			// normalization stores and what the builder renders.
			if val == "" {
				val = asStringValue(f.Default)
			}
			if val == "" && cfg != nil {
				for _, opt := range cfg.ResolveOptions(f) {
					if opt.Value != "" {
						val = opt.Value
						break
					}
				}
			}
			out[f.Key] = val
			// An inline_builder dropdown carries a nested component builder.
			// Parse the referenced component's fields under "<name>_ib_" and
			// store them under a nested key so the cost engine can recurse.
			if f.InlineBuilder != nil && val != "" && cfg != nil {
				if comp, ok := cfg.ComponentByKind(f.InlineBuilder.Kind, val); ok {
					out[f.Key+"_ib"] = readFieldValues(cfg, comp.Fields, name+"_ib_", r)
				}
			}
		default:
			out[f.Key] = r.FormValue(name)
		}

	}
	return out
}

// asStringValue renders a config default as a string ("" for nil).
func asStringValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// readRowValues reads a multiselect/conditions repeatable field from the form. Rows

// are posted with an index in the name, e.g. "<name>_0_type", "<name>_0_value".
// The posted "<name>_count" holds the number of rows.
func readRowValues(f config.Field, name string, r *http.Request) []map[string]any {
	count, _ := strconv.Atoi(r.FormValue(name + "_count"))
	rows := make([]map[string]any, 0, count)
	for i := 0; i < count; i++ {
		rowPrefix := fmt.Sprintf("%s_%d_", name, i)
		row := map[string]any{}
		empty := true
		for _, rf := range f.RowFields {
			v := r.FormValue(rowPrefix + rf.Key)
			row[rf.Key] = v
			if v != "" {
				empty = false
			}
		}
		if !empty {
			rows = append(rows, row)
		}
	}
	return rows
}
