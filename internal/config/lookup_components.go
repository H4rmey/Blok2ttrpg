// Component lookups: resolving perk types, enactments and interactions by id, and applying the UI-only allowed/blocked filtering between them.
package config

// PerkType returns the perk-type component with the given id.
func (c *Config) PerkType(id string) (Component, bool) {
	if comp, ok := c.PerkTypes.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Enactment returns the enactment component with the given id.
func (c *Config) Enactment(id string) (Component, bool) {
	if comp, ok := c.Enactments.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Interaction returns the interaction component with the given id.
func (c *Config) Interaction(id string) (Component, bool) {
	if comp, ok := c.Interactions.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// filterByList applies the allow/block UI-filtering rule to an ordered id
// list. When allowed is non-empty only those ids (in their allowed order) are
// kept; otherwise when blocked is non-empty everything except the blocked ids
// is kept (in the input order); otherwise the input is returned unchanged.
// This is display-only and never enforced on save.
func filterByList(ids, allowed, blocked []string) []string {
	if len(allowed) > 0 {
		have := map[string]bool{}
		for _, id := range ids {
			have[id] = true
		}
		out := make([]string, 0, len(allowed))
		for _, id := range allowed {
			if have[id] {
				out = append(out, id)
			}
		}
		return out
	}
	if len(blocked) > 0 {
		block := map[string]bool{}
		for _, id := range blocked {
			block[id] = true
		}
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			if !block[id] {
				out = append(out, id)
			}
		}
		return out
	}
	return ids
}

// InteractionsFor returns the interaction components visible for the given
// enactment id, filtered by that enactment's allowed_interactions/
// blocked_interactions lists. Filtering is UI-only. An unknown enactment id
// yields the full interaction list.
func (c *Config) InteractionsFor(enactmentID string) []*Component {
	all := c.Interactions.List()
	comp, ok := c.Enactments.Get(enactmentID)
	if !ok {
		return all
	}
	ids := make([]string, 0, len(all))
	for _, ic := range all {
		ids = append(ids, ic.ID)
	}
	keep := filterByList(ids, comp.AllowedInteractions, comp.BlockedInteractions)
	out := make([]*Component, 0, len(keep))
	for _, id := range keep {
		if ic, ok := c.Interactions.Get(id); ok {
			out = append(out, ic)
		}
	}
	return out
}

// ValidationFieldsFor returns the validation fields visible for the given
// enactment id, filtered by that enactment's allowed_validations/
// blocked_validations lists (keyed by field key). Filtering is UI-only. An
// unknown enactment id yields the full validation field list. An enactment
// that opts into flat-DC validation shows the engage source plus the
// configured DC field instead of the counter_skill list, which is unused in
// that mode.
func (c *Config) ValidationFieldsFor(enactmentID string) []Field {
	all := c.Validations.Fields
	comp, ok := c.Enactments.Get(enactmentID)
	if !ok {
		return all
	}
	if comp.UsesDCValidation() {
		out := make([]Field, 0, len(all)+1)
		for _, f := range all {
			if f.Key == "counter_skill" {
				continue
			}
			out = append(out, f)
		}
		if dc := c.Validations.DCField(); dc != nil {
			out = append(out, *dc)
		}
		return out
	}
	ids := make([]string, 0, len(all))
	for _, f := range all {
		ids = append(ids, f.Key)
	}
	keep := filterByList(ids, comp.AllowedValidations, comp.BlockedValidations)
	keepSet := map[string]bool{}
	for _, k := range keep {
		keepSet[k] = true
	}
	// Preserve the config order of the surviving fields when using a block
	// list; use the allowed order when an allow list is set.
	if len(comp.AllowedValidations) > 0 {
		byKey := map[string]Field{}
		for _, f := range all {
			byKey[f.Key] = f
		}
		out := make([]Field, 0, len(keep))
		for _, k := range keep {
			if f, ok := byKey[k]; ok {
				out = append(out, f)
			}
		}
		return out
	}
	out := make([]Field, 0, len(all))
	for _, f := range all {
		if keepSet[f.Key] {
			out = append(out, f)
		}
	}
	return out
}

// EnactmentsFor returns the enactment components visible for the given
// perk-type id, filtered by that perk type's allowed_enactments/
// blocked_enactments lists. Filtering is UI-only. An unknown perk-type id
// yields the full enactment list.
func (c *Config) EnactmentsFor(perkTypeID string) []*Component {
	all := c.Enactments.List()
	comp, ok := c.PerkTypes.Get(perkTypeID)
	if !ok {
		return all
	}
	ids := make([]string, 0, len(all))
	for _, ec := range all {
		ids = append(ids, ec.ID)
	}
	keep := filterByList(ids, comp.AllowedEnactments, comp.BlockedEnactments)
	out := make([]*Component, 0, len(keep))
	for _, id := range keep {
		if ec, ok := c.Enactments.Get(id); ok {
			out = append(out, ec)
		}
	}
	return out
}

// ComponentByKind resolves a component id against the map named by kind. It

// backs the generic inline_builder feature so a dropdown can reference any
// enactment, interaction or perk type.
func (c *Config) ComponentByKind(kind, id string) (Component, bool) {
	switch kind {
	case "enactment":
		return c.Enactment(id)
	case "interaction":
		return c.Interaction(id)
	case "perk_type":
		return c.PerkType(id)
	default:
		return Component{}, false
	}
}
