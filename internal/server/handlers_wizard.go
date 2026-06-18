package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/blok2ttrpg/charsheet/internal/gamedata"
	"github.com/blok2ttrpg/charsheet/internal/models"
)

// WizardVM is the view model for the ability edit wizard.
type WizardVM struct {
	AbilityIndex    int
	Ability         *models.Ability
	TypePerks       []gamedata.PerkDef
	Enactments      []EnactmentVM
	CompatibleEnact []gamedata.EnactmentDef
	Triggers        []gamedata.TriggerDef
	Knockouts       []string
	PointsAvailable int
	PointsTotal     int
	CostSummary     CostSummaryVM
	AllPerks        []PerkSummaryRow
	Overview        string // Human-readable description of what the ability does
	QuickSummary    string // One-line quick reference
	OpenSection     string // Which details section to keep open after perk action
}

// CostSummaryVM holds the cost breakdown for the ability.
type CostSummaryVM struct {
	TotalAddCost       int
	TotalEnergyCost    int
	OptionalEnergyCost int // Energy that can be saved by skipping optional perks
	MinEnergyCost      int // Minimum energy cost (with all optionals skipped)
}

// PerkSummaryRow represents a perk in the aggregate summary.
type PerkSummaryRow struct {
	Source      string // "Type", "Enactment: Enact Damage", "Interaction: Ranged", "Validation"
	Description string
	AddCost     int
	EnergyCost  int
	IsOptional  bool
	// For removal identification
	Target    string
	PerkIndex int
}

// EnactmentVM represents an enactment in the wizard.
type EnactmentVM struct {
	Index            int
	Enactment        *models.Enactment
	IsFirst          bool
	EnergyCost       int
	AddCost          int
	ComputedDice     string   // Effective dice after perk modifications
	ComputedSummary  []string // Multi-line computed summary based on perks
	InteractionSummary []string // Computed interaction details reflecting perks
	ValidationSummary []string // Computed validation details reflecting perks
	InheritedInteraction bool // True if interaction is inherited from previous
	HasGenericEngageRoll bool // For perk prerequisites
	HasGenericCounterRoll bool // For perk prerequisites
	HasRemovedCounterRoll bool // Has "Remove one Counter Roll option" perk
	Perks            []gamedata.PerkDef
	InteractionPerks []gamedata.PerkDef
	ValidationPerks  []gamedata.PerkDef
	InteractionTypes []string
	OffensiveTraits  []string
	DefensiveTraits  []string
}

func (s *Server) handleAbilityEdit(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	vm := s.buildWizardVM(char, index)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardAddEnactment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	indexStr := r.FormValue("ability_index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	enactmentType := r.FormValue("enactment_type")
	ability := &char.Abilities[index]

	// Determine base cost (first enactment is free)
	isFirst := len(ability.Enactments) == 0
	baseCost := 0
	if !isFirst {
		// Find cost from catalog
		for _, e := range gamedata.CompatibleEnactments(string(ability.Type)) {
			if e.Type == enactmentType {
				baseCost = e.EnergyCost
				break
			}
		}
	}

	newEnactment := models.Enactment{
		Type:                    models.EnactmentType(enactmentType),
		IsOptional:              !isFirst,
		BaseEnactmentEnergyCost: baseCost,
	}

	// Set defaults based on enactment type
	switch models.EnactmentType(enactmentType) {
	case models.EnactmentDamage:
		newEnactment.DamageDice = "1d4"
	case models.EnactmentHealing:
		newEnactment.HealingDice = "1d4"
	case models.EnactmentMovement:
		newEnactment.MinimalDistance = "1m"
		newEnactment.Origin = "Engager"
		newEnactment.DirectionOptions = []string{"Away"}
	case models.EnactmentProficiencyShift:
		newEnactment.ShiftDirection = "UP"
		newEnactment.ShiftAmount = 1
		newEnactment.ShiftUses = 1
	case models.EnactmentPersistentEffect:
		newEnactment.Duration = "2 rounds"
		newEnactment.TriggerTiming = "Start of Target's Turn"
		newEnactment.Solutions = []string{"Dexterity", "Constitution"}
	}

	ability.Enactments = append(ability.Enactments, newEnactment)

	vm := s.buildWizardVM(char, index)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardRemoveEnactment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	abilityStr := r.URL.Query().Get("ability_index")
	abilityIndex, err := strconv.Atoi(abilityStr)
	if err != nil || abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	enactStr := r.URL.Query().Get("enactment_index")
	enactIndex, err := strconv.Atoi(enactStr)
	if err != nil {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	if enactIndex < 0 || enactIndex >= len(ability.Enactments) {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	ability.Enactments = append(ability.Enactments[:enactIndex], ability.Enactments[enactIndex+1:]...)

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardAddPerk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.FormValue("ability_index"))
	perkTarget := r.FormValue("perk_target") // "ability", "enactment-0", "enactment-1", etc.
	perkDesc := r.FormValue("perk_description")
	perkAddCost, _ := strconv.Atoi(r.FormValue("perk_add_cost"))
	perkEnergyCost, _ := strconv.Atoi(r.FormValue("perk_energy_cost"))
	isOptionalStr := r.FormValue("perk_is_optional")
	perkExtra := r.FormValue("perk_extra")

	// Append extra input to description if provided
	if perkExtra != "" {
		perkDesc = perkDesc + ": " + perkExtra
	}

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	newPerk := models.Perk{
		Description:  perkDesc,
		AddCost:      perkAddCost,
		Amount:       1,
		TotalAddCost: perkAddCost,
		EnergyCost:   perkEnergyCost,
		IsOptional:   isOptionalStr == "true",
	}

	// Check budget (only if add cost is positive)
	if perkAddCost > 0 && perkAddCost > char.AbilityPointsAvailable() {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "insufficient ability points (need %d, have %d)"}`, perkAddCost, char.AbilityPointsAvailable()))
		vm := s.buildWizardVM(char, abilityIndex)
		s.renderTemplate(w, "ability_wizard", vm)
		return
	}

	if perkTarget == "ability" {
		ability.Perks = append(ability.Perks, newPerk)
	} else if len(perkTarget) > 10 && perkTarget[:9] == "enactment" {
		enactIdx, _ := strconv.Atoi(perkTarget[10:])
		if enactIdx >= 0 && enactIdx < len(ability.Enactments) {
			ability.Enactments[enactIdx].Perks = append(ability.Enactments[enactIdx].Perks, newPerk)
		}
	}

	vm := s.buildWizardVM(char, abilityIndex)
	vm.OpenSection = perkTarget // Keep the section open that was just interacted with
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardRemovePerk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.URL.Query().Get("ability_index"))
	perkTarget := r.URL.Query().Get("perk_target")
	perkIndex, _ := strconv.Atoi(r.URL.Query().Get("perk_index"))

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]

	if perkTarget == "ability" {
		if perkIndex >= 0 && perkIndex < len(ability.Perks) {
			ability.Perks = append(ability.Perks[:perkIndex], ability.Perks[perkIndex+1:]...)
		}
	} else if len(perkTarget) > 10 && perkTarget[:9] == "enactment" {
		enactIdx, _ := strconv.Atoi(perkTarget[10:])
		if enactIdx >= 0 && enactIdx < len(ability.Enactments) {
			e := &ability.Enactments[enactIdx]
			if perkIndex >= 0 && perkIndex < len(e.Perks) {
				e.Perks = append(e.Perks[:perkIndex], e.Perks[perkIndex+1:]...)
			}
		}
	}

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardBack(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}
	vm := BuildAbilityListVM(char)
	s.renderTemplate(w, "tab_abilities", vm)
}

func (s *Server) buildWizardVM(char *models.Character, index int) WizardVM {
	ability := &char.Abilities[index]
	typePerks := gamedata.AbilityTypePerks(string(ability.Type))
	compatEnact := gamedata.CompatibleEnactments(string(ability.Type))

	enactments := make([]EnactmentVM, len(ability.Enactments))
	for i := range ability.Enactments {
		perks := gamedata.EnactmentPerks(string(ability.Enactments[i].Type))

		// Get interaction perks if an interaction is set
		var interactionPerks []gamedata.PerkDef
		var validationPerks []gamedata.PerkDef
		hasOwnInteraction := len(ability.Enactments[i].Interactions) > 0
		if hasOwnInteraction {
			interType := string(ability.Enactments[i].Interactions[0].Type)
			if interType != "Inherited" {
				interactionPerks = gamedata.InteractionPerks(interType)
			}
			validationPerks = gamedata.ValidationPerks()
		} else {
			validationPerks = gamedata.ValidationPerks()
		}

		// Check validation perk prerequisites
		hasGenericEngage := false
		hasGenericCounter := false
		hasRemovedCounter := false
		if hasOwnInteraction && ability.Enactments[i].Interactions[0].Validation != nil {
			for _, vp := range ability.Enactments[i].Interactions[0].Validation.Perks {
				if contains(vp.Description, "Engage Roll") && contains(vp.Description, "Generic") {
					hasGenericEngage = true
				}
				if contains(vp.Description, "Counter Roll") && contains(vp.Description, "Generic") {
					hasGenericCounter = true
				}
				if contains(vp.Description, "Remove one Counter Roll") {
					hasRemovedCounter = true
				}
			}
		}

		// Filter validation perks based on prerequisites
		filteredValPerks := filterValidationPerks(validationPerks, hasGenericEngage, hasGenericCounter)

		enactments[i] = EnactmentVM{
			Index:                 i,
			Enactment:             &ability.Enactments[i],
			IsFirst:               i == 0,
			EnergyCost:            ability.Enactments[i].BaseEnactmentEnergyCost,
			ComputedDice:          computeEffectiveDice(&ability.Enactments[i]),
			ComputedSummary:       computeEnactmentSummary(&ability.Enactments[i]),
			InteractionSummary:    computeInteractionSummary(&ability.Enactments[i]),
			ValidationSummary:     computeValidationSummary(&ability.Enactments[i]),
			InheritedInteraction:  !hasOwnInteraction && i > 0,
			HasGenericEngageRoll:  hasGenericEngage,
			HasGenericCounterRoll: hasGenericCounter,
			HasRemovedCounterRoll: hasRemovedCounter,
			Perks:                 perks,
			InteractionPerks:      interactionPerks,
			ValidationPerks:       filteredValPerks,
			InteractionTypes:      gamedata.AllInteractionTypes(),
			OffensiveTraits:       gamedata.OffensiveTraits(),
			DefensiveTraits:       gamedata.DefensiveTraits(),
		}
	}

	var triggers []gamedata.TriggerDef
	if ability.Type == models.AbilityTypeReaction {
		triggers = gamedata.ReactionTriggers()
	}

	var knockouts []string
	if ability.Type == models.AbilityTypePhase {
		knockouts = gamedata.KnockoutRequirements()
	}

	overview, quickSummary := buildAbilityOverview(ability)

	return WizardVM{
		AbilityIndex:    index,
		Ability:         ability,
		TypePerks:       typePerks,
		Enactments:      enactments,
		CompatibleEnact: compatEnact,
		Triggers:        triggers,
		Knockouts:       knockouts,
		PointsAvailable: char.AbilityPointsAvailable(),
		PointsTotal:     char.TotalAbilityPoints(),
		CostSummary:     buildCostSummary(ability),
		AllPerks:        buildAllPerks(ability),
		Overview:        overview,
		QuickSummary:    quickSummary,
	}
}

func buildCostSummary(a *models.Ability) CostSummaryVM {
	totalAdd := a.TotalAddCost()
	totalEnergy := a.TotalEnergyCost()

	// Calculate optional energy (energy from optional perks that can be skipped)
	optionalEnergy := 0
	for _, p := range a.Perks {
		if p.IsOptional {
			optionalEnergy += p.EnergyCost
		}
	}
	for _, e := range a.Enactments {
		for _, p := range e.Perks {
			if p.IsOptional {
				optionalEnergy += p.EnergyCost
			}
		}
		for _, inter := range e.Interactions {
			for _, p := range inter.Perks {
				if p.IsOptional {
					optionalEnergy += p.EnergyCost
				}
			}
			if inter.Validation != nil {
				for _, p := range inter.Validation.Perks {
					if p.IsOptional {
						optionalEnergy += p.EnergyCost
					}
				}
			}
		}
	}

	return CostSummaryVM{
		TotalAddCost:       totalAdd,
		TotalEnergyCost:    totalEnergy,
		OptionalEnergyCost: optionalEnergy,
		MinEnergyCost:      totalEnergy - optionalEnergy,
	}
}

func buildAllPerks(a *models.Ability) []PerkSummaryRow {
	var rows []PerkSummaryRow

	// Type-level perks
	for i, p := range a.Perks {
		rows = append(rows, PerkSummaryRow{
			Source:      "Type",
			Description: p.Description,
			AddCost:     p.TotalAddCost,
			EnergyCost:  p.EnergyCost,
			IsOptional:  p.IsOptional,
			Target:      "ability",
			PerkIndex:   i,
		})
	}

	// Enactment perks
	for ei, e := range a.Enactments {
		for pi, p := range e.Perks {
			rows = append(rows, PerkSummaryRow{
				Source:      fmt.Sprintf("Enact: %s", e.Type),
				Description: p.Description,
				AddCost:     p.TotalAddCost,
				EnergyCost:  p.EnergyCost,
				IsOptional:  p.IsOptional,
				Target:      fmt.Sprintf("enactment-%d", ei),
				PerkIndex:   pi,
			})
		}
		// Interaction perks
		for _, inter := range e.Interactions {
			for pi, p := range inter.Perks {
				rows = append(rows, PerkSummaryRow{
					Source:      fmt.Sprintf("Inter: %s", inter.Type),
					Description: p.Description,
					AddCost:     p.TotalAddCost,
					EnergyCost:  p.EnergyCost,
					IsOptional:  p.IsOptional,
					Target:      fmt.Sprintf("interaction-%d", ei),
					PerkIndex:   pi,
				})
			}
			if inter.Validation != nil {
				for pi, p := range inter.Validation.Perks {
					rows = append(rows, PerkSummaryRow{
						Source:      "Validation",
						Description: p.Description,
						AddCost:     p.TotalAddCost,
						EnergyCost:  p.EnergyCost,
						IsOptional:  p.IsOptional,
						Target:      fmt.Sprintf("validation-%d", ei),
						PerkIndex:   pi,
					})
				}
			}
		}
	}

	return rows
}

func computeEffectiveDice(e *models.Enactment) string {
	baseDice := e.DamageDice
	if baseDice == "" {
		baseDice = e.HealingDice
	}
	if baseDice == "" {
		return ""
	}

	// Check if dice was replaced by a trait
	traitReplace := ""
	tierShifts := 0
	flatBonus := 0
	offensiveTrait := ""
	medicineTrait := false
	for _, p := range e.Perks {
		desc := p.Description
		if contains(desc, "Change Damage Dice to one of your traits") || contains(desc, "Change Heal Dice to one of your traits") {
			// Extract the trait name after ": "
			idx := 0
			for i := 0; i <= len(desc)-2; i++ {
				if desc[i] == ':' && desc[i+1] == ' ' {
					idx = i + 2
					break
				}
			}
			if idx > 0 && idx < len(desc) {
				traitReplace = desc[idx:]
			}
		}
		if (contains(desc, "Shift Dice Tier") && contains(desc, "up")) {
			tierShifts += p.Amount
		}
		if contains(desc, "flat +1 bonus") {
			flatBonus += p.Amount
		}
		if contains(desc, "Offensive Trait Dice") {
			// Extract trait after ": " if present
			idx := 0
			for i := 0; i <= len(desc)-2; i++ {
				if desc[i] == ':' && desc[i+1] == ' ' {
					idx = i + 2
					break
				}
			}
			if idx > 0 && idx < len(desc) {
				offensiveTrait = desc[idx:]
			} else {
				offensiveTrait = "Offensive Trait"
			}
		}
		if contains(desc, "Medicine Trait Dice") {
			medicineTrait = true
		}
	}

	var result string
	if traitReplace != "" {
		result = traitReplace
	} else {
		// Dice progression
		diceTiers := []string{"1d4", "1d6", "1d8", "1d10", "1d12", "1d20"}
		currentIdx := 0
		for i, d := range diceTiers {
			if baseDice == d || baseDice == d[1:] {
				currentIdx = i
				break
			}
		}
		newIdx := currentIdx + tierShifts
		if newIdx >= len(diceTiers) {
			newIdx = len(diceTiers) - 1
		}
		if newIdx < 0 {
			newIdx = 0
		}
		result = diceTiers[newIdx]
	}

	if flatBonus > 0 {
		result += fmt.Sprintf(" + %d", flatBonus)
	}
	if offensiveTrait != "" {
		result += " + " + offensiveTrait
	}
	if medicineTrait {
		result += " + Medicine"
	}
	return result
}

func computeEnactmentSummary(e *models.Enactment) []string {
	var lines []string

	switch e.Type {
	case models.EnactmentDamage:
		dice := computeEffectiveDice(e)
		lines = append(lines, fmt.Sprintf("Damage: %s", dice))
		for _, p := range e.Perks {
			if contains(p.Description, "Will Always Resolve") || contains(p.Description, "Will always resolve") {
				lines = append(lines, "Always resolves")
			}
			if contains(p.Description, "Change Damage Dice to one of your traits") {
				lines = append(lines, "Damage dice = one of your traits")
			}
		}
	case models.EnactmentHealing:
		dice := computeEffectiveDice(e)
		lines = append(lines, fmt.Sprintf("Healing: %s", dice))
		for _, p := range e.Perks {
			if contains(p.Description, "Will Always Resolve") || contains(p.Description, "Will always resolve") {
				lines = append(lines, "Always resolves")
			}
			if contains(p.Description, "Medicine Trait Dice") {
				lines = append(lines, "+ Medicine trait dice")
			}
		}
	case models.EnactmentMovement:
		// Compute effective distance
		extraMeters := 0
		usesTraitDist := false
		originChanged := false
		for _, p := range e.Perks {
			if contains(p.Description, "Add 1m") {
				extraMeters += p.Amount
			}
			if contains(p.Description, "Change total movement to any other trait") {
				usesTraitDist = true
			}
			if contains(p.Description, "Change Origin") {
				originChanged = true
			}
			if contains(p.Description, "Will Always Resolve") || contains(p.Description, "Will always resolve") {
				// handled below
			}
		}
		var dist string
		if usesTraitDist {
			dist = "Trait roll"
			if extraMeters > 0 {
				dist += fmt.Sprintf(" + %dm", extraMeters)
			}
		} else {
			baseM := 1 + extraMeters
			dist = fmt.Sprintf("%dm", baseM)
		}
		dirs := joinStrings(e.DirectionOptions, ", ")
		origin := e.Origin
		if originChanged {
			origin += " (custom - discuss with GM)"
		}
		lines = append(lines, fmt.Sprintf("Movement: %s", dist))
		lines = append(lines, fmt.Sprintf("Directions: %s", dirs))
		lines = append(lines, fmt.Sprintf("Origin: %s", origin))
		for _, p := range e.Perks {
			if contains(p.Description, "Will Always Resolve") || contains(p.Description, "Will always resolve") {
				lines = append(lines, "Always resolves")
			}
		}
	case models.EnactmentProficiencyShift:
		// Compute effective shift
		shiftAmt := e.ShiftAmount
		shiftUses := e.ShiftUses
		canChoose := false
		for _, p := range e.Perks {
			if contains(p.Description, "Shift the same Trait a second time") {
				shiftAmt += p.Amount
			}
			if contains(p.Description, "Add another use") {
				shiftUses += p.Amount
			}
			if contains(p.Description, "You may choose") {
				canChoose = true
			}
		}
		trait := e.ShiftedTrait
		if trait == "" {
			trait = "(select trait)"
		}
		lines = append(lines, fmt.Sprintf("Shift: %s %s by %d tier(s)", trait, e.ShiftDirection, shiftAmt))
		lines = append(lines, fmt.Sprintf("Uses: %d", shiftUses))
		if canChoose {
			lines = append(lines, "Optional: you may choose whether to use the shifted proficiency")
		}
		for _, p := range e.Perks {
			if contains(p.Description, "Will Always Resolve") || contains(p.Description, "Will always resolve") {
				lines = append(lines, "Always resolves")
			}
		}
	case models.EnactmentPersistentEffect:
		// Compute effective duration
		duration := 2 // default
		for _, p := range e.Perks {
			if contains(p.Description, "Add another round") {
				duration += p.Amount
			}
		}
		flavorLabel := "effect"
		if e.EffectFlavor != "" {
			flavorLabel = e.EffectFlavor
		}
		lines = append(lines, fmt.Sprintf("Effect: %s", flavorLabel))
		lines = append(lines, fmt.Sprintf("Duration: %d rounds", duration))
		lines = append(lines, fmt.Sprintf("Trigger: %s", e.TriggerTiming))
		if len(e.Solutions) > 0 {
			lines = append(lines, fmt.Sprintf("Solutions: %s", joinStrings(e.Solutions, ", ")))
		}
	}

	if e.IsOptional {
		lines = append(lines, "[optional enactment - can skip to save energy]")
	}

	return lines
}

func computeInteractionSummary(e *models.Enactment) []string {
	if len(e.Interactions) == 0 {
		return nil
	}
	inter := &e.Interactions[0]
	if string(inter.Type) == "Inherited" {
		return []string{"Interaction: Inherited from previous"}
	}

	var lines []string
	interType := string(inter.Type)

	// Count perks by summing amounts across all matching perks
	countPerk := func(perks []models.Perk, substr string) int {
		total := 0
		for _, p := range perks {
			if contains(p.Description, substr) {
				total += p.Amount
			}
		}
		return total
	}
	hasPerk := func(perks []models.Perk, substr string) bool {
		return countPerk(perks, substr) > 0
	}

	switch interType {
	case "Self":
		lines = append(lines, "Target: Self")
	case "Direct":
		extraRange := countPerk(inter.Perks, "Increase range")
		extraTargets := countPerk(inter.Perks, "Add another target")
		effectiveRange := 1 + extraRange
		targets := 1 + extraTargets
		lines = append(lines, fmt.Sprintf("Type: Direct | Range: %dm | Targets: %d", effectiveRange, targets))
	case "Ranged":
		extraRange := countPerk(inter.Perks, "Increase range") * 2
		decreaseRange := countPerk(inter.Perks, "Decrease range") * 2
		extraTargets := countPerk(inter.Perks, "Add another target")
		notVisible := hasPerk(inter.Perks, "not have to be visible") || hasPerk(inter.Perks, "Target does not have to be visible")
		mayObstruct := hasPerk(inter.Perks, "may be obstructed") || hasPerk(inter.Perks, "Target may be obstructed")
		removedPenalty := hasPerk(inter.Perks, "Remove the Engagement Roll Penalty")

		effectiveRange := 10 + extraRange - decreaseRange
		targets := 1 + extraTargets
		line := fmt.Sprintf("Type: Ranged | Range: %dm | Targets: %d", effectiveRange, targets)
		if notVisible {
			line += " | No LOS needed"
		}
		if mayObstruct {
			line += " | Can be obstructed"
		}
		if removedPenalty {
			line += " | No -2 penalty"
		} else {
			line += " | -2 engage penalty"
		}
		lines = append(lines, line)
	case "Area":
		extraRadius := countPerk(inter.Perks, "Increase radius")
		extraAreaRange := countPerk(inter.Perks, "Increase Range") * 2
		originChanged := hasPerk(inter.Perks, "Change Origin")

		origin := "Engager"
		if inter.Origin != "" && inter.Origin != "Engager" {
			origin = inter.Origin
		}
		if originChanged && origin == "Engager" {
			origin = "(custom - discuss with GM)"
		}
		lines = append(lines, fmt.Sprintf("Type: Area | Radius: %dm | Range: %dm | Origin: %s",
			1+extraRadius, 0+extraAreaRange, origin))
	case "Area of Effect":
		extraRadius := countPerk(inter.Perks, "Increase radius")
		extraAoERange := countPerk(inter.Perks, "Increase Range") * 2
		extraDuration := countPerk(inter.Perks, "Increase the amount of rounds")
		originChanged := hasPerk(inter.Perks, "Change Origin")
		immune := hasPerk(inter.Perks, "immune")

		origin := "Engager"
		if inter.Origin != "" && inter.Origin != "Engager" {
			origin = inter.Origin
		}
		if originChanged && origin == "Engager" {
			origin = "(custom - discuss with GM)"
		}
		line := fmt.Sprintf("Type: AoE | Radius: %dm | Range: %dm | Duration: %d rounds | Origin: %s",
			1+extraRadius, 0+extraAoERange, 2+extraDuration, origin)
		if immune {
			line += " | Engager immune"
		}
		lines = append(lines, line)
	}

	if hasPerk(inter.Perks, "Will always resolve") || hasPerk(inter.Perks, "Will Always Resolve") {
		lines = append(lines, "Always resolves")
	}

	return lines
}

func computeValidationSummary(e *models.Enactment) []string {
	if len(e.Interactions) == 0 {
		return nil
	}
	inter := &e.Interactions[0]
	if inter.Validation == nil {
		return nil
	}
	val := inter.Validation

	var lines []string

	// Compute effective engagement roll
	engageRoll := val.EngagementRoll
	isGenericEngage := false
	engageTierShifts := 0
	for _, p := range val.Perks {
		if contains(p.Description, "Engage Roll") && contains(p.Description, "Generic") {
			isGenericEngage = true
			engageRoll = "Generic 1d6"
		}
		if contains(p.Description, "Generic Engagement Roll UP") {
			engageTierShifts += p.Amount
		}
		if contains(p.Description, "Generic Engagement Roll DOWN") {
			engageTierShifts -= p.Amount
		}
		if contains(p.Description, "Replace the Engagement Roll") {
			engageRoll = val.EngagementRoll + " (replaced)"
		}
	}
	if isGenericEngage {
		diceTiers := []string{"1d4", "1d6", "1d8", "1d10", "1d12", "1d20"}
		baseIdx := 1 // 1d6 default
		newIdx := baseIdx + engageTierShifts
		if newIdx < 0 {
			newIdx = 0
		}
		if newIdx >= len(diceTiers) {
			newIdx = len(diceTiers) - 1
		}
		engageRoll = fmt.Sprintf("Generic %s", diceTiers[newIdx])
	}

	// Compute effective counter roll
	isGenericCounter := false
	counterTierShifts := 0
	counterRolls := make([]string, len(val.CounterRoll))
	copy(counterRolls, val.CounterRoll)

	for _, p := range val.Perks {
		if contains(p.Description, "Counter Roll") && contains(p.Description, "Generic") && !contains(p.Description, "Shift") {
			isGenericCounter = true
			counterRolls = []string{"Generic 1d12"}
		}
		if contains(p.Description, "Generic Counter Roll UP") {
			counterTierShifts += p.Amount
		}
		if contains(p.Description, "Generic Counter Roll DOWN") {
			counterTierShifts -= p.Amount
		}
		if contains(p.Description, "Remove one Counter Roll") {
			if len(counterRolls) > 1 {
				counterRolls = counterRolls[:1]
			}
		}
	}
	if isGenericCounter {
		diceTiers := []string{"1d4", "1d6", "1d8", "1d10", "1d12", "1d20"}
		baseIdx := 4 // 1d12 default
		newIdx := baseIdx + counterTierShifts
		if newIdx < 0 {
			newIdx = 0
		}
		if newIdx >= len(diceTiers) {
			newIdx = len(diceTiers) - 1
		}
		counterRolls = []string{fmt.Sprintf("Generic %s", diceTiers[newIdx])}
	}

	counterStr := joinStrings(counterRolls, " or ")
	lines = append(lines, fmt.Sprintf("Engage: %s vs Counter: %s", engageRoll, counterStr))

	// Note special conditions
	for _, p := range val.Perks {
		if contains(p.Description, "Will always resolve") || contains(p.Description, "Will Always Resolve") {
			lines = append(lines, "Always resolves (no roll needed)")
		}
	}

	return lines
}

func filterValidationPerks(allPerks []gamedata.PerkDef, hasGenericEngage, hasGenericCounter bool) []gamedata.PerkDef {
	var filtered []gamedata.PerkDef
	for _, p := range allPerks {
		// "Shift Generic Engagement Roll UP/DOWN" requires generic engage roll
		if contains(p.Description, "Shift") && contains(p.Description, "Generic") && contains(p.Description, "Engagement") {
			if !hasGenericEngage {
				continue
			}
		}
		// "Shift Generic Counter Roll UP/DOWN" requires generic counter roll
		if contains(p.Description, "Shift") && contains(p.Description, "Generic") && contains(p.Description, "Counter") {
			if !hasGenericCounter {
				continue
			}
		}
		filtered = append(filtered, p)
	}
	return filtered
}

func buildAbilityOverview(a *models.Ability) (overview string, quickSummary string) {
	// Compute effective type-level stats from perks
	extraRange := 0
	extraUses := 0
	for _, p := range a.Perks {
		if contains(p.Description, "Add 1 meter to reaction range") {
			extraRange += p.Amount
		}
		if contains(p.Description, "Add one more use per round") {
			extraUses += p.Amount
		}
	}

	// Quick summary line
	parts := []string{string(a.Type)}
	if a.ActionCost > 0 {
		parts = append(parts, fmt.Sprintf("%d Actions", a.ActionCost))
	}
	parts = append(parts, fmt.Sprintf("%d Energy", a.TotalEnergyCost()))

	// Reaction-specific
	if a.Type == models.AbilityTypeReaction {
		effectiveRange := a.Range + extraRange
		effectiveUses := a.Uses + extraUses
		parts = append(parts, fmt.Sprintf("Range: %dm", effectiveRange))
		parts = append(parts, fmt.Sprintf("Uses: %d/round", effectiveUses))
	}

	for _, e := range a.Enactments {
		desc := string(e.Type)
		computed := computeEffectiveDice(&e)
		if computed != "" {
			desc += " " + computed
		}
		if len(e.Interactions) > 0 {
			inter := e.Interactions[0]
			desc += " → " + string(inter.Type)
			if inter.Range != "" && inter.Range != "0m" {
				desc += " " + inter.Range
			}
		}
		parts = append(parts, desc)
	}
	quickSummary = joinStrings(parts, " | ")

	// Full overview in plain English
	var lines []string
	switch a.Type {
	case models.AbilityTypeReaction:
		effectiveRange := a.Range + extraRange
		effectiveUses := a.Uses + extraUses
		lines = append(lines, fmt.Sprintf("This is a Reaction costing %d energy. Range: %dm, %d use(s) per round.",
			a.TotalEnergyCost(), effectiveRange, effectiveUses))
		if len(a.Triggers) > 0 {
			lines = append(lines, fmt.Sprintf("Trigger: %s.", a.Triggers[0].Name))
		}
	case models.AbilityTypePhase:
		lines = append(lines, fmt.Sprintf("This is a Phase ability costing %d energy. Active for %d rounds, then Reverse Phase starts.",
			a.TotalEnergyCost(), a.PhaseDuration))
	default:
		lines = append(lines, fmt.Sprintf("This is a %s ability costing %d energy and %d action(s).",
			a.Type, a.TotalEnergyCost(), a.ActionCost))
	}

	for i, e := range a.Enactments {
		prefix := fmt.Sprintf("Step %d:", i+1)
		computed := computeEffectiveDice(&e)
		switch e.Type {
		case models.EnactmentDamage:
			if computed != "" {
				lines = append(lines, fmt.Sprintf("%s Deal %s damage.", prefix, computed))
			} else {
				lines = append(lines, fmt.Sprintf("%s Deal damage.", prefix))
			}
		case models.EnactmentHealing:
			if computed != "" {
				lines = append(lines, fmt.Sprintf("%s Heal for %s.", prefix, computed))
			} else {
				lines = append(lines, fmt.Sprintf("%s Heal target.", prefix))
			}
		case models.EnactmentMovement:
			dirs := joinStrings(e.DirectionOptions, "/")
			lines = append(lines, fmt.Sprintf("%s Move target %s %s.", prefix, e.MinimalDistance, dirs))
		case models.EnactmentProficiencyShift:
			lines = append(lines, fmt.Sprintf("%s Shift %s %s by %d tier(s) for %d use(s).",
				prefix, e.ShiftedTrait, e.ShiftDirection, e.ShiftAmount, e.ShiftUses))
		case models.EnactmentPersistentEffect:
			lines = append(lines, fmt.Sprintf("%s Apply persistent effect for %s.", prefix, e.Duration))
		}
		if len(e.Interactions) > 0 {
			inter := e.Interactions[0]
			interDesc := string(inter.Type)
			if inter.Range != "" && inter.Range != "0m" {
				interDesc += ", range " + inter.Range
			}
			if inter.Radius != "" {
				interDesc += ", radius " + inter.Radius
			}
			if inter.TargetAmount > 0 {
				interDesc += fmt.Sprintf(", %d target(s)", inter.TargetAmount)
			}
			lines = append(lines, fmt.Sprintf("  Delivered via: %s.", interDesc))
			if inter.Validation != nil && inter.Validation.EngagementRoll != "" {
				counters := joinStrings(inter.Validation.CounterRoll, " or ")
				lines = append(lines, fmt.Sprintf("  Roll: %s vs %s.", inter.Validation.EngagementRoll, counters))
			}
		} else if i > 0 {
			lines = append(lines, "  Interaction: inherited from previous enactment.")
		}
	}
	overview = joinStrings(lines, "\n")
	return
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func stringContains(s, sub string) bool {
	return contains(s, sub)
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}

func (s *Server) handleAbilityWizardSetInteraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.FormValue("ability_index"))
	enactmentIndex, _ := strconv.Atoi(r.FormValue("enactment_index"))
	interactionType := r.FormValue("interaction_type")

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	if enactmentIndex < 0 || enactmentIndex >= len(ability.Enactments) {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	enactment := &ability.Enactments[enactmentIndex]

	// Create or replace the interaction
	newInteraction := models.Interaction{
		Type:    models.InteractionType(interactionType),
		Engager: "Self",
	}

	// Set defaults based on interaction type
	switch interactionType {
	case "Self":
		newInteraction.Validation = &models.Validation{
			EngagementRoll: "Power",
			CounterRoll:    []string{"d8"},
		}
	case "Direct":
		newInteraction.TargetAmount = 1
		newInteraction.Range = "1m"
		newInteraction.Validation = &models.Validation{
			EngagementRoll: "Power",
			CounterRoll:    []string{"Reflex", "Constitution"},
		}
	case "Ranged":
		newInteraction.TargetAmount = 1
		newInteraction.Range = "10m"
		newInteraction.Visibility = "Visible"
		newInteraction.Obstruction = "Not obstructed"
		newInteraction.Validation = &models.Validation{
			EngagementRoll: "Precision - 2",
			CounterRoll:    []string{"Reflex", "Constitution"},
		}
	case "Area":
		newInteraction.Radius = "1m"
		newInteraction.Range = "0m"
		newInteraction.Origin = "Engager"
		newInteraction.Validation = &models.Validation{
			EngagementRoll: "Power",
			CounterRoll:    []string{"Reflex", "Constitution"},
		}
	case "Area of Effect":
		newInteraction.Radius = "1m"
		newInteraction.Range = "0m"
		newInteraction.Origin = "Engager"
		newInteraction.Duration = "2 rounds"
		newInteraction.Validation = &models.Validation{
			EngagementRoll: "Power",
			CounterRoll:    []string{"Reflex", "Constitution"},
		}
	}

	// Replace or add interaction
	if len(enactment.Interactions) == 0 {
		enactment.Interactions = []models.Interaction{newInteraction}
	} else {
		enactment.Interactions[0] = newInteraction
	}

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardUpdateValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.FormValue("ability_index"))
	enactmentIndex, _ := strconv.Atoi(r.FormValue("enactment_index"))
	engagementRoll := r.FormValue("engagement_roll")
	counterRoll1 := r.FormValue("counter_roll_1")
	counterRoll2 := r.FormValue("counter_roll_2")

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	if enactmentIndex < 0 || enactmentIndex >= len(ability.Enactments) {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	enactment := &ability.Enactments[enactmentIndex]
	if len(enactment.Interactions) == 0 {
		// Create a minimal interaction to hold the validation (interaction inherited from previous)
		enactment.Interactions = []models.Interaction{{
			Type:    "Inherited",
			Engager: "Self",
		}}
	}

	interaction := &enactment.Interactions[0]
	if interaction.Validation == nil {
		interaction.Validation = &models.Validation{}
	}

	interaction.Validation.EngagementRoll = engagementRoll
	counterRolls := []string{}
	if counterRoll1 != "" {
		counterRolls = append(counterRolls, counterRoll1)
	}
	if counterRoll2 != "" {
		// Prevent duplicate counter rolls unless "Remove one Counter Roll option" perk is active
		if counterRoll2 == counterRoll1 {
			// Check if they have the remove perk (which means they only need 1 counter)
			hasRemovePerk := false
			if interaction.Validation != nil {
				for _, p := range interaction.Validation.Perks {
					if contains(p.Description, "Remove one Counter Roll") {
						hasRemovePerk = true
						break
					}
				}
			}
			if !hasRemovePerk {
				// Silently skip the duplicate - they need different rolls
				// Don't add the second one
			} else {
				counterRolls = append(counterRolls, counterRoll2)
			}
		} else {
			counterRolls = append(counterRolls, counterRoll2)
		}
	}
	interaction.Validation.CounterRoll = counterRolls

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardAddInteractionPerk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.FormValue("ability_index"))
	enactmentIndex, _ := strconv.Atoi(r.FormValue("enactment_index"))
	perkDesc := r.FormValue("perk_description")
	perkAddCost, _ := strconv.Atoi(r.FormValue("perk_add_cost"))
	perkEnergyCost, _ := strconv.Atoi(r.FormValue("perk_energy_cost"))
	target := r.FormValue("target") // "interaction" or "validation"
	perkExtra := r.FormValue("perk_extra")

	// Append extra input to description if provided
	if perkExtra != "" {
		perkDesc = perkDesc + ": " + perkExtra
	}

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	if enactmentIndex < 0 || enactmentIndex >= len(ability.Enactments) {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	enactment := &ability.Enactments[enactmentIndex]
	if len(enactment.Interactions) == 0 {
		// Auto-create a minimal interaction to hold validation perks (inherited interaction case)
		enactment.Interactions = []models.Interaction{{
			Type:    "Inherited",
			Engager: "Self",
			Validation: &models.Validation{
				EngagementRoll: "Power",
				CounterRoll:    []string{"Reflex", "Constitution"},
			},
		}}
	}

	// Budget check
	if perkAddCost > 0 && perkAddCost > char.AbilityPointsAvailable() {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "insufficient ability points (need %d, have %d)"}`, perkAddCost, char.AbilityPointsAvailable()))
		vm := s.buildWizardVM(char, abilityIndex)
		s.renderTemplate(w, "ability_wizard", vm)
		return
	}

	newPerk := models.Perk{
		Description:  perkDesc,
		AddCost:      perkAddCost,
		Amount:       1,
		TotalAddCost: perkAddCost,
		EnergyCost:   perkEnergyCost,
		IsOptional:   false,
	}

	interaction := &enactment.Interactions[0]
	if target == "validation" {
		if interaction.Validation == nil {
			interaction.Validation = &models.Validation{}
		}
		interaction.Validation.Perks = append(interaction.Validation.Perks, newPerk)
	} else {
		interaction.Perks = append(interaction.Perks, newPerk)
	}

	vm := s.buildWizardVM(char, abilityIndex)
	// Keep the relevant section open
	if target == "validation" {
		vm.OpenSection = fmt.Sprintf("validation-%d", enactmentIndex)
	} else {
		vm.OpenSection = fmt.Sprintf("interaction-%d", enactmentIndex)
	}
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardRemoveInteractionPerk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.URL.Query().Get("ability_index"))
	enactmentIndex, _ := strconv.Atoi(r.URL.Query().Get("enactment_index"))
	perkIndex, _ := strconv.Atoi(r.URL.Query().Get("perk_index"))
	target := r.URL.Query().Get("target") // "interaction" or "validation"

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	if enactmentIndex < 0 || enactmentIndex >= len(ability.Enactments) {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	enactment := &ability.Enactments[enactmentIndex]
	if len(enactment.Interactions) == 0 {
		http.Error(w, "No interaction set", http.StatusBadRequest)
		return
	}

	interaction := &enactment.Interactions[0]
	if target == "validation" && interaction.Validation != nil {
		if perkIndex >= 0 && perkIndex < len(interaction.Validation.Perks) {
			interaction.Validation.Perks = append(interaction.Validation.Perks[:perkIndex], interaction.Validation.Perks[perkIndex+1:]...)
		}
	} else {
		if perkIndex >= 0 && perkIndex < len(interaction.Perks) {
			interaction.Perks = append(interaction.Perks[:perkIndex], interaction.Perks[perkIndex+1:]...)
		}
	}

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardToggleOptional(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.FormValue("ability_index"))
	perkTarget := r.FormValue("perk_target")
	perkIndex, _ := strconv.Atoi(r.FormValue("perk_index"))

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]

	// Find and toggle the perk
	if perkTarget == "ability" {
		if perkIndex >= 0 && perkIndex < len(ability.Perks) {
			ability.Perks[perkIndex].IsOptional = !ability.Perks[perkIndex].IsOptional
		}
	} else if len(perkTarget) > 10 && perkTarget[:9] == "enactment" {
		enactIdx, _ := strconv.Atoi(perkTarget[10:])
		if enactIdx >= 0 && enactIdx < len(ability.Enactments) {
			e := &ability.Enactments[enactIdx]
			if perkIndex >= 0 && perkIndex < len(e.Perks) {
				e.Perks[perkIndex].IsOptional = !e.Perks[perkIndex].IsOptional
			}
		}
	}

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}

func (s *Server) handleAbilityWizardUpdateEffectFlavor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	abilityIndex, _ := strconv.Atoi(r.FormValue("ability_index"))
	enactmentIndex, _ := strconv.Atoi(r.FormValue("enactment_index"))
	flavor := r.FormValue("effect_flavor")

	if abilityIndex < 0 || abilityIndex >= len(char.Abilities) {
		http.Error(w, "Invalid ability index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[abilityIndex]
	if enactmentIndex < 0 || enactmentIndex >= len(ability.Enactments) {
		http.Error(w, "Invalid enactment index", http.StatusBadRequest)
		return
	}

	ability.Enactments[enactmentIndex].EffectFlavor = flavor

	vm := s.buildWizardVM(char, abilityIndex)
	s.renderTemplate(w, "ability_wizard", vm)
}
