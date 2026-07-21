// Ability Builder — JS-rendered cards
//
// All cards (ability-type, enactment, interaction, validation) are rendered
// entirely in JS from BUILDER_DATA. The server renders only shell + form
// fields that are not card-scoped (ability name, description, etc.).
//
// Multi-enactments are tracked by an internal index counter; "+ Add
// Enactment" appends a new block. The form-submit handler prefixes fields
// with their block index so multiple-enactment POSTs can be parsed on the
// server.
//
// Each section card uses a `data-section="..."` attribute. The JS uses
// `closest(".section-card")` to scope queries, so multiple enactments do
// not collide.

(function () {
  'use strict';

  // Escape HTML for use in template literals
  function esc(s) {
    if (s === null || s === undefined) return '';
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
  function ucfirst(s) { return s ? s.charAt(0).toUpperCase() + s.slice(1) : ''; }
  function selected(s, v) { return s === v ? 'selected' : ''; }
  function checked(v) { return v ? 'checked' : ''; }
  function hiddenIf(v) { return v ? '' : 'hidden'; }

  // --- Cost display helpers -------------------------------------------------
  // Build a " (+/-xpt, +/-yE)" suffix. A positive add/energy value increases
  // the total cost (shown with '+'); a negative value decreases the total
  // cost (shown with '-').
  // Example: add=5, energy=3 -> " (+5pt, +3E)"; add=-2 -> " (-2pt, +0E)".
  function costSuffix(add, energy) {
    add = Number(add) || 0;
    energy = Number(energy) || 0;
    if (add === 0 && energy === 0) return '';
    function sgn(v) { return v > 0 ? '+' : '-'; }
    var parts = [];
    if (add !== 0) parts.push(sgn(add) + Math.abs(add) + 'pt');
    if (energy !== 0) parts.push(sgn(energy) + Math.abs(energy) + 'E');
    return ' (' + parts.join(', ') + ')';
  }
  // Build a single <option> with a cost suffix.
  function opt(label, value, isSel, add, energy) {
    return '<option value="' + esc(value) + '" ' + selected(isSel, value) + '>' + esc(label) + costSuffix(add, energy) + '</option>';
  }

  var D = window.BUILDER_DATA;
  var ABILITY_TYPES = D.abilityTypes;
  var ENACT_TYPES = D.allEnactmentTypes;
  var INTER_TYPES = D.interactionTypes;

  // Global registry of generic field schemas keyed by card uid. Each generic-
  // rendered card gets a data-generic-id attribute; the schema lives in this
  // map so it survives HTML insertion and round-trips through the DOM.
  var __genericFieldRegistry = {};
  var __genericIdCounter = 0;
  function registerGenericFields(fields) {
    var id = 'gf_' + (++__genericIdCounter);
    __genericFieldRegistry[id] = fields;
    return id;
  }
  function getGenericFieldsForCard(card) {
    if (!card) return null;
    var id = card.getAttribute('data-generic-id');
    if (!id) return null;
    return __genericFieldRegistry[id] || null;
  }

  // =========================================================================
  // ID generators. We use unique indices within each block instead of global
  // ids so multiple enactments do not collide. Helper to make input names
  // unique across all blocks: each form input gets a prefix from its block
  // (e.g. "enact_0_"). The submit handler renames fields using these indices.
  // =========================================================================
  var enactmentCounter = 0;
  function nextIndex() { return enactmentCounter++; }

  function bindRange(select, value) {
    if (value === undefined) return '';
    var n = Number(value);
    var opts = '<option value="'+n+'">'+n+'m</option>';
    return opts;
  }

  // =========================================================================
  // Section: ability-type cards
  // =========================================================================

  function renderAbilityTypeCard(type, data) {
    data = data || {};
    var cfg = (C.ability_types && C.ability_types[type.toLowerCase()]) || {};
    var inner = '';

    if (cfg.fields && cfg.fields.length) {
      var fieldsHolder = '<div class="space-y-2">';
      var hydration = data.fields || data;
      for (var i = 0; i < cfg.fields.length; i++) {
        var visClass = isFieldVisible(cfg.fields[i], hydration, cfg.fields) ? '' : 'hidden';
        fieldsHolder += '<div data-field-key="' + esc(cfg.fields[i].key) + '"' + (visClass ? ' ' + visClass : '') + '>' + renderFieldHTML(cfg.fields[i], hydration) + '</div>';
      }
      fieldsHolder += '</div>';
      inner = '<h3 class="text-md font-semibold text-indigo-400">Ability Type — ' + esc(type) + '</h3>' +
              overview(false) + fieldsHolder + breakdown();
      var gid = registerGenericFields(cfg.fields);
      return '<div class="section-card ability-type-card bg-gray-800 rounded-lg border border-gray-700 p-5 space-y-4" data-section="ability-type" data-ability-type="'+esc(type)+'" data-generic-id="'+gid+'" data-build="0" data-cast="0">' + inner + '</div>';
    } else if (type === 'Execution') {
      inner = renderExecutionCard(data, cfg);
    } else if (type === 'Reaction') {
      inner = renderReactionCard(data, cfg);
    } else if (type === 'Phase') {
      inner = renderPhaseCard(data, cfg);
    } else if (type === 'Minion') {
      inner = renderMinionCard(data, cfg);
    } else if (type === 'Preparation') {
      inner = renderPreparationCard(data, cfg);
    } else if (type === 'Concentration') {
      inner = renderConcentrationCard(data, cfg);
    } else {
      inner = '<p class="text-yellow-400">Unknown ability type: '+esc(type)+'</p>';
    }

    return '<div class="section-card ability-type-card bg-gray-800 rounded-lg border border-gray-700 p-5 space-y-4" data-section="ability-type" data-ability-type="'+esc(type)+'">'+inner+'</div>';
  }

  function stepCostFn(cfg, key) {
    return function (v) {
      var dir = v > 0 ? 'increase' : (v < 0 ? 'decrease' : null);
      if (!dir) return { add: 0, energy: 0 };
      var c = stepCost(cfg.step_costs && cfg.step_costs[key], dir);
      var n = Math.abs(v);
      return { add: n * c.add, energy: n * c.energy };
    };
  }
  function cumCostFn(base, perUnit) {
    perUnit = perUnit || { add_cost: 0, energy_cost: 0 };
    return function (i) {
      var steps = Math.max(0, i - base);
      return { add: steps * (perUnit.add_cost || 0), energy: steps * (perUnit.energy_cost || 0) };
    };
  }

  function renderExecutionCard(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var itemName = d.item_name || '';
    var itemWrap = d.item_dep ? '' : 'hidden';
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Execution</h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        itemDepCheckbox(d.item_dep, perkCost(cfg.perks, 'item_dependency')),
        stepSelect('energy_steps', 'Energy ±', [-2,-1,0,1,2], d.energy_steps || 0, stepCostFn(cfg, 'energy')),
        stepSelect('action_steps', 'Action ±', [-1,0,1], d.action_steps || 0, stepCostFn(cfg, 'action')),
      '</div>',
      '<div data-wrap="item-name" '+itemWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(itemName)+'" placeholder="e.g. Silver Dagger" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderReactionCard(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var triggerNeedsTrait = d.trigger === 'Target makes a trait check';
    var triggerWrap = triggerNeedsTrait ? '' : 'hidden';
    var triggerOpts = D.reactionTriggers.map(function(t){
      var c = findPerk(cfg.triggers, t) || { add_cost: 0, energy_cost: 0 };
      return opt(t, t, d.trigger, c.add_cost, c.energy_cost);
    }).join('');
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Reaction</h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
        '<div><label class="block text-xs text-gray-400 mb-1">Trigger</label>',
        '<select name="trigger" onchange="onReactionTriggerChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
          '<option value="">-- Select --</option>',
          triggerOpts,
        '</select></div>',
        '<div data-wrap="trigger-trait" '+triggerWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Trigger Trait</label>',
          '<select name="trigger_trait" onchange="recalcAll()" class="bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white w-full">',
            traitOptions('defense', d.trigger_trait),
          '</select></div>',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        intSelect('range', 'Range', 1, 6, d.reaction_range || 1, 'm', cumCostFn(1, cfg.range_cost)),
        intSelect('uses', 'Uses', 1, 3, d.reaction_uses || 1, '', cumCostFn(1, cfg.uses_cost)),
        itemDepCheckbox(d.item_dep, perkCost(cfg.perks, 'item_dependency')),
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" placeholder="e.g. Shield" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderPhaseCard(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var kos = d.knockouts || [];
    var showKos = !d.no_knockout;
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Phase</h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
        intSelect('phase_rounds', 'Phase Duration', 2, 5, d.phase_duration || 2, ' rounds', cumCostFn(2, cfg.duration_cost)),
        intSelect('reverse_rounds', 'Reverse Rounds', 1, 5, d.reverse_phase_rounds || 2, ' rounds', function(i){
          var rev = cfg.reverse_duration_refund || { add_cost: 0, energy_cost: 0 };
          var steps = Math.max(0, 2 - i);
          return { add: steps * (rev.add_cost || 0), energy: steps * (rev.energy_cost || 0) };
        }),
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">',
        checkbox('all_req', 'All knockout requirements have to be met', d.all_knockouts_req, perkCost(cfg.perks, 'all_knockouts_req')),
        checkbox('reverse_knockout', 'Knockout can be used on reverse phase', d.reverse_knockout_ok, perkCost(cfg.perks, 'reverse_knockout')),
      '</div>',
      '<div>',
        '<label class="flex items-center gap-2 text-sm text-gray-300 mb-2">',
          '<input type="checkbox" name="no_knockout" onchange="onNoKnockoutChange(this)" '+checked(d.no_knockout)+' class="rounded bg-gray-700 border-gray-600">',
          '<span><strong>No knockout possible</strong> — the phase cannot be ended by any condition'+costSuffix((findPerk(cfg.perks, 'no_knockout')||{}).add_cost||0, (findPerk(cfg.perks, 'no_knockout')||{}).energy_cost||0)+'</span>',
        '</label>',
        '<div data-wrap="knockouts" '+hiddenIf(showKos)+'>',
          '<div class="text-xs text-gray-400 uppercase mb-1">Knockouts</div>',
          knockoutList(kos, cfg),
        '</div>',
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderMinionCard(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    return [
      '<h3 class="text-md font-semibold text-indigo-400 flex items-center gap-2">Ability Type — Minion <span class="text-xs text-yellow-400">(WIP)</span></h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        intSelectFlat('hp', 'Health Bonus', 0, 5, d.hp_bonus || 0, cumCostFn(0, cfg.health_bonus_cost)),
        intSelectFlat('life', 'Extra Lifetime', 0, 5, d.extra_lifetime || 0, cumCostFn(0, cfg.lifetime_bonus_cost)),
        itemDepCheckbox(d.item_dep, perkCost(cfg.perks, 'item_dependency')),
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderPreparationCard(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var triggerNeedsTrait = d.trigger === 'Target makes a trait check';
    var triggerWrap = triggerNeedsTrait ? '' : 'hidden';
    var triggerOpts = (cfg.triggers || []).map(function(t){
      var c = { add_cost: t.add_cost || 0, energy_cost: t.energy_cost || 0 };
      return opt(t.id, t.id, d.trigger, c.add_cost, c.energy_cost);
    }).join('');
    if (D.reactionTriggers) {
      triggerOpts = D.reactionTriggers.map(function(t){
        var c = findPerk(cfg.triggers, t) || { add_cost: 0, energy_cost: 0 };
        return opt(t, t, d.trigger, c.add_cost, c.energy_cost);
      }).join('');
    }
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Preparation</h3>',
      overview(false),
      '<p class="text-xs text-gray-400">A Preparation works like a Reaction but must be prepared with an action. The first trigger is free; additional triggers cost per the table below.</p>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
        '<div><label class="block text-xs text-gray-400 mb-1">Trigger</label>',
        '<select name="trigger" onchange="onReactionTriggerChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
          '<option value="">-- Select --</option>',
          triggerOpts,
        '</select></div>',
        '<div data-wrap="trigger-trait" '+triggerWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Trigger Trait</label>',
          '<select name="trigger_trait" onchange="recalcAll()" class="bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white w-full">',
            traitOptions('defense', d.trigger_trait),
          '</select></div>',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
        intSelect('range', 'Range', 1, 6, d.reaction_range || 1, 'm', cumCostFn(1, cfg.range_cost)),
        intSelect('uses', 'Uses', 1, 3, d.reaction_uses || 1, '', cumCostFn(1, cfg.uses_cost)),
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        stepSelect('action_steps', 'Action ±', [-1,0,1], d.action_steps || 0, stepCostFn(cfg, 'action')),
        stepSelect('energy_steps', 'Energy ±', [-2,-1,0,1,2], d.energy_steps || 0, stepCostFn(cfg, 'energy')),
        itemDepCheckbox(d.item_dep, perkCost(cfg.perks, 'item_dependency')),
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" placeholder="e.g. Shield" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderConcentrationCard(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var itemWrap = d.item_dep ? '' : 'hidden';
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Concentration</h3>',
      overview(false),
      '<p class="text-xs text-gray-400">A Concentration Ability persists for multiple rounds. At the start of each of your turns you must pay the upkeep cost ('+esc(cfg.base_upkeep_action||1)+' Action or '+esc(cfg.base_upkeep_energy||1)+' Energy) or the ability ends.</p>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
        stepSelect('energy_steps', 'Energy ±', [-2,-1,0,1,2], d.energy_steps || 0, stepCostFn(cfg, 'energy')),
        itemDepCheckbox(d.item_dep, perkCost(cfg.perks, 'item_dependency')),
      '</div>',
      '<div data-wrap="item-name" '+itemWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" placeholder="e.g. Crystal Focus" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      '<div class="space-y-1">',
        '<div class="text-xs text-gray-400 uppercase">Concentration Perks</div>',
        perkCheckbox('effortless', 'Effortless (upkeep is free)', d.effortless, perkCost(cfg.perks, 'effortless')),
        perkCheckbox('iron_will', 'Iron Will (shift counter roll up on damage)', d.iron_will, perkCost(cfg.perks, 'iron_will')),
        perkCheckbox('dual_focus', 'Dual Focus (allow a second Concentration)', d.dual_focus, perkCost(cfg.perks, 'dual_focus')),
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function perkCheckbox(name, label, value, cost) {
    var suffix = cost ? costSuffix(cost.add, cost.energy) : '';
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="'+esc(name)+'" onchange="recalcAll()" '+checked(value)+' class="rounded bg-gray-700 border-gray-600">'+
      esc(label)+suffix+
    '</label>';
  }

  function knockoutList(values, cfg) {
    values = values || [];
    var rows = '';
    rows += knockoutRow(values[0], cfg);
    if (values.length > 1) {
      for (var i = 1; i < values.length; i++) rows += knockoutRow(values[i], cfg);
    } else {
      rows += knockoutRow(null, cfg);
    }
    return '<div data-list="knockouts" class="space-y-2">'+rows+'</div>';
  }
  function knockoutRow(value, cfg) {
    cfg = cfg || {};
    var kos = cfg.knockout_requirements || [];
    var opts = '<option value="">-- Select --</option>' +
      D.knockoutOptions.map(function(k){
        var c = findPerk(kos, k) || { add_cost: 0, energy_cost: 0 };
        return opt(k, k, value, c.add_cost, c.energy_cost);
      }).join('');
    return '<div class="flex items-center gap-2">'+
      '<select name="knockout" onchange="onKnockoutChange(this)" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
      '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>'+
    '</div>';
  }

  // Config helpers
  var C = D.cfg ? (typeof D.cfg === 'string' ? JSON.parse(D.cfg) : D.cfg) : {};
  function abilityTypeConfig() {
    var card = document.querySelector('.section-card[data-section="ability-type"]');
    if (!card) return {};
    var fields = getGenericFieldsForCard(card);
    if (fields) return C.ability_types && C.ability_types[card.dataset.abilityType.toLowerCase()] || {};
    var type = card.dataset.abilityType;
    return C.ability_types && C.ability_types[type.toLowerCase()] || {};
  }
  function findPerk(perks, id) {
    if (!perks) return null;
    for (var i = 0; i < perks.length; i++) if (perks[i].id === id) return perks[i];
    return null;
  }
  function perkCost(perks, id) {
    var p = findPerk(perks, id);
    return p ? { add: p.add_cost || 0, energy: p.energy_cost || 0 } : { add: 0, energy: 0 };
  }
  function stepCost(steps, direction) {
    if (!steps) return { add: 0, energy: 0 };
    var s = steps[direction];
    return s ? { add: s.add_cost || 0, energy: s.energy_cost || 0 } : { add: 0, energy: 0 };
  }
  function getEnactConfig(type) {
    if (!type) return {};
    var key = type.toLowerCase().replace(/enact /g, '').replace(/ /g, '_');
    return C.enactments && C.enactments[key] || {};
  }
  function getInterConfig(type) {
    if (!type) return {};
    var key = type.toLowerCase().replace(/ /g, '_');
    return C.interactions && C.interactions[key] || {};
  }
  function getValidationConfig() { return C.validations || {}; }

  // =========================================================================
  // Generic field renderer (schema-driven). Mirrors server's internal/config/
  // field_costs.go. Returns HTML strings so the host card can include them.
  // =========================================================================
  function resolveOptions(source) {
    if (!source) return [];
    switch (source) {
      case 'traits_general': return D.generalTraits || [];
      case 'traits_offense': return D.offenseTraits || [];
      case 'traits_defense': return D.defenseTraits || [];
      case 'traits_all':
        return (D.generalTraits || []).concat(D.offenseTraits || [], D.defenseTraits || []);
      case 'dice_damage': return D.damageDiceOptions || [];
      case 'dice_generic': return D.genericDieOptions || [];
      case 'states_general': return (C.states && C.states.general_states) || [];
      case 'states_specific': return (C.states && C.states.specific_states) || [];
      case 'directions_all':
      case 'directions':
        return D.directionOptions || [];
      case 'shift_directions': return D.shiftDirectionOptions || [];
      case 'trigger_timings': return D.triggerTimings || [];
      case 'aoe_trigger_timings': return D.aoeTriggerTimings || [];
      case 'knockout_options': return D.knockoutOptions || [];
      case 'reaction_triggers': return D.reactionTriggers || [];
      case 'ability_types': return D.abilityTypes || [];
      case 'enactment_types': return D.allEnactmentTypes || [];
      case 'interaction_types': return D.interactionTypes || [];
    }
    return [];
  }
  function isFieldVisible(field, allValues, fields) {
    if (!field.visibility_when) return true;
    var v = allValues ? allValues[field.visibility_when] : undefined;
    if (v === undefined || v === null || v === '') {
      if (fields) {
        for (var i = 0; i < fields.length; i++) {
          if (fields[i].key === field.visibility_when) {
            if (fields[i].default !== undefined) {
              v = fields[i].default;
            }
            break;
          }
        }
      }
    }
    if (v === undefined || v === null) v = '';
    if (Array.isArray(v)) v = v.length ? v[0] : '';
    return String(v) === String(field.show_when);
  }
  function renderFieldHTML(field, data) {
    data = data || {};
    var name = field.key;
    var value = data[name];
    var label = '<label class="block text-xs text-gray-400 mb-1">' + esc(field.label) + '</label>';
    var costSuffixStr = field.cost ? costSuffix(field.cost.add_cost, field.cost.energy_cost) : '';
    switch (field.type) {
      case 'checkbox': {
        var isChecked = value !== undefined ? toBool(value) : toBool(field.default);
        return '<label class="flex items-center gap-2 text-sm text-gray-300">' +
          '<input type="checkbox" name="' + esc(name) + '" data-generic-field="' + esc(name) + '" onchange="onGenericFieldChange(this)" ' + checked(isChecked) +
          ' class="rounded bg-gray-700 border-gray-600">' +
          esc(field.label) + costSuffixStr + '</label>';
      }
      case 'dropdown': {
        var sel = toStr(value);
        if (sel === '' && field.default !== undefined) sel = String(field.default);
        var opts = field.options && field.options.length ? field.options : null;
        if (opts) {
          var html = label + '<select name="' + esc(name) + '" data-generic-field="' + esc(name) + '" onchange="onGenericFieldChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">';
          html += '<option value="">-- Select --</option>';
          for (var i = 0; i < opts.length; i++) {
            var o = opts[i];
            var cs = o.cost ? costSuffix(o.cost.add_cost, o.cost.energy_cost) : '';
            html += '<option value="' + esc(o.value) + '" ' + selected(sel, o.value) + '>' + esc(o.label) + cs + '</option>';
          }
          html += '</select>';
          return html;
        }
        var list = resolveOptions(field.options_source);
        var dynHtml = label + '<select name="' + esc(name) + '" data-generic-field="' + esc(name) + '" onchange="onGenericFieldChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">';
        dynHtml += '<option value="">-- Select --</option>';
        if (field.options_source === 'traits_all') {
          dynHtml += traitOptionsGrouped(toStr(value));
        } else if (field.options_source === 'traits_general') {
          dynHtml += traitOptionsForCategory('general', toStr(value));
        } else if (field.options_source === 'traits_offense') {
          dynHtml += traitOptionsForCategory('offense', toStr(value));
        } else if (field.options_source === 'traits_defense') {
          dynHtml += traitOptionsForCategory('defense', toStr(value));
        } else {
          for (var k = 0; k < list.length; k++) {
            var item = list[k];
            var itemVal = typeof item === 'string' ? item : item.id;
            var itemLabel = typeof item === 'string' ? item : item.name;
            var rowSel = toStr(value) || (field.default !== undefined ? String(field.default) : '');
            dynHtml += '<option value="' + esc(itemVal) + '" ' + selected(rowSel, itemVal) + '>' + esc(itemLabel) + '</option>';
          }
        }
        dynHtml += '</select>';
        return dynHtml;
      }
      case 'free_text': {
        var v = toStr(value);
        return label + '<input type="text" name="' + esc(name) + '" data-generic-field="' + esc(name) + '" value="' + esc(v) + '" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">';
      }
      case 'free_number': {
        var nv = toStr(value);
        if (!nv) nv = String(toInt(field.default));
        var mn = field.min || 0;
        var mx = field.max || 100;
        var opts2 = '';
        for (var s = mn; s <= mx; s++) {
          var c2 = calcFreeNumCost(field, s);
          opts2 += '<option value="' + s + '" ' + selected(Number(nv) === s, s) + '>' + s + costSuffix(c2.add, c2.energy) + '</option>';
        }
        return label + '<select name="' + esc(name) + '" data-generic-field="' + esc(name) + '" onchange="onGenericFieldChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">' + opts2 + '</select>';
      }
      case 'solutions': {
        var arr = Array.isArray(value) ? value : [];
        if (arr.length === 0 && field.default) {
          arr = Array.isArray(field.default) ? field.default : [field.default];
        }
        var rows = [];
        for (var i = 0; i < arr.length; i++) {
          var item = arr[i];
          if (typeof item === 'string') {
            rows.push({ value: item });
          } else if (item && typeof item === 'object') {
            rows.push({ value: item.value || item.solution || '', type: item.type || 'defense' });
          }
        }
        var defaultCount = field.default_count || 1;
        while (rows.length < defaultCount) rows.push({});
        var rowHTML = '';
        for (var ri = 0; ri < rows.length; ri++) {
          rowHTML += renderSolutionRow(name, ri, rows[ri], field);
        }
        var headerHtml = '<div class="flex items-center justify-between mb-1"><span class="text-xs text-gray-400 uppercase">' + esc(field.label) + '</span>' +
          '<button type="button" onclick="addGenericRow(this, \'' + esc(name) + '\')" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Add</button>' +
          '</div>';
        return headerHtml + '<div data-generic-list="' + esc(name) + '" data-generic-type="solutions" class="space-y-1">' + rowHTML + '</div>';
      }
      case 'states': {
        var statesList = Array.isArray(value) ? value : [];
        if (statesList.length === 0) {
          statesList = [{}];
        }
        var rendered = '';
        for (var si = 0; si < statesList.length; si++) {
          rendered += renderStateRow(name, si, statesList[si] || {}, field);
        }
        var header = '<div class="flex items-center justify-between mb-1"><span class="text-xs text-gray-400 uppercase">' + esc(field.label) + '</span>' +
          '<button type="button" onclick="addStateRow(this, \'' + esc(name) + '\')" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Add State</button>' +
          '</div>';
        return header + '<div data-states-list="' + esc(name) + '" data-generic-type="states" class="space-y-2">' + rendered + '</div>';
      }
    }
    return '';
  }
  function calcFreeNumCost(field, val) {
    var def = toInt(field.default);
    var step = field.step || 1;
    var delta = val - def;
    if (delta === 0 || !field.per_step) return { add: 0, energy: 0 };
    var dir = delta > 0 ? 'increase' : 'decrease';
    var sc = stepCostOf(field.per_step, dir);
    var steps;
    if (field.rounding === 'ceil') {
      steps = Math.ceil(delta / step);
      if (steps < 0) steps = 0;
    } else if (field.rounding === 'floor') {
      steps = Math.floor(delta / step);
    } else {
      steps = delta / step;
    }
    if (steps === 0) return { add: 0, energy: 0 };
    return { add: Math.abs(steps) * sc.add, energy: Math.abs(steps) * sc.energy };
  }
  function renderSolutionRow(name, idx, row, field) {
    row = row || {};
    var selectedVal = row.value || row.solution || '';
    var rowType = row.type || '';
    var list = resolveOptions(field.options_source);
    var rowFields = field.row_fields || [];
    var inner = '';
    for (var ri = 0; ri < rowFields.length; ri++) {
      var rf = rowFields[ri];
      var rv = row[rf.key];
      var fullKey = name + '__' + rf.key;
      var fieldCopy = {};
      for (var k in rf) { if (rf.hasOwnProperty(k)) fieldCopy[k] = rf[k]; }
      fieldCopy.key = fullKey;
      fieldCopy.label = '';
      if (rf.key === 'value' && rowType && rowType !== 'previous') {
        fieldCopy.options_source = 'traits_' + rowType;
      }
      inner += '<div data-row-field="' + esc(rf.key) + '">' + renderFieldHTML(fieldCopy, { [fullKey]: rv }) + '</div>';
    }
    return '<div class="bg-gray-900 rounded p-2 space-y-1" data-row>' + inner +
      '<div class="flex items-center gap-2">' +
      '<button type="button" onclick="this.closest(\'[data-row]\').remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs ml-auto">−</button>' +
      '</div></div>';
  }
  function renderStateRow(name, idx, row, field) {
    row = row || {};
    var kind = row.state_kind || '';
    var specs = (C.states && C.states.specific_states) || [];
    var gens = (C.states && C.states.general_states) || [];
    var specificOpts = '<option value="">-- Select --</option>';
    var generalOpts = '<option value="">-- Select --</option>';
    var kindOpts = '<option value="">-- Select --</option>' +
      opt('Specific', 'specific', kind, 0, 0) +
      opt('General', 'general', kind, 0, 0);
    for (var i = 0; i < specs.length; i++) {
      var s = specs[i];
      specificOpts += '<option value="' + esc(s.id) + '" ' + selected(row.specific_state === s.id, s.id) + '>' + esc(s.name || s.id) + '</option>';
    }
    for (var j = 0; j < gens.length; j++) {
      var g = gens[j];
      generalOpts += '<option value="' + esc(g.id) + '" ' + selected(row.general_state === g.id, g.id) + '>' + esc(g.name || g.id) + '</option>';
    }
    var selectedGeneral = null;
    for (var gj = 0; gj < gens.length; gj++) {
      if (gens[gj].id === row.general_state) { selectedGeneral = gens[gj]; break; }
    }
    var minShift = selectedGeneral ? selectedGeneral.min_shift : -6;
    var maxShift = selectedGeneral ? selectedGeneral.max_shift : 6;
    if (selectedGeneral && (row.shift_amount === undefined || row.shift_amount === null || row.shift_amount === '')) {
      row.shift_amount = Math.max(minShift, Math.min(maxShift, 1));
    }
    var shiftVal = Number(row.shift_amount) || 0;
    var shiftOptions = '';
    for (var s2 = minShift; s2 <= maxShift; s2++) {
      var abs = Math.abs(s2);
      var shiftCost = selectedGeneral ? selectedGeneral.shift_cost : { add_cost: 0, energy_cost: 0 };
      shiftOptions += '<option value="' + s2 + '" ' + selected(shiftVal === s2, s2) + '>' + s2 + costSuffix(abs * (shiftCost.add_cost || 0), abs * (shiftCost.energy_cost || 0)) + '</option>';
    }
    var kindHidden = '';
    var generalHidden = 'hidden';
    var specificHidden = '';
    if (kind === 'general') { specificHidden = 'hidden'; generalHidden = ''; }
    return '<div class="bg-gray-900 rounded p-3 space-y-2" data-state-row>' +
      '<div class="flex items-center gap-2"><span class="text-xs text-gray-400">State</span>' +
      '<select name="' + esc(name) + '__state_kind" onchange="onStateKindChange(this)" class="bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white">' +
        kindOpts +
      '</select>' +
      '<button type="button" onclick="this.closest(\'[data-state-row]\').remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs ml-auto">− Remove State</button>' +
      '</div>' +
      '<div data-wrap="specific" ' + specificHidden + '>' +
        '<label class="block text-xs text-gray-400 mb-1">Specific State</label>' +
        '<select name="' + esc(name) + '__specific_state" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">' + specificOpts + '</select>' +
      '</div>' +
      '<div data-wrap="general" ' + generalHidden + '>' +
        '<label class="block text-xs text-gray-400 mb-1">General State</label>' +
        '<select name="' + esc(name) + '__general_state" onchange="onStateGeneralChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">' + generalOpts + '</select>' +
      '</div>' +
      '<div data-wrap="shift" ' + generalHidden + '>' +
        '<label class="block text-xs text-gray-400 mb-1">Shift Amount</label>' +
        '<select name="' + esc(name) + '__shift_amount" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">' + shiftOptions + '</select>' +
      '</div>' +
      '</div>';
  }
  window.onStateKindChange = function (sel) {
    var row = sel.closest('[data-state-row]');
    if (!row) return;
    var kind = sel.value;
    var setWrap = function (key, show) {
      var el = row.querySelector('[data-wrap="' + key + '"]');
      if (el) el.hidden = !show;
    };
    setWrap('specific', kind === 'specific');
    setWrap('general', kind === 'general');
    setWrap('shift', kind === 'general');
    recalcAll();
  };
  window.onStateGeneralChange = function (sel) {
    var row = sel.closest('[data-state-row]');
    if (!row) return;
    var card = sel.closest('.section-card');
    var fieldKey = sel.name.replace('__general_state', '');
    var fields = card ? getGenericFieldsForCard(card) : null;
    if (!fields) return;
    var field = null;
    for (var i = 0; i < fields.length; i++) {
      if (fields[i].key === fieldKey) { field = fields[i]; break; }
    }
    if (!field) return;
    var gens = (C.states && C.states.general_states) || [];
    var g = null;
    for (var j = 0; j < gens.length; j++) {
      if (gens[j].id === sel.value) { g = gens[j]; break; }
    }
    if (!g) return;
    var shiftSel = row.querySelector('select[name="' + fieldKey + '__shift_amount"]');
    if (!shiftSel) return;
    shiftSel.innerHTML = '';
    for (var s = g.min_shift; s <= g.max_shift; s++) {
      var abs = Math.abs(s);
      var c = g.shift_cost || { add_cost: 0, energy_cost: 0 };
      var optEl = document.createElement('option');
      optEl.value = s;
      optEl.textContent = s + costSuffix(abs * (c.add_cost || 0), abs * (c.energy_cost || 0));
      shiftSel.appendChild(optEl);
    }
    shiftSel.value = Math.max(g.min_shift, Math.min(g.max_shift, 1));
    recalcAll();
  };
  window.onGenericFieldChange = function (el) {
    var card = el.closest('.section-card');
    if (!card) { recalcAll(); return; }
    var fields = getGenericFieldsForCard(card);
    if (!fields) { recalcAll(); return; }
    var field = null;
    for (var i = 0; i < fields.length; i++) {
      if (fields[i].key === el.name || (el.getAttribute('data-generic-field') === fields[i].key)) { field = fields[i]; break; }
    }
    var values = readGenericCardValues(card);
    for (var k = 0; k < fields.length; k++) {
      var f = fields[k];
      if (f.visibility_when && f.visibility_when === (field ? field.key : el.name)) {
        var wrap = card.querySelector('[data-field-key="' + f.key + '"]');
        if (wrap) wrap.hidden = !isFieldVisible(f, values, fields);
      }
    }
    // For Enact Persistent Effect, the "Applies" dropdown also controls
    // which inline editor body is shown. Swap the body when effect_type
    // changes (the generic effect_type field keeps its value, so cost
    // calc and form submission are unaffected).
    if (card.dataset.enactType === 'Enact Persistent Effect' && el.name === 'effect_type') {
      swapInlineEffectEditor(card);
    }
    recalcAll();
  };
  window.addGenericRow = function (btn, fieldKey) {
    var list = btn.closest('div').parentElement.querySelector('[data-generic-list="' + fieldKey + '"]');
    if (!list) return;
    var card = btn.closest('.section-card');
    var fields = card ? getGenericFieldsForCard(card) : null;
    if (!fields) return;
    var field = null;
    for (var i = 0; i < fields.length; i++) {
      if (fields[i].key === fieldKey) { field = fields[i]; break; }
    }
    if (!field) return;
    var fakeRow = document.createElement('div');
    fakeRow.innerHTML = renderSolutionRow(fieldKey, list.children.length, {}, field);
    list.appendChild(fakeRow.firstChild);
    recalcAll();
  };
  window.addStateRow = function (btn, fieldKey) {
    var list = btn.closest('div').parentElement.querySelector('[data-states-list="' + fieldKey + '"]');
    if (!list) return;
    var card = btn.closest('.section-card');
    var fields = card ? getGenericFieldsForCard(card) : null;
    if (!fields) return;
    var field = null;
    for (var i = 0; i < fields.length; i++) {
      if (fields[i].key === fieldKey) { field = fields[i]; break; }
    }
    if (!field) return;
    var fakeRow = document.createElement('div');
    fakeRow.innerHTML = renderStateRow(fieldKey, list.children.length, {}, field);
    list.appendChild(fakeRow.firstChild);
    recalcAll();
  };

  // Read all generic field values from a card DOM. Returns a map compatible
  // with the Go FieldValueMap shape.
  function readGenericCardValues(card) {
    var fields = card ? getGenericFieldsForCard(card) : null;
    var values = {};
    if (!fields || !card) return values;
    for (var i = 0; i < fields.length; i++) {
      var f = fields[i];
      if (f.type === 'checkbox') {
        var cb = card.querySelector('[data-generic-field="' + f.key + '"]');
        values[f.key] = cb ? cb.checked : false;
      } else if (f.type === 'free_number' || f.type === 'dropdown' || f.type === 'free_text') {
        var sel = card.querySelector('[data-generic-field="' + f.key + '"]');
        if (sel) values[f.key] = sel.value;
      } else if (f.type === 'solutions') {
        var rows = card.querySelectorAll('[data-generic-list="' + f.key + '"] [data-row]');
        var arr = [];
        for (var r = 0; r < rows.length; r++) {
          var row = rows[r];
          var rowVal = {};
          for (var rf = 0; rf < f.row_fields.length; rf++) {
            var rfi = f.row_fields[rf];
            var fullKey = f.key + '__' + rfi.key;
            if (rfi.type === 'checkbox') {
              var cb2 = row.querySelector('[data-generic-field="' + fullKey + '"]');
              rowVal[rfi.key] = cb2 ? cb2.checked : false;
            } else {
              var s2 = row.querySelector('[data-generic-field="' + fullKey + '"]');
              if (s2) {
                if (rfi.type === 'free_number') rowVal[rfi.key] = Number(s2.value);
                else rowVal[rfi.key] = s2.value;
              }
            }
          }
          if (rowVal.value !== '' && rowVal.value !== undefined) {
            arr.push(rowVal);
          }
        }
        values[f.key] = arr;
      } else if (f.type === 'states') {
        var srows = card.querySelectorAll('[data-states-list="' + f.key + '"] [data-state-row]');
        var sarr = [];
        for (var s = 0; s < srows.length; s++) {
          var row = srows[s];
          var kindSel = row.querySelector('select[name="' + f.key + '__state_kind"]');
          var specSel = row.querySelector('select[name="' + f.key + '__specific_state"]');
          var genSel = row.querySelector('select[name="' + f.key + '__general_state"]');
          var shiftSel = row.querySelector('select[name="' + f.key + '__shift_amount"]');
          var rowVal = {
            state_kind: kindSel ? kindSel.value : '',
            specific_state: specSel ? specSel.value : '',
            general_state: genSel ? genSel.value : '',
            shift_amount: shiftSel ? Number(shiftSel.value) : 0
          };
          if (rowVal.state_kind || rowVal.specific_state || rowVal.general_state) {
            sarr.push(rowVal);
          }
        }
        values[f.key] = sarr;
      }
    }
    return values;
  }

  // =========================================================================
  // Generic cost evaluator (mirrors internal/config/field_costs.go). Used for
  // live UI feedback. Server-side evaluation remains authoritative.
  // =========================================================================
  function toBool(v) {
    if (v === true) return true;
    if (typeof v === 'string') return v === 'on' || v === 'true' || v === '1' || v === 'yes';
    return false;
  }
  function toStr(v) {
    if (v === null || v === undefined) return '';
    return String(v);
  }
  function toInt(v) {
    var n = Number(v);
    return isNaN(n) ? 0 : n;
  }
  function stepCostOf(perStep, direction) {
    if (!perStep) return { add: 0, energy: 0 };
    var d = perStep[direction] || { add_cost: 0, energy_cost: 0 };
    return { add: d.add_cost || 0, energy: d.energy_cost || 0 };
  }
  function evalFieldJS(field, raw, allValues) {
    var build = 0, cast = 0;
    if (raw === undefined || raw === null) return { build: build, cast: cast };
    if (raw === '' || raw === false) return { build: build, cast: cast };
    if (Array.isArray(raw) && raw.length === 0 && field.type !== 'solutions') return { build: build, cast: cast };
    switch (field.type) {
      case 'checkbox':
        if (toBool(raw) && field.cost) {
          build += field.cost.add_cost || 0;
          cast += field.cost.energy_cost || 0;
        }
        break;
      case 'dropdown': {
        var sel = toStr(raw);
        if (sel) {
          if (field.cost) {
            build += field.cost.add_cost || 0;
            cast += field.cost.energy_cost || 0;
          }
          var opts = field.options || [];
          for (var i = 0; i < opts.length; i++) {
            if (opts[i].value === sel) {
              if (opts[i].cost) {
                build += opts[i].cost.add_cost || 0;
                cast += opts[i].cost.energy_cost || 0;
              }
              var childVals = (allValues && allValues.__cascadeFor && allValues.__cascadeFor[field.key]) || {};
              var cc = evalFieldsJS(opts[i].fields || [], childVals);
              build += cc.build; cast += cc.cast;
              break;
            }
          }
        }
        break;
      }
      case 'free_text':
        break;
      case 'free_number': {
        var n = toInt(raw);
        var def = toInt(field.default);
        var step = field.step || 1;
        var delta = n - def;
        if (delta !== 0) {
          var dir = delta > 0 ? 'increase' : 'decrease';
          var sc = stepCostOf(field.per_step, dir);
          var steps;
          if (field.rounding === 'ceil') {
            steps = Math.ceil(delta / step);
            if (steps < 0) steps = 0;
          } else if (field.rounding === 'floor') {
            steps = Math.floor(delta / step);
          } else {
            steps = delta / step;
          }
          if (steps !== 0) {
            build += Math.abs(steps) * sc.add;
            cast += Math.abs(steps) * sc.energy;
          }
        }
        break;
      }
      case 'solutions': {
        var arr = Array.isArray(raw) ? raw : [];
        for (var ri = 0; ri < arr.length; ri++) {
          var rv = arr[ri] || {};
          var rc = evalFieldsJS(field.row_fields || [], rv);
          build += rc.build; cast += rc.cast;
        }
        if (field.per_item) {
          var dc = field.default_count || 0;
          var diff = arr.length - dc;
          if (diff !== 0) {
            var pi = diff > 0 ? field.per_item.increase : field.per_item.decrease;
            build += Math.abs(diff) * (pi.add_cost || 0);
            cast += Math.abs(diff) * (pi.energy_cost || 0);
          }
        }
        break;
      }
      case 'states': {
        var rows = Array.isArray(raw) ? raw : [];
        for (var si = 0; si < rows.length; si++) {
          var sr = rows[si] || {};
          if (!sr.state_kind && !sr.specific_state && !sr.general_state) continue;
          var src = evalFieldsJS(field.row_fields || [], sr);
          build += src.build; cast += src.cast;
        }
        break;
      }
    }
    return { build: build, cast: cast };
  }
  function evalFieldsJS(fields, values) {
    var build = 0, cast = 0;
    if (!fields) return { build: build, cast: cast };
    for (var i = 0; i < fields.length; i++) {
      var f = fields[i];
      if (!isFieldVisible(f, values)) continue;
      var raw = values ? values[f.key] : undefined;
      var res = evalFieldJS(f, raw, values);
      build += res.build;
      cast += res.cast;
    }
    return { build: build, cast: cast };
  }
  function evalStateRowJS(row, gens, specs, surcharge) {
    var build = 0, cast = 0;
    if (!row) return { build: build, cast: cast };
    if (row.state_kind === 'general') {
      var gid = row.general_state;
      for (var i = 0; i < gens.length; i++) {
        if (gens[i].id === gid) {
          var abs = Math.abs(Number(row.shift_amount) || 0);
          build += abs * (gens[i].shift_cost ? (gens[i].shift_cost.add_cost || 0) : 0);
          cast += abs * (gens[i].shift_cost ? (gens[i].shift_cost.energy_cost || 0) : 0);
          break;
        }
      }
    } else {
      var sid = row.specific_state;
      for (var j = 0; j < specs.length; j++) {
        if (specs[j].id === sid) {
          build += specs[j].add_cost || 0;
          cast += specs[j].energy_cost || 0;
          break;
        }
      }
    }
    return { build: build, cast: cast };
  }
  function evalStatesSurchargeJS(surcharge, rowCount) {
    if (!surcharge || rowCount <= 1) return { build: 0, cast: 0 };
    var extra = rowCount - 1;
    return { build: extra * (surcharge.add_cost || 0), cast: extra * (surcharge.energy_cost || 0) };
  }
  function genericCalcJS(cfg, values) {
    var base = (cfg && cfg.base_cost) || { add_cost: 0, energy_cost: 0 };
    var b = base.add_cost || 0;
    var c = base.energy_cost || 0;
    var res = evalFieldsJS(cfg && cfg.fields, values);
    return { build: b + res.build, cast: c + res.cast };
  }
  function abilityTypeCalcJS(cfg, values) {
    var base = (cfg && cfg.base_energy) || 0;
    var action = (cfg && cfg.base_action) || 0;
    var res = evalFieldsJS(cfg && cfg.fields, values);
    var totalBuild = res.build;
    var totalEnergy = base + res.cast;
    return { build: totalBuild, energy: totalEnergy, action: action };
  }


  function overview(showResolve) {
    showResolve = showResolve !== false;
    var stats = [
      statCard('Build Cost', '0', 'build'),
      statCard('Cast Cost', '0', 'cast'),
      statCard('Final', '...', 'formula'),
    ];
    if (showResolve) {
      stats.unshift(statCard('Always Resolve', 'No', 'resolve'));
    }
    return [
      '<div class="grid grid-cols-1 md:grid-cols-'+(showResolve ? '4' : '3')+' gap-3 text-sm">',
      stats.join(''),
      '</div>',
    ].join('\n');
  }
  function statCard(label, value, outKey) {
    return '<div class="bg-gray-900 rounded p-3 border border-gray-700">'+
      '<div class="text-gray-400 text-xs uppercase">'+esc(label)+'</div>'+
      '<div class="text-white font-mono text-lg" data-out="'+outKey+'">'+esc(value)+'</div>'+
    '</div>';
  }
  function breakdown() {
    return '<div><div class="text-xs text-gray-400 uppercase mb-1">Cost Breakdown</div>'+
      '<ul data-out="breakdown" class="text-sm text-gray-300 list-disc list-inside space-y-1"></ul></div>';
  }
  function stepSelect(name, label, options, selectedValue, costFn) {
    costFn = costFn || function () { return { add: 0, energy: 0 }; };
    var opts = options.map(function(o){
      var lbl = (o > 0 ? '+' : '') + o;
      var c = costFn(o);
      return '<option value="'+o+'" '+selected(selectedValue, o)+'>'+lbl+costSuffix(c.add, c.energy)+'</option>';
    }).join('');
    return '<div class="flex items-center gap-2">'+
      '<span class="text-sm text-gray-400 whitespace-nowrap">'+esc(label)+'</span>'+
      '<select name="'+name+'" onchange="recalcAll()" class="bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white flex-1">'+opts+'</select>'+
    '</div>';
  }
  function intSelect(name, label, min, max, selectedValue, suffix, costFn) {
    suffix = suffix || 'm';
    costFn = costFn || function () { return { add: 0, energy: 0 }; };
    var opts = '';
    for (var i = min; i <= max; i++) {
      var c = costFn(i);
      opts += '<option value="'+i+'" '+selected(selectedValue, i)+'>'+i+suffix+costSuffix(c.add, c.energy)+'</option>';
    }
    return '<div><label class="block text-xs text-gray-400 mb-1">'+esc(label)+'</label>'+
      '<select name="'+name+'" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select></div>';
  }
  function intSelectFlat(name, label, min, max, selectedValue, costFn) {
    costFn = costFn || function () { return { add: 0, energy: 0 }; };
    var opts = '';
    for (var i = min; i <= max; i++) {
      var c = costFn(i);
      opts += '<option value="'+i+'" '+selected(selectedValue, i)+'>'+i+costSuffix(c.add, c.energy)+'</option>';
    }
    return '<div><label class="block text-xs text-gray-400 mb-1">'+esc(label)+'</label>'+
      '<select name="'+name+'" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select></div>';
  }
  function checkbox(name, label, value, cost) {
    var suffix = cost ? costSuffix(cost.add, cost.energy) : '';
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="'+name+'" onchange="recalcAll()" '+checked(value)+' class="rounded bg-gray-700 border-gray-600">'+
      esc(label)+suffix+
    '</label>';
  }
  function itemDepCheckbox(value, cost) {
    var suffix = cost ? costSuffix(cost.add, cost.energy) : '';
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="item_dep" onchange="onItemDepChange(this)" '+checked(value)+' class="rounded bg-gray-700 border-gray-600">'+
      'Has Item Dependency'+suffix+
    '</label>';
  }
  function traitOptions(category, selectedValue) {
    var list = D.allTraits;
    if (category === 'offense') list = D.offenseTraits;
    if (category === 'defense') list = D.defenseTraits;
    if (category === 'general') list = D.generalTraits;
    return list.map(function(t){
      return '<option value="'+esc(t)+'" '+selected(selectedValue,t)+'>'+esc(t)+'</option>';
    }).join('');
  }

  function traitOptionsGrouped(selectedValue) {
    function group(label, list) {
      var opts = list.map(function(t){
        return '<option value="'+esc(t)+'" data-cat="'+esc(label.toLowerCase())+'" '+selected(selectedValue,t)+'>'+esc(t)+'</option>';
      }).join('');
      return '<optgroup label="'+esc(label)+'">'+opts+'</optgroup>';
    }
    return group('General', D.generalTraits) +
           group('Offense', D.offenseTraits) +
           group('Defense', D.defenseTraits);
  }

  function traitOptionsForCategory(category, selectedValue) {
    if (category === 'general' || category === 'offense' || category === 'defense') {
      var label = category.charAt(0).toUpperCase() + category.slice(1);
      var list = category === 'general' ? D.generalTraits : (category === 'offense' ? D.offenseTraits : D.defenseTraits);
      var opts = list.map(function(t){
        return '<option value="'+esc(t)+'" data-cat="'+esc(category)+'" '+selected(selectedValue,t)+'>'+esc(t)+'</option>';
      }).join('');
      return '<optgroup label="'+esc(label)+'">'+opts+'</optgroup>';
    }
    return traitOptionsGrouped(selectedValue);
  }

  function categoryOfTrait(name) {
    if (!name) return '';
    if (D.generalTraits.indexOf(name) !== -1) return 'general';
    if (D.offenseTraits.indexOf(name) !== -1) return 'offense';
    if (D.defenseTraits.indexOf(name) !== -1) return 'defense';
    return '';
  }

  // =========================================================================
  // Enactment cards
  // =========================================================================

  function legacyStateToRows(data) {
    if (!data) return [{}];
    if (data.states && Array.isArray(data.states)) return data.states;
    return [{
      state_kind: data.state_type || 'specific',
      specific_state: data.specific_state || '',
      general_state: data.general_state || '',
      shift_amount: data.shift_amount !== undefined ? Number(data.shift_amount) : 1
    }];
  }

  function renderEnactCard(type, data) {
    data = data || {};
    var cfg = getEnactConfig(type);
    if (cfg.fields && cfg.fields.length) {
      var hydration = data.fields || data;
      if (type === 'Enact State' && !data.fields) {
        hydration = Object.assign({}, data, { states: legacyStateToRows(data) });
      }
      var fieldsHolder = '<div class="space-y-2">';
      for (var i = 0; i < cfg.fields.length; i++) {
        var visClass = isFieldVisible(cfg.fields[i], hydration, cfg.fields) ? '' : 'hidden';
        fieldsHolder += '<div data-field-key="' + esc(cfg.fields[i].key) + '"' + (visClass ? ' ' + visClass : '') + '>' + renderFieldHTML(cfg.fields[i], hydration) + '</div>';
      }
      fieldsHolder += '</div>';
      var gid = registerGenericFields(cfg.fields);
      // For Enact Persistent Effect, append an inline editor region whose
      // contents swap when the effect_type dropdown changes. The generic
      // effect_type select is kept (for cost calc and form submission); we
      // attach a swap handler to it after render via a delegated listener.
      //
      // The inline editor reads typed fields (source, flat_bonus,
      // offensive_trait, etc.) from `data` directly, not from `data.fields`
      // (which only contains schema fields and omits them).
      var inlineHost = '';
      if (type === 'Enact Persistent Effect') {
        var effectType = (hydration && hydration.effect_type) || '';
        var title = effectType
          ? 'Inline Effect — ' + esc(effectType.replace(/^Enact /, ''))
          : 'Inline Effect';
        inlineHost =
          '<div class="border-t border-indigo-700 pt-3 space-y-2" data-inline-effect-host>'+
            '<div class="flex items-center justify-between">'+
              '<h4 class="text-sm font-semibold text-indigo-300" data-inline-effect-title>'+title+'</h4>'+
              '<span class="text-xs text-gray-500">crafts the effect applied by this persistent effect</span>'+
            '</div>'+
            '<div data-inline-effect-body class="bg-gray-900/40 rounded p-3 space-y-3">'+
              renderInlineEffectEditor(effectType, data, cfg)+
            '</div>'+
          '</div>';
      }
      return '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="'+esc(type)+'" data-generic-id="'+gid+'" data-build="0" data-cast="0">' +
        '<div class="flex items-center justify-between"><h3 class="text-md font-semibold text-indigo-400">Enact — '+esc(type.replace(/^Enact /, ''))+'</h3>' +
        enactTypeSelect(type) +
        '</div>' +
        fieldsHolder +
        inlineHost +
        '</div>';
    }
    if (type === 'Enact Damage') return renderEnactDamage(data, cfg);
    if (type === 'Enact Healing') return renderEnactHealing(data, cfg);
    if (type === 'Enact Movement') return renderEnactMovement(data, cfg);
    if (type === 'Enact Proficiency Shift') return renderEnactProfShift(data, cfg);
    if (type === 'Enact Persistent Effect') return renderEnactPersistent(data, cfg);
    if (type === 'Enact Negation') return renderEnactNegation(data, cfg);
    if (type === 'Enact State') return renderEnactState(data, cfg);
    return '<div class="section-card enact-card bg-gray-800 rounded border border-gray-700 p-4 text-red-400">Unknown enact type: '+esc(type)+'</div>';
  }

  function sourceSelect(d, cfg) {
    cfg = cfg || {};
    var tiers = cfg.dice_tiers || {d4:0,d6:1,d8:2,d10:3,d12:4};
    var tierCost = cfg.dice_tier_cost || {add_cost:0, energy_cost:0};
    var opts = D.damageDiceOptions.map(function(o){
      var tier = tiers[o] || 0;
      return opt(o, o, d.source, tier * (tierCost.add_cost || 0), tier * (tierCost.energy_cost || 0));
    }).join('');
    var traitCost = findPerk(cfg.perks, 'trait_source') || {add_cost:0, energy_cost:0};
    opts += opt('Trait (1d10)', 'trait', d.source, traitCost.add_cost, traitCost.energy_cost);
    var prevCost = findPerk(cfg.perks, 'use_previous') || {add_cost:0, energy_cost:0};
    opts += opt('Use result of previous enactment', 'previous', d.source, prevCost.add_cost, prevCost.energy_cost);
    opts += opt('Another roll result', 'other', d.source, prevCost.add_cost, prevCost.energy_cost);
    return opts;
  }

  function renderEnactDamage(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'offense') : '');
    var traitSelectHTML =
      '<div data-wrap="source-trait" '+hiddenIf(src==='trait')+'>'+
        '<label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="source_category" value="'+esc(srcCat)+'">'+
        '<select name="source_trait" onchange="onSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select>'+
      '</div>';
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Damage" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Damage</h3>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
          '<div>',
            '<label class="block text-xs text-gray-400 mb-1">Source</label>',
            '<select name="source" onchange="onEnactSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
              sourceSelect(d, cfg),
            '</select>',
          '</div>',
          traitSelectHTML,
        '</div>',
        '<div data-wrap="source-other" '+otherWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
          '<input type="text" name="other" value="'+esc(d.other_roll_text||'')+'" placeholder="e.g. previous_enactment.result" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div data-wrap="source-previous" '+prevWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Previous Reference</label>',
          '<input type="text" name="other" value="'+esc(d.other_roll_text||'')+'" placeholder="e.g. previous_enactment.result" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          intSelectFlat('flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0, cumCostFn(0, findPerk(cfg.perks, 'flat_bonus'))),
          '<div><label class="block text-xs text-gray-400 mb-1">Offensive Trait (extra die)</label>',
            '<select name="offense" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              opt('None', '', d.offensive_trait, 0, 0) +
              D.offenseTraits.map(function(t){ var c = perkCost(cfg.perks, 'offensive_trait'); return opt(t, t, d.offensive_trait, c.add, c.energy); }).join('') +
            '</select></div>',
        '</div>',
      '</div>'
    ].join('\n');
  }

  function renderEnactHealing(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'offense') : '');
    var traitSelectHTML =
      '<div data-wrap="source-trait" '+hiddenIf(src==='trait')+'>'+
        '<label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="source_category" value="'+esc(srcCat)+'">'+
        '<select name="source_trait" onchange="onSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select>'+
      '</div>';
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Healing" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Healing</h3>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
          '<div><label class="block text-xs text-gray-400 mb-1">Source</label>',
          '<select name="source" onchange="onEnactSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
            sourceSelect(d, cfg),
          '</select></div>',
          traitSelectHTML,
        '</div>',
        '<div data-wrap="source-other" '+otherWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
          '<input type="text" name="other" value="'+esc(d.other_roll_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div data-wrap="source-previous" '+prevWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Previous Reference</label>',
          '<input type="text" name="other" value="'+esc(d.other_roll_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          intSelectFlat('flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0, cumCostFn(0, findPerk(cfg.perks, 'flat_bonus'))),
          '<div><label class="block text-xs text-gray-400 mb-1">Medicine</label>',
            '<select name="medicine" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              opt('None', '', d.medicine_trait, 0, 0)+
              (function(){ var c = perkCost(cfg.perks, 'medicine_trait'); return opt('Medicine (1d10)', 'Medicine', d.medicine_trait, c.add, c.energy); })() +
            '</select></div>',
        '</div>',
      '</div>'
    ].join('\n');
  }

  function renderEnactMovement(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var dirs = d.directions && d.directions.length ? d.directions : ['Away'];
    var originMode = d.origin_mode || 'engager';
    var otherOrigin = perkCost(cfg.perks, 'other_origin');
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Movement" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Movement</h3>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
          '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
            '<select name="origin_mode" onchange="onOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              opt('Engager', 'engager', originMode, 0, 0)+
              opt('Other Origin', 'other', originMode, otherOrigin.add, otherOrigin.energy)+
            '</select></div>',
        '</div>',
        '<div data-wrap="origin" '+hiddenIf(originMode === 'other')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
          '<input type="text" name="origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          intSelect('distance', 'Distance', 1, 10, d.distance || 1, 'm', cumCostFn(1, cfg.distance_cost)),
          '<div>',
            '<div class="flex items-center justify-between mb-1"><span class="text-xs text-gray-400">Directions</span>',
              '<button type="button" onclick="addDirection(this)" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Direction</button>',
            '</div>',
            directionsList(dirs, cfg),
        '</div>',
      '</div>'
    ].join('\n');
  }
  function directionsList(dirs, cfg) {
    var freeDir = perkCost((cfg||{}).perks, 'free_direction');
    var rows = dirs.map(function(dir){
      var opts = D.directionOptions.map(function(o){
        if (o === 'Free') return opt(o, o, dir, freeDir.add, freeDir.energy);
        return opt(o, o, dir, 0, 0);
      }).join('');
      return '<div class="flex items-center gap-2">'+
        '<select name="direction" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
        '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>'+
      '</div>';
    }).join('');
    return '<div data-list="directions" class="space-y-1">'+rows+'</div>';
  }

  function renderEnactProfShift(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Proficiency Shift" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Proficiency Shift</h3>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          checkbox('always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">',
          '<div><label class="block text-xs text-gray-400 mb-1">Trait</label>',
            '<select name="shifted_trait" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              '<option value="">-- Select --</option>' + traitOptionsGrouped(d.shifted_trait),
            '</select></div>',
          '<div><label class="block text-xs text-gray-400 mb-1">Direction</label>',
            '<select name="shift_dir" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              D.shiftDirectionOptions.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.shift_dir,t)+'>'+esc(t)+'</option>';}).join(''),
            '</select></div>',
          intSelectFlat('shift_amount', 'Amount', 1, 5, d.shift_amount || 1, cumCostFn(1, cfg.shift_amount_cost)),
          intSelectFlat('shift_uses', 'Uses', 1, 5, d.shift_uses || 1, cumCostFn(1, cfg.shift_uses_cost)),
        '</div>',
      '</div>'
    ].join('\n');
  }

  function renderEnactPersistent(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var sols = (d.solutions && d.solutions.length ? d.solutions : ['Dexterity', 'Constitution']);
    var effects = cfg.effects || [];
    var effectByDesc = {};
    effects.forEach(function(e){ effectByDesc[e.description] = e; });
    var effectOpts = D.persistentEffectTypes.map(function(t){
      var e = effectByDesc[t] || { add_cost: 0, energy_cost: 0 };
      return opt(t, t, d.effect_type, e.add_cost, e.energy_cost);
    }).join('');
    var effectType = d.effect_type || '';
    // The inline editor's values are the same typed fields already on the
    // enactment (Source/SourceTrait/FlatBonus/OffensiveTrait/Medicine/Origin
    // /Distance/Directions/Shift*). On the initial render we just feed the
    // incoming data; the per-effect-type cache maintained in
    // onPersistentEffectTypeChange kicks in on subsequent swaps.
    var inlineData = d;
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Persistent Effect" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Persistent Effect</h3>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          '<div><label class="block text-xs text-gray-400 mb-1">Name</label>',
            '<input type="text" name="effect_name" value="'+esc(d.effect_name||'Burning')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white"></div>',
          checkbox('always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          '<div><label class="block text-xs text-gray-400 mb-1">Applies</label>',
            '<select name="effect_type" onchange="onPersistentEffectTypeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              effectOpts +
            '</select></div>',
          intSelectFlat('duration', 'Duration', 2, 8, d.duration || 2, cumCostFn(2, cfg.duration_cost)),
          '<div><label class="block text-xs text-gray-400 mb-1">Trigger</label>',
            '<select name="trigger_timing" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              D.triggerTimings.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.trigger_timing,t)+'>'+esc(t)+'</option>';}).join('') +
            '</select></div>',
        '</div>',
        '<div>',
          '<div class="flex items-center justify-between mb-1"><span class="text-xs text-gray-400 uppercase">Solutions</span>',
            '<button type="button" onclick="addSolution(this)" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Solution</button>',
          '</div>',
          solutionsList(sols),
        '</div>',
        '<div class="border-t border-indigo-700 pt-3 space-y-2" data-inline-effect-host>',
          '<div class="flex items-center justify-between">',
            '<h4 class="text-sm font-semibold text-indigo-300" data-inline-effect-title>'+
              (effectType ? ('Inline Effect — ' + esc(effectType.replace(/^Enact /, ''))) : 'Inline Effect') +
            '</h4>',
            '<span class="text-xs text-gray-500">crafts the effect applied by this persistent effect</span>',
          '</div>',
          '<div data-inline-effect-body class="bg-gray-900/40 rounded p-3 space-y-3">',
            renderInlineEffectEditor(effectType, inlineData, cfg),
          '</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }

  // Maps persistent-effect "Applies" type to the corresponding standalone
  // enactment type whose inner fields the inline editor mirrors.
  var PERSISTENT_TO_ENACT = {
    'Enact Damage': 'Enact Damage',
    'Enact Healing': 'Enact Healing',
    'Enact Movement': 'Enact Movement',
    'Enact Proficiency Shift': 'Enact Proficiency Shift'
  };

  // Returns the inner HTML for the inline effect editor. All input `name=`
  // attributes are prefixed with `effect_` to avoid colliding with the
  // persistent-effect card's own fields, and to make server-side parsing
  // unambiguous. Field options, perks, and cost helpers are the same as the
  // standalone enactments, so dropdowns and costs match 1:1.
  function renderInlineEffectEditor(effectType, d, cfg) {
    d = d || {};
    cfg = cfg || {};
    if (!effectType || !PERSISTENT_TO_ENACT[effectType]) {
      return '<p class="text-xs text-gray-500 italic">Select an effect type above to configure its inline editor.</p>';
    }
    var enactCfg = getEnactConfig(effectType);
    var t = effectType;
    if (t === 'Enact Damage')  return renderInlineDamageBody(d, enactCfg);
    if (t === 'Enact Healing') return renderInlineHealingBody(d, enactCfg);
    if (t === 'Enact Movement') return renderInlineMovementBody(d, enactCfg);
    if (t === 'Enact Proficiency Shift') return renderInlineProfShiftBody(d, enactCfg);
    return '';
  }

  function renderInlineDamageBody(d, cfg) {
    cfg = cfg || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'offense') : '');
    var traitSelectHTML =
      '<div data-wrap="effect-source-trait" '+hiddenIf(src==='trait')+'>'+
        '<label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="effect_source_category" value="'+esc(srcCat)+'">'+
        '<select name="effect_source_trait" onchange="onInlineEffectSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select>'+
      '</div>';
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        checkbox('effect_always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
        '<div>',
          '<label class="block text-xs text-gray-400 mb-1">Source</label>',
          '<select name="effect_source" onchange="onInlineEffectSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
            sourceSelect(d, cfg),
          '</select>',
        '</div>',
        traitSelectHTML,
      '</div>',
      '<div data-wrap="effect-source-other" '+otherWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
        '<input type="text" name="effect_other" value="'+esc(d.other_roll_text||'')+'" placeholder="e.g. previous_enactment.result" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
      '</div>',
      '<div data-wrap="effect-source-previous" '+prevWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Previous Reference</label>',
        '<input type="text" name="effect_other" value="'+esc(d.other_roll_text||'')+'" placeholder="e.g. previous_enactment.result" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
        intSelectFlat('effect_flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0, cumCostFn(0, findPerk(cfg.perks, 'flat_bonus'))),
        '<div><label class="block text-xs text-gray-400 mb-1">Offensive Trait (extra die)</label>',
          '<select name="effect_offense" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            opt('None', '', d.offensive_trait, 0, 0) +
            D.offenseTraits.map(function(x){ var c = perkCost(cfg.perks, 'offensive_trait'); return opt(x, x, d.offensive_trait, c.add, c.energy); }).join('') +
          '</select></div>',
      '</div>',
    ].join('\n');
  }

  function renderInlineHealingBody(d, cfg) {
    cfg = cfg || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'offense') : '');
    var traitSelectHTML =
      '<div data-wrap="effect-source-trait" '+hiddenIf(src==='trait')+'>'+
        '<label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="effect_source_category" value="'+esc(srcCat)+'">'+
        '<select name="effect_source_trait" onchange="onInlineEffectSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select>'+
      '</div>';
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        checkbox('effect_always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
        '<div><label class="block text-xs text-gray-400 mb-1">Source</label>',
        '<select name="effect_source" onchange="onInlineEffectSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
          sourceSelect(d, cfg),
        '</select></div>',
        traitSelectHTML,
      '</div>',
      '<div data-wrap="effect-source-other" '+otherWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
        '<input type="text" name="effect_other" value="'+esc(d.other_roll_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
      '</div>',
      '<div data-wrap="effect-source-previous" '+prevWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Previous Reference</label>',
        '<input type="text" name="effect_other" value="'+esc(d.other_roll_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
        intSelectFlat('effect_flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0, cumCostFn(0, findPerk(cfg.perks, 'flat_bonus'))),
        '<div><label class="block text-xs text-gray-400 mb-1">Medicine</label>',
          '<select name="effect_medicine" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            opt('None', '', d.medicine_trait, 0, 0)+
            (function(){ var c = perkCost(cfg.perks, 'medicine_trait'); return opt('Medicine (1d10)', 'Medicine', d.medicine_trait, c.add, c.energy); })() +
          '</select></div>',
      '</div>',
    ].join('\n');
  }

  function renderInlineMovementBody(d, cfg) {
    cfg = cfg || {};
    var dirs = d.directions && d.directions.length ? d.directions : ['Away'];
    var originMode = d.origin_mode || 'engager';
    var otherOrigin = perkCost(cfg.perks, 'other_origin');
    return [
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        checkbox('effect_always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
        '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
          '<select name="effect_origin_mode" onchange="onInlineEffectOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            opt('Engager', 'engager', originMode, 0, 0)+
            opt('Other Origin', 'other', originMode, otherOrigin.add, otherOrigin.energy)+
          '</select></div>',
      '</div>',
      '<div data-wrap="effect-origin" '+hiddenIf(originMode === 'other')+'>',
        '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
        '<input type="text" name="effect_origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
        intSelect('effect_distance', 'Distance', 1, 10, d.distance || 1, 'm', cumCostFn(1, cfg.distance_cost)),
        '<div>',
          '<div class="flex items-center justify-between mb-1"><span class="text-xs text-gray-400">Directions</span>',
            '<button type="button" onclick="addInlineEffectDirection(this)" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Direction</button>',
          '</div>',
          inlineEffectDirectionsList(dirs, cfg),
      '</div>',
    ].join('\n');
  }
  function inlineEffectDirectionsList(dirs, cfg) {
    var freeDir = perkCost((cfg||{}).perks, 'free_direction');
    var rows = dirs.map(function(dir){
      var opts = D.directionOptions.map(function(o){
        if (o === 'Free') return opt(o, o, dir, freeDir.add, freeDir.energy);
        return opt(o, o, dir, 0, 0);
      }).join('');
      return '<div class="flex items-center gap-2">'+
        '<select name="effect_direction" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
        '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>'+
      '</div>';
    }).join('');
    return '<div data-list="effect-directions" class="space-y-1">'+rows+'</div>';
  }

  function renderInlineProfShiftBody(d, cfg) {
    cfg = cfg || {};
    return [
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
        checkbox('effect_always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">',
        '<div><label class="block text-xs text-gray-400 mb-1">Trait</label>',
          '<select name="effect_shifted_trait" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            '<option value="">-- Select --</option>' + traitOptionsGrouped(d.shifted_trait),
          '</select></div>',
        '<div><label class="block text-xs text-gray-400 mb-1">Direction</label>',
          '<select name="effect_shift_dir" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            D.shiftDirectionOptions.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.shift_dir,t)+'>'+esc(t)+'</option>';}).join(''),
          '</select></div>',
        intSelectFlat('effect_shift_amount', 'Amount', 1, 5, d.shift_amount || 1, cumCostFn(1, cfg.shift_amount_cost)),
        intSelectFlat('effect_shift_uses', 'Uses', 1, 5, d.shift_uses || 1, cumCostFn(1, cfg.shift_uses_cost)),
      '</div>',
    ].join('\n');
  }

  // Swap the inline effect editor when the "Applies" dropdown changes.
  // Caches the entered data per chosen effect_type on the card dataset so
  // switching back to a previously edited effect restores its values.
  // Accepts either a <select> element (legacy non-generic renderer) or a
  // card element (generic renderer) for flexibility.
  window.onPersistentEffectTypeChange = function (sel) {
    var card = sel.closest ? sel.closest('.section-card') : sel;
    if (!card) return;
    return swapInlineEffectEditor(card);
  };

  function swapInlineEffectEditor(card) {
    var sel = card.querySelector('[name="effect_type"]');
    if (!sel) return;
    var effectType = sel.value;
    var body = card.querySelector('[data-inline-effect-body]');
    var title = card.querySelector('[data-inline-effect-title]');
    if (!body) return;
    // Persist current inline values back into the cache for the previously
    // selected effect type, so switching back later restores them.
    var prev = card.dataset.effectType || '';
    if (prev) {
      var prevCache = readInlineEffectData(card, prev);
      var allCache = readInlineEffectCache(card);
      allCache[prev] = prevCache;
      card.dataset.effectCache = JSON.stringify(allCache);
    }
    var cache = readInlineEffectCache(card);
    var data = (effectType && cache[effectType]) || {};
    // The inline editor mirrors the chosen enactment's config (Damage,
    // Healing, Movement, Proficiency Shift) so its dropdowns and perk costs
    // match the standalone enactments 1:1.
    var inlineCfg = effectType ? getEnactConfig(effectType) : {};
    body.innerHTML = renderInlineEffectEditor(effectType, data, inlineCfg);
    if (title) {
      title.textContent = effectType
        ? 'Inline Effect — ' + effectType.replace(/^Enact /, '')
        : 'Inline Effect';
    }
    card.dataset.effectType = effectType;
    recalcAll();
  };

  // Read currently-entered values from the inline effect editor for the
  // given (or current) effect type.
  function readInlineEffectData(card, effectType) {
    effectType = effectType || (card.querySelector('[name="effect_type"]') || {}).value || '';
    var out = { effect_type: effectType };
    if (!effectType) return out;
    function v(name) { var el = card.querySelector('[name="'+name+'"]'); return el ? el.value : ''; }
    function b(name) { var el = card.querySelector('[name="'+name+'"]'); return el ? !!el.checked : false; }
    out.always = b('effect_always');
    if (effectType === 'Enact Damage' || effectType === 'Enact Healing') {
      out.source = v('effect_source');
      out.source_trait = v('effect_source_trait');
      out.source_category = v('effect_source_category');
      out.other_roll_text = v('effect_other');
      out.flat_bonus = Number(v('effect_flat')) || 0;
      if (effectType === 'Enact Damage') {
        out.offensive_trait = v('effect_offense');
      } else {
        out.medicine_trait = v('effect_medicine');
      }
    } else if (effectType === 'Enact Movement') {
      out.origin_mode = v('effect_origin_mode');
      out.origin_text = v('effect_origin_text');
      out.distance = Number(v('effect_distance')) || 1;
      out.directions = Array.from(card.querySelectorAll('[name="effect_direction"]')).map(function(s){return s.value;}).filter(Boolean);
    } else if (effectType === 'Enact Proficiency Shift') {
      out.shifted_trait = v('effect_shifted_trait');
      out.shift_dir = v('effect_shift_dir');
      out.shift_amount = Number(v('effect_shift_amount')) || 1;
      out.shift_uses = Number(v('effect_shift_uses')) || 1;
    }
    return out;
  }
  function readInlineEffectCache(card) {
    try { return JSON.parse(card.dataset.effectCache || '{}') || {}; }
    catch (e) { return {}; }
  }

  window.addInlineEffectDirection = function (btn) {
    var list = btn.parentElement.parentElement.querySelector('[data-list="effect-directions"]');
    if (!list) return;
    var first = list.querySelector('select[name="effect_direction"]');
    var val = first ? first.value : 'Away';
    var card = btn.closest('.section-card');
    // The Directions UI only appears when the chosen effect is Movement, so
    // use the Movement config (same as the standalone Enact Movement card).
    var cfg = getEnactConfig('Enact Movement');
    var freeDir = perkCost(cfg.perks, 'free_direction');
    var opts = D.directionOptions.map(function(o){
      if (o === 'Free') return opt(o, o, val, freeDir.add, freeDir.energy);
      return opt(o, o, val, 0, 0);
    }).join('');
    var row = document.createElement('div');
    row.className = 'flex items-center gap-2';
    row.innerHTML =
      '<select name="effect_direction" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
      '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>';
    list.appendChild(row);
    recalcAll();
  };
  function solutionsList(sols) {
    var rows = sols.map(function(s){
      var opts = '<option value="">-- Select --</option>' + traitOptionsGrouped(s);
      return '<div class="flex items-center gap-2">'+
        '<select name="solution" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
        '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>'+
      '</div>';
    }).join('');
    return '<div data-list="solutions" class="space-y-1">'+rows+'</div>';
  }

  function renderEnactNegation(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'defense') : '');
    var traitSelectHTML =
      '<div data-wrap="source-trait" '+hiddenIf(src==='trait')+'>'+
        '<label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="source_category" value="'+esc(srcCat)+'">'+
        '<select name="source_trait" onchange="onSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select>'+
      '</div>';
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Negation" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Negation</h3>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve', d.always, perkCost(cfg.perks, 'always_resolve')),
          '<div>',
            '<label class="block text-xs text-gray-400 mb-1">Source</label>',
            '<select name="source" onchange="onEnactSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
              sourceSelect(d, cfg),
            '</select>',
          '</div>',
          traitSelectHTML,
        '</div>',
        '<div data-wrap="source-other" '+otherWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
          '<input type="text" name="other" value="'+esc(d.other_roll_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div data-wrap="source-previous" '+prevWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Previous Reference</label>',
          '<input type="text" name="other" value="'+esc(d.other_roll_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          intSelectFlat('flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0, cumCostFn(0, findPerk(cfg.perks, 'flat_bonus'))),
          '<div><label class="block text-xs text-gray-400 mb-1">Defensive Trait (extra die)</label>',
            '<select name="offense" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              opt('None', '', d.offensive_trait, 0, 0) +
              D.defenseTraits.map(function(t){ var c = perkCost(cfg.perks, 'offensive_trait'); return opt(t, t, d.offensive_trait, c.add, c.energy); }).join('') +
            '</select></div>',
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">',
          checkbox('counter_negation', 'Apply Negation to the counter roll', d.counter_negation, perkCost(cfg.perks, 'counter_negation')),
          checkbox('full_counter', 'Ability hits the Engager instead (Engager rolls counter)', d.full_counter, perkCost(cfg.perks, 'full_counter')),
        '</div>',
      '</div>'
    ].join('\n');
  }

  function renderEnactState(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact State" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400 flex items-center gap-2">Enact — State <span class="text-xs text-yellow-400">(WIP)</span></h3>',
        '<p class="text-sm text-gray-400">Applies a state or condition to a target. The Enact State rules are still being finalised — this card is a placeholder.</p>',
        '<div>',
          '<label class="block text-xs text-gray-400 mb-1">State Name</label>',
          '<input type="text" name="effect_name" value="'+esc(d.effect_name||'')+'" placeholder="e.g. Stunned" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
      '</div>'
    ].join('\n');
  }

  // =========================================================================
  // Interaction cards
  // =========================================================================

  function renderInterCard(type, data) {
    data = data || {};
    var cfg = getInterConfig(type);
    if (cfg.fields && cfg.fields.length) {
      var hydration = data.fields || data;
      var fieldsHolder = '<div class="space-y-2">';
      for (var i = 0; i < cfg.fields.length; i++) {
        var visClass = isFieldVisible(cfg.fields[i], hydration, cfg.fields) ? '' : 'hidden';
        fieldsHolder += '<div data-field-key="' + esc(cfg.fields[i].key) + '"' + (visClass ? ' ' + visClass : '') + '>' + renderFieldHTML(cfg.fields[i], hydration) + '</div>';
      }
      fieldsHolder += '</div>';
      var gid = registerGenericFields(cfg.fields);
      return '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="'+esc(type)+'" data-generic-id="'+gid+'" data-build="0" data-cast="0">' +
        '<div class="flex items-center justify-between"><h4 class="text-sm font-semibold text-cyan-300">Interaction — '+esc(type)+'</h4>' +
        interTypeSelect(type) +
        '</div>' +
        fieldsHolder +
        '</div>';
    }
    if (type === 'Self')         return renderInterSelf(data, cfg);
    if (type === 'Direct')       return renderInterDirect(data, cfg);
    if (type === 'Ranged')       return renderInterRanged(data, cfg);
    if (type === 'Area')         return renderInterArea(data, cfg);
    if (type === 'Area of Effect') return renderInterAoE(data, cfg);
    return '<div class="section-card inter-card bg-gray-800 rounded border border-gray-700 p-4 text-red-400">Unknown inter type: '+esc(type)+'</div>';
  }

  function interTypeSelect(selectedValue) {
    return '<label class="text-xs text-gray-400 flex items-center gap-1">Interaction: ' +
      '<select onchange="onInterTypeChange(this)" class="inter-type-select bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white">' +
        '<option value="">-- Select --</option>' +
        INTER_TYPES.map(function(t){return '<option value="'+esc(t)+'" '+selected(selectedValue,t)+'>'+esc(t)+'</option>';}).join('') +
      '</select>' +
    '</label>';
  }

  function enactTypeSelect(selectedValue) {
    return '<label class="text-xs text-gray-400 flex items-center gap-1">Enactment: ' +
      '<select onchange="onEnactSwapType(this)" class="enact-type-swap bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white">' +
        '<option value="">-- Select --</option>' +
        ENACT_TYPES.map(function(t){return '<option value="'+esc(t)+'" '+selected(selectedValue,t)+'>'+esc(t)+'</option>';}).join('') +
      '</select>' +
    '</label>';
  }

  function renderInterSelf(d, cfg) {
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Self" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Self</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTypeSelect('Self'),
          '<div class="text-sm text-gray-300"><strong>Type =</strong> Self + <strong>Target =</strong> Self + <strong>Counter =</strong> d8</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }

  function renderInterDirect(d, cfg) {
    cfg = cfg || {};
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Direct" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Direct</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTypeSelect('Direct'),
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
            intSelect('range', 'Range', 1, 10, d.range || 1, 'm', cumCostFn(1, cfg.range_cost)),
            intSelectFlat('targets', 'Targets', 1, 5, d.targets || 1, cumCostFn(1, cfg.target_cost)),
          '</div>',
          '<div>'+usePrevCheck(d.use_previous, perkCost(cfg.perks, 'use_previous'))+'</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }
  function renderInterRanged(d, cfg) {
    cfg = cfg || {};
    var ext = cfg.range_extension_cost || {add_cost:0, energy_cost:0, step:2};
    var step = ext.step || 2;
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Ranged" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Ranged</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTypeSelect('Ranged'),
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
            intSelect('range', 'Range', 10, 20, d.range || 10, 'm', function(i){
              var inc = Math.floor(Math.max(0, i - 10) / step);
              return { add: inc * (ext.add_cost || 0), energy: inc * (ext.energy_cost || 0) };
            }),
            intSelectFlat('targets', 'Targets', 1, 5, d.targets || 1, cumCostFn(1, cfg.target_cost)),
          '</div>',
          '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">',
            checkbox('visible', 'Target may be not visible', d.visible_ok, perkCost(cfg.perks, 'not_visible')),
            checkbox('obstructed', 'Target may be obstructed', d.obstructed_ok, perkCost(cfg.perks, 'obstructed')),
            checkbox('remove_penalty', 'Remove engagement penalty', d.remove_penalty, perkCost(cfg.perks, 'remove_penalty')),
          '</div>',
          '<div>'+usePrevCheck(d.use_previous, perkCost(cfg.perks, 'use_previous'))+'</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }
  function renderInterArea(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var om = d.origin_mode || 'engager';
    var rangeCost = cfg.range_cost || {add_cost:0, energy_cost:0, step:2};
    var step = rangeCost.step || 2;
    var otherOrigin = perkCost(cfg.perks, 'other_origin');
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Area" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Area</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTypeSelect('Area'),
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
            intSelect('radius', 'Radius', 1, 6, d.radius || 1, 'm', cumCostFn(1, cfg.radius_cost)),
            intSelect('range', 'Range', 0, 10, d.range || 0, 'm', function(i){
              var rng = Math.ceil(Math.max(0, i) / step);
              return { add: rng * (rangeCost.add_cost || 0), energy: rng * (rangeCost.energy_cost || 0) };
            }),
          '</div>',
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
            '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
              '<select name="origin_mode" onchange="onOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
                opt('Engager', 'engager', om, 0, 0)+
                opt('Other Origin', 'other', om, otherOrigin.add, otherOrigin.energy)+
              '</select></div>',
            '<div data-wrap="origin" '+hiddenIf(om==='other')+'>',
              '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
              '<input type="text" name="origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
            '</div>',
          '</div>',
          '<div>'+usePrevCheck(d.use_previous, perkCost(cfg.perks, 'use_previous'))+'</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }
  function renderInterAoE(d, cfg) {
    d = d || {};
    cfg = cfg || {};
    var om = d.origin_mode || 'engager';
    var rangeCost = cfg.range_cost || {add_cost:0, energy_cost:0, step:2};
    var step = rangeCost.step || 2;
    var otherOrigin = perkCost(cfg.perks, 'other_origin');
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Area of Effect" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Area of Effect</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTypeSelect('Area of Effect'),
          '<div class="grid grid-cols-1 md:grid-cols-3 gap-3">',
            intSelect('radius', 'Radius', 1, 6, d.radius || 1, 'm', cumCostFn(1, cfg.radius_cost)),
            intSelect('range', 'Range', 0, 10, d.range || 0, 'm', function(i){
              var rng = Math.ceil(Math.max(0, i) / step);
              return { add: rng * (rangeCost.add_cost || 0), energy: rng * (rangeCost.energy_cost || 0) };
            }),
            intSelectFlat('duration', 'Duration', 2, 6, d.duration || 2, cumCostFn(2, cfg.duration_cost)),
          '</div>',
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
            '<div><label class="block text-xs text-gray-400 mb-1">Trigger Timing</label>',
              '<select name="timing" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
                D.aoeTriggerTimings.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.timing,t)+'>'+esc(t)+'</option>';}).join(''),
              '</select></div>',
            '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
              '<select name="origin_mode" onchange="onOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
                opt('Engager', 'engager', om, 0, 0)+
                opt('Other Origin', 'other', om, otherOrigin.add, otherOrigin.energy)+
              '</select></div>',
          '</div>',
          '<div data-wrap="origin" '+hiddenIf(om!=='other')+'>',
            '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
            '<input type="text" name="origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
          '</div>',
          checkbox('immune', 'Engager is immune', d.immune, perkCost(cfg.perks, 'immune')),
          '<div>'+usePrevCheck(d.use_previous, perkCost(cfg.perks, 'use_previous'))+'</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }
  function usePrevCheck(v, cost) {
    var suffix = cost ? costSuffix(cost.add, cost.energy) : ' (costs extra)';
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="use_previous" onchange="recalcAll()" '+checked(v)+' class="rounded bg-gray-700 border-gray-600">'+
      'Use result of previous interaction/validation'+suffix+
    '</label>';
  }

  // =========================================================================
  // Validation card
  // =========================================================================

  function renderValidationCard(d) {
    d = d || {};
    var cfg = getValidationConfig();
    if (cfg.fields && cfg.fields.length) {
      var hydration = d.fields || d;
      var fieldsHolder = '<div class="space-y-2">';
      for (var i = 0; i < cfg.fields.length; i++) {
        var visClass = isFieldVisible(cfg.fields[i], hydration, cfg.fields) ? '' : 'hidden';
        fieldsHolder += '<div data-field-key="' + esc(cfg.fields[i].key) + '"' + (visClass ? ' ' + visClass : '') + '>' + renderFieldHTML(cfg.fields[i], hydration) + '</div>';
      }
      fieldsHolder += '</div>';
      var gid = registerGenericFields(cfg.fields);
      return '<div class="section-card validation-card bg-gray-800 rounded-lg border border-rose-700 p-4 space-y-3" data-section="validation" data-generic-id="'+gid+'" data-build="0" data-cast="0">' +
        '<h4 class="text-sm font-semibold text-rose-300">Validation</h4>' +
        fieldsHolder +
        '</div>';
    }
    var mode = d.engage_mode || 'trait';
    var cat = d.engage_trait_category || (d.engage_trait ? (categoryOfTrait(d.engage_trait) || 'offense') : 'offense');
    var counters = d.counter_entries || d.counter_rolls || [];

    var modes = (cfg.engagement && cfg.engagement.modes) || [];
    var modeOpts = ['trait','generic','other','previous'].map(function(m){
      var c = findPerk(modes, m) || { add_cost: 0, energy_cost: 0 };
      var label = { trait: 'Trait Roll', generic: 'Generic Roll', other: 'Another roll result', previous: 'Use result of previous interaction' }[m];
      return opt(label, m, mode, c.add_cost, c.energy_cost);
    }).join('');

    var dieTiers = { d6: 0, d8: 1, d10: 2, d12: 3 };
    var engageUp = findPerk((cfg.counter && cfg.counter.tier_shifts) || [], 'engage_up') || { add_cost: 0, energy_cost: 0 };
    var dieOpts = D.genericDieOptions.map(function(o){
      var tier = dieTiers[o] || 0;
      return opt(o, o, d.engage_die, tier * (engageUp.add_cost || 0), tier * (engageUp.energy_cost || 0));
    }).join('');

    return [
      '<div class="section-card validation-card bg-gray-800 rounded-lg border border-rose-700 p-4 space-y-3" data-section="validation" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-rose-300">Validation</h4>',
          '<span class="collapse-chevron text-rose-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
          '<div><label class="block text-xs text-gray-400 mb-1">Engage Roll Type</label>',
            '<select name="engage_mode" onchange="onEngageModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              modeOpts +
            '</select></div>',
        '</div>',
        '<div data-wrap="engage-trait" '+hiddenIf(mode==='trait')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Trait</label>',
          '<input type="hidden" name="engage_trait_category" value="'+esc(cat)+'">',
          '<select name="engage_trait" onchange="onEngageTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            '<option value="">-- Select --</option>' + traitOptionsGrouped(d.engage_trait) +
          '</select></div>',
        '<div data-wrap="engage-generic" '+hiddenIf(mode==='generic')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Die</label>',
          '<select name="engage_die" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            dieOpts +
          '</select></div>',
        '<div data-wrap="engage-other" '+hiddenIf(mode==='other')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
          '<input type="text" name="engage_other" value="'+esc(d.engage_other||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div data-wrap="engage-previous" '+hiddenIf(mode==='previous')+'>',
          '<p class="text-xs text-yellow-400">Engagement roll = result of previous interaction/validation (costs extra)</p>',
        '</div>',
        '<div>',
          '<div class="flex items-center justify-between mb-1">',
            '<span class="text-xs text-gray-400 uppercase">Counter Rolls</span>',
            '<button type="button" onclick="addCounter(this)" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Counter</button>',
          '</div>',
          counterList(counters, cfg),
        '</div>',
        '</div>',
      '</div>'
    ].join('\n');
  }
  function counterList(items, cfg) {
    items = items.length ? items : [{type:'defense', trait:'Reflex'}, {type:'defense', trait:'Constitution'}];
    var rows = items.map(function(item){
      if (typeof item === 'string') item = {type:'defense', trait:item};
      return counterRow(item, cfg);
    }).join('');
    return '<div data-list="counters" class="space-y-2">'+rows+'</div>';
  }
  function counterRow(item, cfg) {
    cfg = cfg || {};
    var types = (cfg.counter && cfg.counter.types) || [];
    var typeDefs = {
      defense: 'Defensive Trait (default)',
      general: 'General Trait',
      offense: 'Offensive Trait',
      previous: 'Use result of previous'
    };
    var opts = ['defense','general','offense','previous'].map(function(t){
      var c = findPerk(types, t) || { add_cost: 0, energy_cost: 0 };
      return opt(typeDefs[t] || t, t, item.type, c.add_cost, c.energy_cost);
    }).join('');
    var traitSelect = '';
    if (item.type === 'defense' || item.type === 'general' || item.type === 'offense') {
      var cat = item.type === 'defense' ? 'defense' : item.type === 'general' ? 'general' : 'offense';
      traitSelect = '<select name="counter_trait" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
        '<option value="">-- Select --</option>'+traitOptions(cat, item.trait)+
      '</select>';
    } else {
      traitSelect = '<input type="text" name="counter_trait" value="'+esc(item.trait||'')+'" placeholder="(reference)" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">';
    }
    return '<div class="flex items-center gap-2">'+
      '<select name="counter_type" onchange="onCounterTypeChange(this)" class="bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
      traitSelect +
      '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>'+
    '</div>';
  }

  // =========================================================================
  // Logic: change handlers
  // =========================================================================

  window.onAbilityTypeChange = function (val) {
    document.getElementById('hidden-ability-type').value = val;
    var host = document.getElementById('ability-type-card-container');
    host.innerHTML = '';
    if (!val) return;
    var init = (initialState && initialState.ability_type === val) ? initialState : {};
    host.innerHTML = renderAbilityTypeCard(val, init);
    recalcAll();
  };

  window.toggleCollapse = function (btn) {
    var card = btn.closest('.section-card, .enactment-block, section');
    if (!card) return;
    var body = card.querySelector('.collapsible-content');
    if (!body) return;
    var expanded = btn.getAttribute('aria-expanded') !== 'false';
    if (expanded) {
      body.setAttribute('hidden', '');
      btn.setAttribute('aria-expanded', 'false');
    } else {
      body.removeAttribute('hidden');
      btn.setAttribute('aria-expanded', 'true');
    }
  };

  window.onReactionTriggerChange = function (sel) {
    var wrap = sel.closest('.section-card').querySelector('[data-wrap="trigger-trait"]');
    if (wrap) wrap.hidden = (sel.value !== 'Target makes a trait check');
    recalcAll();
  };

  window.onNoKnockoutChange = function (cb) {
    var wrap = cb.closest('.section-card').querySelector('[data-wrap="knockouts"]');
    if (wrap) wrap.hidden = cb.checked;
    recalcAll();
  };

  window.onItemDepChange = function (cb) {
    var wrap = cb.closest('.section-card').querySelector('[data-wrap="item-name"]');
    if (wrap) wrap.hidden = !cb.checked;
    recalcAll();
  };

  window.onKnockoutChange = function (sel) {
    var block = sel.closest('.enactment-block') || sel.closest('.section-card');
    var noKnockoutCb = block.querySelector('[name="no_knockout"]');
    var wrap = block.querySelector('[data-wrap="knockouts"]');
    if (sel.value === 'None') {
      if (noKnockoutCb) noKnockoutCb.checked = true;
      if (wrap) wrap.hidden = true;
    } else {
      if (noKnockoutCb) noKnockoutCb.checked = false;
      if (wrap) wrap.hidden = false;
    }
    recalcAll();
  };

  window.onEnactTypeChange = function (sel) {
    var block = sel.closest('.enactment-block');
    var host = block.querySelector('.enact-card-container');
    var val = sel.value;
    var cur = readCardData(host);
    host.innerHTML = '';
    block.dataset.enactType = val;
    if (!val) return;
    host.innerHTML = renderEnactCard(val, cur);
    recalcAll();
  };

  function readCardData(container) {
    if (!container) return {};
    var data = {};
    var inputs = container.querySelectorAll('input,select,textarea');
    inputs.forEach(function(el) {
      if (!el.name) return;
      if (el.type === 'checkbox') {
        data[el.name] = el.checked;
      } else if (el.tagName === 'SELECT' && el.multiple) {
        var vals = [];
        Array.from(el.selectedOptions).forEach(function(o){ vals.push(o.value); });
        data[el.name] = vals;
      } else {
        data[el.name] = el.value;
      }
    });
    return data;
  }

  window.onInterTypeChange = function (sel) {
    var block = sel.closest('.enactment-block');
    var host = block.querySelector('.inter-card-container');
    var val = sel.value;
    var cur = readCardData(host);
    host.innerHTML = '';
    if (!val) return;
    host.innerHTML = renderInterCard(val, cur);
    recalcAll();
  };

  window.onEnactSwapType = function (sel) {
    var block = sel.closest('.enactment-block');
    var host = block.querySelector('.enact-card-container');
    var val = sel.value;
    var cur = readCardData(host);
    host.innerHTML = '';
    if (!val) return;
    block.dataset.enactType = val;
    if (block.querySelector('.enact-type-select')) {
      block.querySelector('.enact-type-select').value = val;
    }
    host.innerHTML = renderEnactCard(val, cur);
    recalcAll();
  };

  window.onEnactSourceChange = function (sel) {
    var c = sel.closest('.section-card');
    var v = sel.value;
    setWrap(c, 'source-trait', v === 'trait');
    setWrap(c, 'source-other', v === 'other');
    setWrap(c, 'source-previous', v === 'previous');
    recalcAll();
  };

  // Inline-effect equivalents of the change handlers. These mirror the
  // standalone ones but operate on the `effect_`-prefixed name and wrap
  // attributes used by the persistent-effect inline editor.
  window.onInlineEffectSourceChange = function (sel) {
    var c = sel.closest('.section-card');
    var v = sel.value;
    setWrap(c, 'effect-source-trait', v === 'trait');
    setWrap(c, 'effect-source-other', v === 'other');
    setWrap(c, 'effect-source-previous', v === 'previous');
    recalcAll();
  };

  window.onSourceTraitChange = function (sel) {
    var c = sel.closest('.section-card');
    var hidden = c.querySelector('[name="source_category"]');
    if (hidden) hidden.value = categoryOfTrait(sel.value);
    recalcAll();
  };

  window.onInlineEffectSourceTraitChange = function (sel) {
    var c = sel.closest('.section-card');
    var hidden = c.querySelector('[name="effect_source_category"]');
    if (hidden) hidden.value = categoryOfTrait(sel.value);
    recalcAll();
  };

  window.onOriginModeChange = function (sel) {
    var c = sel.closest('.section-card');
    setWrap(c, 'origin', sel.value === 'other');
    recalcAll();
  };

  window.onInlineEffectOriginModeChange = function (sel) {
    var c = sel.closest('.section-card');
    setWrap(c, 'effect-origin', sel.value === 'other');
    recalcAll();
  };

  window.onEngageModeChange = function (sel) {
    var c = sel.closest('.section-card');
    var v = sel.value;
    setWrap(c, 'engage-trait', v === 'trait');
    setWrap(c, 'engage-generic', v === 'generic');
    setWrap(c, 'engage-other', v === 'other');
    setWrap(c, 'engage-previous', v === 'previous');
    recalcAll();
  };

  window.onEngageTraitChange = function (sel) {
    var c = sel.closest('.section-card');
    var hidden = c.querySelector('[name="engage_trait_category"]');
    if (hidden) hidden.value = categoryOfTrait(sel.value);
    recalcAll();
  };

  window.onCounterTypeChange = function (sel) {
    var row = sel.parentElement;
    var existingTraitSel = row.querySelector('[name="counter_trait"]');
    var t = sel.value;
    var newSelHTML = '';
    if (t === 'defense' || t === 'general' || t === 'offense') {
      var cat = t === 'defense' ? 'defense' : t === 'general' ? 'general' : 'offense';
      newSelHTML = '<select name="counter_trait" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
        '<option value="">-- Select --</option>'+traitOptions(cat, existingTraitSel ? existingTraitSel.value : '')+
      '</select>';
    } else {
      newSelHTML = '<input type="text" name="counter_trait" placeholder="(reference)" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">';
    }
    existingTraitSel.outerHTML = newSelHTML;
    recalcAll();
  };

  // Generic counter trait rows: filter the value dropdown by selected type.
  function updateCounterTraitValue(typeSel) {
    var row = typeSel.closest('[data-row]');
    if (!row) return;
    var valueSel = row.querySelector('select[name="counter_trait__value"]');
    if (!valueSel) return;
    var selectedValue = valueSel.value;
    var category = typeSel.value;
    valueSel.innerHTML = '<option value="">-- Select --</option>' + traitOptionsForCategory(category, selectedValue);
  }
  document.addEventListener('change', function(e) {
    if (e.target.name === 'counter_trait__type') {
      updateCounterTraitValue(e.target);
      recalcAll();
    }
  });

  window.addDirection = function (btn) {
    var list = btn.parentElement.parentElement.querySelector('[data-list="directions"]');
    var first = list.querySelector('select[name="direction"]');
    var val = first ? first.value : 'Away';
    var card = btn.closest('.section-card[data-section="enact"]');
    var cfg = card ? getEnactConfig(card.dataset.enactType) : {};
    var freeDir = perkCost(cfg.perks, 'free_direction');
    var opts = D.directionOptions.map(function(o){
      if (o === 'Free') return opt(o, o, val, freeDir.add, freeDir.energy);
      return opt(o, o, val, 0, 0);
    }).join('');
    var row = document.createElement('div');
    row.className = 'flex items-center gap-2';
    row.innerHTML =
      '<select name="direction" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
      '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>';
    list.appendChild(row);
    recalcAll();
  };

  window.addSolution = function (btn) {
    var list = btn.parentElement.parentElement.querySelector('[data-list="solutions"]');
    var first = list.querySelector('select[name="solution"]');
    var val = first ? first.value : 'Dexterity';
    var opts = '<option value="">-- Select --</option>' + traitOptionsGrouped(val);
    var row = document.createElement('div');
    row.className = 'flex items-center gap-2';
    row.innerHTML =
      '<select name="solution" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
      '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>';
    list.appendChild(row);
    recalcAll();
  };

  window.addCounter = function (btn) {
    var list = btn.parentElement.parentElement.querySelector('[data-list="counters"]');
    list.insertAdjacentHTML('beforeend', counterRow({type:'defense', trait:''}, getValidationConfig()));
    recalcAll();
  };

  function setWrap(c, key, show) {
    var el = c.querySelector('[data-wrap="'+key+'"]');
    if (el) el.hidden = !show;
  }

  // =========================================================================
  // Block management
  // =========================================================================

  function makeEmptyEnactBlock() {
    var idx = nextIndex();
    var block = document.createElement('div');
    block.className = 'enactment-block border border-indigo-700 rounded-lg p-4 bg-gray-900 space-y-3';
    block.dataset.index = idx;
    block.innerHTML = [
      '<div class="flex items-center justify-between flex-wrap gap-2">',
        '<button type="button" class="collapse-toggle flex items-center gap-2 text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<span class="collapse-chevron text-indigo-400 text-xs">&#9660;</span>',
          '<h3 class="text-sm font-semibold text-indigo-400">Enactment #'+idx+'</h3>',
        '</button>',
        '<div class="flex items-center gap-2 flex-wrap">',
          '<label class="text-xs text-gray-400 flex items-center gap-1">Name: ',
            '<input type="text" name="enactment_name" placeholder="e.g. Main Strike" class="enactment-name bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white w-40">',
          '</label>',
          '<button type="button" onclick="removeEnactment(this)" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">Remove</button>',
        '</div>',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">',
        statCard('Always Resolve', 'No', 'resolve'),
        statCard('Build Cost', '0', 'build'),
        statCard('Cast Cost', '0', 'cast'),
      '</div>',
      '<div class="collapsible-content space-y-3">',
        '<div class="section-card bg-gray-800 rounded-lg border border-gray-700 p-4 space-y-3">',
          '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="false" onclick="toggleCollapse(this)">',
            '<h4 class="text-sm font-semibold text-gray-300">Description</h4>',
            '<span class="collapse-chevron text-gray-300 text-xs">&#9654;</span>',
          '</button>',
          '<div class="collapsible-content space-y-3 mt-3" hidden>',
            '<textarea name="enactment_description" rows="3" placeholder="Describe this enactment..." class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-white text-sm"></textarea>',
          '</div>',
        '</div>',
        '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-4 space-y-3" data-section="enactment-type">',
          '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
            '<h4 class="text-sm font-semibold text-indigo-300">Enactment Type</h4>',
            '<span class="collapse-chevron text-indigo-300 text-xs">&#9660;</span>',
          '</button>',
          '<div class="collapsible-content space-y-3 mt-3">',
            '<div class="flex items-center gap-3 flex-wrap">',
              '<label class="text-xs text-gray-400 flex items-center gap-1">Enactment Type: ',
                '<select onchange="onEnactTypeChange(this)" class="enact-type-select bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white">',
                  '<option value="">-- Select --</option>',
                  ENACT_TYPES.map(function(t){return '<option value="'+esc(t)+'">'+esc(t)+'</option>';}).join(''),
                '</select>',
              '</label>',
            '</div>',
            '<div class="enact-card-container space-y-2"></div>',
          '</div>',
        '</div>',
        '<div class="inter-card-container space-y-2">',
          '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-build="0" data-cast="0">',
            '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
              '<h4 class="text-sm font-semibold text-cyan-300">Interaction</h4>',
              '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
            '</button>',
            '<div class="collapsible-content space-y-3 mt-3">',
              interTypeSelect(''),
            '</div>',
          '</div>',
        '</div>',
        '<div class="validation-card-container"></div>',
      '</div>',
    ].join('');
    return block;
  }

  window.addEnactment = function () {
    var c = document.getElementById('enactments-container');
    var block = makeEmptyEnactBlock();
    c.appendChild(block);
    // Always render the validation card at the bottom of the new block
    block.querySelector('.validation-card-container').innerHTML = renderValidationCard({});
    recalcAll();
  };

  window.removeEnactment = function (btn) {
    var block = btn.closest('.enactment-block');
    block.parentElement.removeChild(block);
    recalcAll();
  };

  // =========================================================================
  // Initial state hydration (from JSON encoded by the server for edit mode)
  // =========================================================================

  var initialState = null;
  function loadInitialState() {
    var el = document.getElementById('initial-state');
    if (!el) return null;
    var raw = el.value;
    if (!raw || raw === '') return null;
    try { return JSON.parse(raw); } catch (e) { return null; }
  }

  // =========================================================================
  // Cost calculation (purely live UI feedback; server has its own copy)
  // =========================================================================

  function readNumber(card, name, fallback) {
    var el = card && card.querySelector('[name="'+name+'"]');
    return el ? (Number(el.value) || fallback || 0) : (fallback || 0);
  }
  function readBool(card, name) {
    var el = card && card.querySelector('[name="'+name+'"]');
    return el ? el.checked : false;
  }

  function calcAbilityType() {
    var card = document.querySelector('.section-card[data-section="ability-type"]');
    if (!card) return;
    var cfg = abilityTypeConfig();

    if (getGenericFieldsForCard(card)) {
      var values = readGenericCardValues(card);
      var res = abilityTypeCalcJS(cfg, values);
      setOut(card, 'build', res.build);
      setOut(card, 'cast', res.energy);
      card.dataset.build = res.build;
      card.dataset.cast = res.energy;
      return;
    }

    var lines = [];
    var build = 0, energy = cfg.base_energy || 0;

    var item = findPerk(cfg.perks, 'item_dependency');
    if (item && readBool(card, 'item_dep')) {
      build += item.add_cost || 0;
      lines.push('Has item dependency (add '+(item.add_cost||0)+', energy '+(item.energy_cost||0)+')');
    }

    var es = readNumber(card, 'energy_steps', 0);
    var as = readNumber(card, 'action_steps', 0);

    var energySteps = stepCost(cfg.step_costs && cfg.step_costs.energy, 'increase');
    var energyStepDec = stepCost(cfg.step_costs && cfg.step_costs.energy, 'decrease');
    var actionSteps = stepCost(cfg.step_costs && cfg.step_costs.action, 'increase');
    var actionStepDec = stepCost(cfg.step_costs && cfg.step_costs.action, 'decrease');

    if (es > 0) { build += es * energySteps.add; energy += es * energySteps.energy; lines.push('Increase energy cost (add '+(es*energySteps.add)+', energy '+(es*energySteps.energy)+')'); }
    else if (es < 0) { build += Math.abs(es) * energyStepDec.add; energy += Math.abs(es) * energyStepDec.energy; lines.push('Reduce energy cost (add '+(Math.abs(es)*energyStepDec.add)+', energy '+(Math.abs(es)*energyStepDec.energy)+')'); }
    if (as > 0) { build += as * actionSteps.add; energy += as * actionSteps.energy; lines.push('Increase action cost (add '+(as*actionSteps.add)+', energy '+(as*actionSteps.energy)+')'); }
    else if (as < 0) { build += Math.abs(as) * actionStepDec.add; energy += Math.abs(as) * actionStepDec.energy; lines.push('Reduce action cost (add '+(Math.abs(as)*actionStepDec.add)+', energy '+(Math.abs(as)*actionStepDec.energy)+')'); }

    if (card.querySelector('[name="trigger"]')) {
      var range = readNumber(card, 'range', 1);
      var uses  = readNumber(card, 'uses', 1);
      var rangeCost = cfg.range_cost || {add_cost:0, energy_cost:0};
      var usesCost = cfg.uses_cost || {add_cost:0, energy_cost:0};
      if (range > 1) { build += (range - 1) * rangeCost.add_cost; energy += (range - 1) * rangeCost.energy_cost; lines.push('Add reaction range (add '+((range-1)*rangeCost.add_cost)+', energy '+((range-1)*rangeCost.energy_cost)+')'); }
      if (uses > 1)  { build += (uses-1) * usesCost.add_cost; energy += (uses-1) * usesCost.energy_cost; lines.push('Add uses (add '+((uses-1)*usesCost.add_cost)+', energy '+((uses-1)*usesCost.energy_cost)+')'); }
      var trigger = card.querySelector('[name="trigger"]').value || '';
      var triggerCost = findPerk(cfg.triggers, trigger);
      if (triggerCost) { build += triggerCost.add_cost || 0; energy += triggerCost.energy_cost || 0; lines.push('Trigger (add '+(triggerCost.add_cost||0)+', energy '+(triggerCost.energy_cost||0)+')'); }
      var triggerTrait = card.querySelector('[data-wrap="trigger-trait"]') && !card.querySelector('[data-wrap="trigger-trait"]').hidden
        ? card.querySelector('[name="trigger_trait"]').value : '';
      var typeLabel = card.dataset.abilityType === 'Preparation' ? 'Preparation' : 'Reaction';
      setOut(card, 'formula', typeLabel+', '+uses+' uses/round, range '+range+'m, trigger: '+trigger+(triggerTrait?(' of type '+triggerTrait):''));
    }
    if (card.querySelector('[name="phase_rounds"]')) {
      var phase = readNumber(card, 'phase_rounds', 2);
      var rev   = readNumber(card, 'reverse_rounds', 1);
      var durCost = cfg.duration_cost || {add_cost:0, energy_cost:0};
      var revRefund = cfg.reverse_duration_refund || {add_cost:0, energy_cost:0};
      if (phase > 2) { build += (phase-2) * durCost.add_cost; energy += (phase-2) * durCost.energy_cost; lines.push('Add phase rounds (add '+((phase-2)*durCost.add_cost)+', energy '+((phase-2)*durCost.energy_cost)+')'); }
      if (rev < phase) { build += (phase-rev) * revRefund.add_cost; energy += (phase-rev) * revRefund.energy_cost; lines.push('Remove reverse-phase rounds (add '+((phase-rev)*revRefund.add_cost)+', energy '+((phase-rev)*revRefund.energy_cost)+')'); }
      var allReq = findPerk(cfg.perks, 'all_knockouts_req');
      if (allReq && readBool(card, 'all_req'))        { build += allReq.add_cost; energy += allReq.energy_cost; lines.push('All knockout requirements met (add '+allReq.add_cost+')'); }
      var revKo = findPerk(cfg.perks, 'reverse_knockout');
      if (revKo && readBool(card, 'reverse_knockout')){ build += revKo.add_cost; energy += revKo.energy_cost; lines.push('Knockout on reverse phase (add '+revKo.add_cost+')'); }
      var noKo = findPerk(cfg.perks, 'no_knockout');
      if (noKo && readBool(card, 'no_knockout'))    { build += noKo.add_cost; energy += noKo.energy_cost; lines.push('No knockout possible (add '+noKo.add_cost+')'); }
      var koReqs = cfg.knockout_requirements || [];
      var kos = Array.from(card.querySelectorAll('[name="knockout"]')).map(function(s){return s.value;}).filter(function(v){return v && v !== 'None';});
      kos.forEach(function(k){
        var koCost = findPerk(koReqs, k);
        if (koCost && (koCost.add_cost || koCost.energy_cost)) {
          build += koCost.add_cost; energy += koCost.energy_cost;
          lines.push('Knockout '+k+' (add '+koCost.add_cost+', energy '+koCost.energy_cost+')');
        }
      });
      setOut(card, 'formula', 'Phase for '+phase+' rounds, reverse '+rev+', knockouts: '+(kos.length?kos.join(' or '):'(none)'));
    }
    if (card.querySelector('[name="hp"]')) {
      energy = cfg.base_energy || 0;
      var hp   = readNumber(card, 'hp', 0);
      var life = readNumber(card, 'life', 0);
      var hpCost = cfg.health_bonus_cost || {add_cost:0, energy_cost:0};
      var lifeCost = cfg.lifetime_bonus_cost || {add_cost:0, energy_cost:0};
      if (hp > 0)   { build += hp * hpCost.add_cost; energy += hp * hpCost.energy_cost; lines.push('Increase health by '+(hp*5)+' (add '+(hp*hpCost.add_cost)+')'); }
      if (life > 0) { build += life * lifeCost.add_cost; energy += life * lifeCost.energy_cost; lines.push('Increase lifetime by '+life+' rounds (add '+(life*lifeCost.add_cost)+', energy '+(life*lifeCost.energy_cost)+')'); }
      setOut(card, 'formula', 'Minion: '+(10+hp*5)+' HP, '+(3+life)+' round lifetime');
    }
    if (card.querySelector('[name="effortless"]')) {
      var eff = findPerk(cfg.perks, 'effortless');
      if (eff && readBool(card, 'effortless')) { build += eff.add_cost || 0; energy += eff.energy_cost || 0; lines.push('Effortless (add '+(eff.add_cost||0)+', energy '+(eff.energy_cost||0)+')'); }
      var iw = findPerk(cfg.perks, 'iron_will');
      if (iw && readBool(card, 'iron_will')) { build += iw.add_cost || 0; energy += iw.energy_cost || 0; lines.push('Iron Will (add '+(iw.add_cost||0)+', energy '+(iw.energy_cost||0)+')'); }
      var df = findPerk(cfg.perks, 'dual_focus');
      if (df && readBool(card, 'dual_focus')) { build += df.add_cost || 0; energy += df.energy_cost || 0; lines.push('Dual Focus (add '+(df.add_cost||0)+', energy '+(df.energy_cost||0)+')'); }
      var upkeep = (cfg.base_upkeep_action||1) + ' Action or ' + (cfg.base_upkeep_energy||1) + ' Energy per turn';
      setOut(card, 'formula', 'Concentration, upkeep: '+upkeep);
    }

    setOut(card, 'resolve', 'No');
    setOut(card, 'build', build);
    setOut(card, 'cast', energy);
    card.dataset.build = build;
    card.dataset.cast = energy;
    fillList(card, lines);
  }

  function setOut(card, key, value) {
    var el = card && card.querySelector('[data-out="'+key+'"]');
    if (el) el.textContent = value;
    var hidden = card && card.querySelector('input[type="hidden"][name="'+key+'"]');
    if (hidden) hidden.value = value;
  }
  function upsertHidden(container, name, value) {
    if (!container) return;
    var el = container.querySelector('input[type="hidden"][name="'+name+'"]');
    if (!el) {
      el = document.createElement('input');
      el.type = 'hidden';
      el.name = name;
      container.appendChild(el);
    }
    el.value = value === undefined || value === null ? '' : value;
  }
  function prepareCardSubmit(container, typeAttr) {
    if (!container) return;
    var card = container.querySelector('.section-card');
    if (!card) return;
    if (typeAttr) upsertHidden(container, 'type', card.dataset[typeAttr] || '');
    upsertHidden(container, 'build', card.dataset.build || '0');
    upsertHidden(container, 'cast', card.dataset.cast || '0');
    var formula = card.querySelector('[data-out="formula"]');
    if (formula) upsertHidden(container, 'formula', formula.textContent || '');
  }
  function fillList(card, lines) {
    var host = card && card.querySelector('[data-out="breakdown"]');
    if (!host) return;
    host.innerHTML = '';
    lines.forEach(function(l){
      var li = document.createElement('li');
      li.textContent = l;
      host.appendChild(li);
    });
  }

  // Computes the source/flat/offense/medicine/distance/shift cost contribution
  // for a set of fields on a card. `prefix` is either '' (standalone
  // enactment) or 'effect_' (persistent-effect inline editor). The returned
  // object mutates `acc` and pushes human-readable lines into `lines`.
  function calcEnactFieldCosts(card, cfg, prefix, acc, lines) {
    prefix = prefix || '';
    var build = 0, energy = 0;

    var srcEl = card.querySelector('[name="'+prefix+'source"]');
    if (srcEl) {
      var s = srcEl.value;
      var tiers = cfg.dice_tiers || {d4:0,d6:1,d8:2,d10:3,d12:4};
      var tierCost = cfg.dice_tier_cost || {add_cost:0, energy_cost:0};
      if (tiers[s] !== undefined) {
        var tier = tiers[s] || 0;
        build += tier * tierCost.add_cost;
        energy += tier * tierCost.energy_cost;
        lines.push('Source 1'+s+' (add '+(tier*tierCost.add_cost)+', energy '+(tier*tierCost.energy_cost)+')');
      } else if (s === 'trait') {
        var cat = (card.querySelector('[name="'+prefix+'source_category"]') || {}).value || 'offense';
        var traitCost = findPerk(cfg.perks, 'trait_source');
        var traitAdd = traitCost ? traitCost.add_cost : 0;
        if (cat === 'general') {
          build += traitAdd + 1; // legacy general trait extra cost
          lines.push('Use general trait as source (add '+(traitAdd+1)+', extra cost)');
        } else {
          build += traitAdd;
          lines.push('Use trait as source (add '+traitAdd+')');
        }
      } else if (s === 'previous') {
        var prevCost = findPerk(cfg.perks, 'use_previous');
        build += prevCost ? prevCost.add_cost : 0;
        energy += prevCost ? prevCost.energy_cost : 0;
        lines.push('Use result of previous enactment (add '+(prevCost?prevCost.add_cost:0)+', energy '+(prevCost?prevCost.energy_cost:0)+')');
      } else if (s === 'other') {
        var otherCost = findPerk(cfg.perks, 'use_previous');
        build += otherCost ? otherCost.add_cost : 0;
        energy += otherCost ? otherCost.energy_cost : 0;
        lines.push('Use another roll result (add '+(otherCost?otherCost.add_cost:0)+', energy '+(otherCost?otherCost.energy_cost:0)+')');
      }
    }

    var flatEl = card.querySelector('[name="'+prefix+'flat"]');
    if (flatEl) {
      var flat = Number(flatEl.value) || 0;
      var flatCost = findPerk(cfg.perks, 'flat_bonus');
      var flatAdd = flatCost ? flatCost.add_cost : 0;
      var flatEnergy = flatCost ? flatCost.energy_cost : 0;
      if (flat > 0) { build += flat * flatAdd; energy += flat * flatEnergy; lines.push('Flat +'+flat+' (add '+(flat*flatAdd)+', energy '+(flat*flatEnergy)+')'); }
    }

    var offenseEl = card.querySelector('[name="'+prefix+'offense"]');
    if (offenseEl && offenseEl.value) {
      var off = findPerk(cfg.perks, 'offensive_trait');
      build += off ? off.add_cost : 0;
      energy += off ? off.energy_cost : 0;
      lines.push('Offensive trait die ('+offenseEl.value+') (add '+(off?off.add_cost:0)+', energy '+(off?off.energy_cost:0)+')');
    }
    var medEl = card.querySelector('[name="'+prefix+'medicine"]');
    if (medEl && medEl.value) {
      var med = findPerk(cfg.perks, 'medicine_trait');
      build += med ? med.add_cost : 0;
      energy += med ? med.energy_cost : 0;
      lines.push('Medicine trait (add '+(med?med.add_cost:0)+', energy '+(med?med.energy_cost:0)+')');
    }

    var distanceEl = card.querySelector('[name="'+prefix+'distance"]');
    if (distanceEl) {
      var dist = Number(distanceEl.value) || 1;
      var dirArr = Array.from(card.querySelectorAll('[name="'+prefix+'direction"]')).map(function(x){return x.value;}).filter(Boolean);
      var originMode = (card.querySelector('[name="'+prefix+'origin_mode"]')||{}).value;
      var distCost = cfg.distance_cost || {add_cost:0, energy_cost:0};
      if (dist > 1) { build += (dist-1) * distCost.add_cost; energy += (dist-1) * distCost.energy_cost; lines.push('Distance +'+(dist-1)+'m (add '+((dist-1)*distCost.add_cost)+', energy '+((dist-1)*distCost.energy_cost)+')'); }
      var originCost = findPerk(cfg.perks, 'other_origin');
      if (originMode === 'other') { build += originCost ? originCost.add_cost : 0; energy += originCost ? originCost.energy_cost : 0; lines.push('Other origin (add '+(originCost?originCost.add_cost:0)+', energy '+(originCost?originCost.energy_cost:0)+')'); }
      var extraDir = findPerk(cfg.perks, 'extra_direction');
      if (dirArr.length > 1) { build += (dirArr.length-1) * (extraDir?extraDir.add_cost:0); energy += (dirArr.length-1) * (extraDir?extraDir.energy_cost:0); lines.push('Extra direction '+(dirArr.length-1)+' (add '+((dirArr.length-1)*(extraDir?extraDir.add_cost:0))+', energy '+((dirArr.length-1)*(extraDir?extraDir.energy_cost:0))+')'); }
      var freeDir = findPerk(cfg.perks, 'free_direction');
      var freeCount = dirArr.filter(function(d){return d==='Free';}).length;
      if (freeCount > 0) { build += freeCount * (freeDir?freeDir.add_cost:0); energy += freeCount * (freeDir?freeDir.energy_cost:0); lines.push('Free direction '+freeCount+'x (add '+(freeCount*(freeDir?freeDir.add_cost:0))+', energy '+(freeCount*(freeDir?freeDir.energy_cost:0))+')'); }
    }

    var shiftTraitEl = card.querySelector('[name="'+prefix+'shifted_trait"]');
    if (shiftTraitEl) {
      var amt   = Number((card.querySelector('[name="'+prefix+'shift_amount"]')||{}).value) || 1;
      var uses  = Number((card.querySelector('[name="'+prefix+'shift_uses"]')||{}).value) || 1;
      var amtCost = cfg.shift_amount_cost || {add_cost:0, energy_cost:0};
      var usesCost = cfg.shift_uses_cost || {add_cost:0, energy_cost:0};
      if (amt > 1) { build += (amt-1) * amtCost.add_cost; energy += (amt-1) * amtCost.energy_cost; lines.push('Shift amount +'+(amt-1)+' (add '+((amt-1)*amtCost.add_cost)+', energy '+((amt-1)*amtCost.energy_cost)+')'); }
      if (uses > 1) { build += (uses-1) * usesCost.add_cost; energy += (uses-1) * usesCost.energy_cost; lines.push('Shift uses +'+(uses-1)+' (add '+((uses-1)*usesCost.add_cost)+', energy '+((uses-1)*usesCost.energy_cost)+')'); }
    }

    acc.build += build;
    acc.energy += energy;
  }

  function calcEnact(card) {
    if (!card) return;
    var cfg = getEnactConfig(card.dataset.enactType);

    if (getGenericFieldsForCard(card)) {
      var values = readGenericCardValues(card);
      if (card.dataset.enactType === 'Enact State' && values.states) {
        var surcharge = (C.states && C.states.additional_state) || null;
        var gens = (C.states && C.states.general_states) || [];
        var specs = (C.states && C.states.specific_states) || [];
        var stateField = null;
        for (var sfi = 0; sfi < cfg.fields.length; sfi++) {
          if (cfg.fields[sfi].type === 'states') { stateField = cfg.fields[sfi]; break; }
        }
        var rowFields = stateField ? stateField.row_fields : [];
        var build = 0, energy = 0;
        for (var ri = 0; ri < values.states.length; ri++) {
          var r = values.states[ri] || {};
          var rowCosts = evalFieldsJS(rowFields, r);
          build += rowCosts.build; energy += rowCosts.cast;
          var stateCost = evalStateRowJS(r, gens, specs, surcharge);
          build += stateCost.build; energy += stateCost.cast;
        }
        var sur = evalStatesSurchargeJS(surcharge, values.states.length);
        build += sur.build; energy += sur.cast;
        card.dataset.build = build;
        card.dataset.cast = energy;
        return;
      }
      var res = genericCalcJS(cfg, values);
      var totalBuild = res.build;
      var totalCast = res.cast;
      // Enact Persistent Effect: if an inline effect has been chosen, add
      // its per-field costs (source dice, flat, offense, medicine,
      // distance, shift) using the matching enactment's config. The
      // base cost of the inline effect is already covered by the
      // effect_type dropdown's cost in the generic fields.
      if (card.dataset.enactType === 'Enact Persistent Effect' && PERSISTENT_TO_ENACT[values.effect_type || '']) {
        var inlineCfg = getEnactConfig(values.effect_type);
        var inlineAcc = { build: 0, energy: 0 };
        var inlineLines = [];
        calcEnactFieldCosts(card, inlineCfg, 'effect_', inlineAcc, inlineLines);
        totalBuild += inlineAcc.build;
        totalCast += inlineAcc.energy;
      }
      card.dataset.build = totalBuild;
      card.dataset.cast = totalCast;
      return;
    }

    var lines = [];
    var build = 0, energy = cfg.base_cost ? (cfg.base_cost.energy_cost || 0) : 0;
    var formula = '';

    var always = findPerk(cfg.perks, 'always_resolve');
    if (always && readBool(card, 'always')) {
      build += always.add_cost || 0;
      energy += always.energy_cost || 0;
      lines.push('Always resolve (add '+(always.add_cost||0)+', energy '+(always.energy_cost||0)+')');
    }

    var acc = { build: 0, energy: 0 };
    calcEnactFieldCosts(card, cfg, '', acc, lines);
    build += acc.build;
    energy += acc.energy;

    var srcEl = card.querySelector('[name="source"]');
    if (srcEl) {
      var s = srcEl.value;
      var tiers = cfg.dice_tiers || {d4:0,d6:1,d8:2,d10:3,d12:4};
      if (tiers[s] !== undefined) {
        formula = '1'+s;
      } else if (s === 'trait') {
        var cat = (card.querySelector('[name="source_category"]') || {}).value || 'offense';
        formula = '1d10 ('+cat+' trait)';
      } else if (s === 'previous') {
        formula = 'previous enactment result';
      } else if (s === 'other') {
        var txt = (card.querySelector('[name="other"]')||{}).value || '';
        formula = txt || '(other roll result)';
      }
      var flatEl = card.querySelector('[name="flat"]');
      if (flatEl) {
        var flat = Number(flatEl.value) || 0;
        if (flat > 0) formula += ' + '+flat;
      }
      var offenseEl = card.querySelector('[name="offense"]');
      if (offenseEl && offenseEl.value) formula += ' + 1d8 ('+offenseEl.value+')';
      var medEl = card.querySelector('[name="medicine"]');
      if (medEl && medEl.value) formula += ' + 1d10 (Medicine)';
    }
    var distanceEl = card.querySelector('[name="distance"]');
    if (distanceEl) {
      var dist = Number(distanceEl.value) || 1;
      var dirArr = Array.from(card.querySelectorAll('[name="direction"]')).map(function(x){return x.value;}).filter(Boolean);
      var originMode = (card.querySelector('[name="origin_mode"]')||{}).value;
      var origin = originMode === 'other' ? (card.querySelector('[name="origin_text"]')||{}).value || '(other)' : 'Engager';
      formula = 'Move target '+dist+'m '+(dirArr.join(' or '))+' from '+origin;
    }
    var shiftTraitEl = card.querySelector('[name="shifted_trait"]');
    if (shiftTraitEl) {
      var trait = shiftTraitEl.value || '(trait)';
      var dir   = (card.querySelector('[name="shift_dir"]')||{}).value || 'UP';
      var amt   = Number((card.querySelector('[name="shift_amount"]')||{}).value) || 1;
      var uses  = Number((card.querySelector('[name="shift_uses"]')||{}).value) || 1;
      formula = 'Shift '+trait+' '+dir+' by '+amt+' for '+uses+' uses';
    }

    var effName = card.querySelector('[name="effect_name"]');
    if (effName) {
      // Persistent effect: base energy from the persistent config's base_cost,
      // plus duration step, plus the chosen effect type's flat cost, plus the
      // inline editor's per-field costs (read with the 'effect_' prefix from
      // the matching enactment config), plus always on the persistent card.
      build = 0;
      energy = cfg.base_cost ? (cfg.base_cost.energy_cost || 0) : 0;
      if (always && readBool(card, 'always')) {
        build += always.add_cost || 0;
        energy += always.energy_cost || 0;
        lines.push('Always resolve (add '+(always.add_cost||0)+', energy '+(always.energy_cost||0)+')');
      }
      var dur = readNumber(card, 'duration', 2);
      var durCost = cfg.duration_cost || {add_cost:0, energy_cost:0};
      if (dur > 2) { build += (dur-2) * durCost.add_cost; energy += (dur-2) * durCost.energy_cost; lines.push('Duration '+dur+' rounds (add '+((dur-2)*durCost.add_cost)+', energy '+((dur-2)*durCost.energy_cost)+')'); }
      var sols = Array.from(card.querySelectorAll('[name="solution"]')).map(function(s){return s.value;}).filter(Boolean);
      var singleSol = findPerk(cfg.perks, 'single_solution');
      var solDiff = 2 - sols.length;
      if (singleSol && solDiff !== 0) {
        build += solDiff * (singleSol.add_cost || 0);
        energy += solDiff * (singleSol.energy_cost || 0);
        var solDir = solDiff > 0 ? 'Remove' : 'Add';
        lines.push(solDir+' '+Math.abs(solDiff)+' solution option'+(Math.abs(solDiff)===1?'':'s')+' (add '+(solDiff*(singleSol.add_cost||0))+', energy '+(solDiff*(singleSol.energy_cost||0))+')');
      }
      var effType = (card.querySelector('[name="effect_type"]')||{}).value || '(effect)';
      var effCfg = findPerk(cfg.effects, effType) || (function(){
        var e = cfg.effects || []; var match = null;
        for (var i = 0; i < e.length; i++) { if (e[i].description === effType) { match = e[i]; break; } }
        return match;
      })();
      if (effCfg) {
        build += effCfg.add_cost || 0;
        energy += effCfg.energy_cost || 0;
        lines.push('Effect '+effType+' (add '+(effCfg.add_cost||0)+', energy '+(effCfg.energy_cost||0)+')');
      }
      // If an inline effect is configured, add its per-field costs using the
      // matching enactment config. The base cost for the inline effect is
      // already covered by the effCfg surcharge above; this only contributes
      // the dice/flat/offense/medicine/distance/shift-style per-field costs.
      if (effType && PERSISTENT_TO_ENACT[effType]) {
        var inlineCfg = getEnactConfig(effType);
        var inlineAcc = { build: 0, energy: 0 };
        var inlineLines = [];
        calcEnactFieldCosts(card, inlineCfg, 'effect_', inlineAcc, inlineLines);
        // Inline always is counted inside the field-cost helper via the
        // `effect_always` checkbox; add it explicitly to be safe.
        if (always && (card.querySelector('[name="effect_always"]') || {}).checked) {
          inlineAcc.build += always.add_cost || 0;
          inlineAcc.energy += always.energy_cost || 0;
        }
        build += inlineAcc.build;
        energy += inlineAcc.energy;
        for (var ili = 0; ili < inlineLines.length; ili++) {
          lines.push('  '+inlineLines[ili]);
        }
      }
      formula = (effName.value||'Effect')+' applies '+effType+' for '+dur+' rounds, solutions: '+(sols.join(' or ')||'(none)');
    }

    // Enact Negation-specific perks (counter_negation / full_counter)
    if (card.dataset.enactType === 'Enact Negation') {
      var cneg = findPerk(cfg.perks, 'counter_negation');
      if (cneg && readBool(card, 'counter_negation')) {
        build += cneg.add_cost || 0;
        energy += cneg.energy_cost || 0;
        lines.push('Negation applied to counter roll (add '+(cneg.add_cost||0)+', energy '+(cneg.energy_cost||0)+')');
      }
      var fc = findPerk(cfg.perks, 'full_counter');
      if (fc && readBool(card, 'full_counter')) {
        build += fc.add_cost || 0;
        energy += fc.energy_cost || 0;
        lines.push('Ability hits Engager instead (add '+(fc.add_cost||0)+', energy '+(fc.energy_cost||0)+')');
      }
    }

    // Enact State (WIP placeholder — no costs yet, just a friendly formula).
    if (card.dataset.enactType === 'Enact State') {
      var stateName = (card.querySelector('[name="effect_name"]') || {}).value || 'State';
      build = 0;
      energy = 0;
      lines = [];
      formula = stateName + ' (WIP — no cost yet)';
    }

    card.dataset.build = build;
    card.dataset.cast = energy;
  }

  function calcInter(card) {
    if (!card) return;
    var cfg = getInterConfig(card.dataset.interType);
    if (getGenericFieldsForCard(card)) {
      var values = readGenericCardValues(card);
      var res = genericCalcJS(cfg, values);
      card.dataset.build = res.build;
      card.dataset.cast = res.cast;
      return;
    }
    var lines = [];
    var build = 0, energy = 0;
    var formula = card.dataset.interType;

    var type = card.dataset.interType;
    if (type === 'Self') {
      formula = 'Self + Target = Self + Counter = d8';
    } else if (type === 'Direct') {
      var r = readNumber(card, 'range', 1);
      var t = readNumber(card, 'targets', 1);
      var rangeCost = cfg.range_cost || {add_cost:0, energy_cost:0};
      var targetCost = cfg.target_cost || {add_cost:0, energy_cost:0};
      if (r > 1) { build += (r-1) * rangeCost.add_cost; energy += (r-1) * rangeCost.energy_cost; lines.push('Range +'+(r-1)+' (add '+((r-1)*rangeCost.add_cost)+', energy '+((r-1)*rangeCost.energy_cost)+')'); }
      if (t > 1) { build += (t-1) * targetCost.add_cost; energy += (t-1) * targetCost.energy_cost; lines.push('Targets +'+(t-1)+' (add '+((t-1)*targetCost.add_cost)+', energy '+((t-1)*targetCost.energy_cost)+')'); }
      formula = 'Direct, '+t+' target(s), range '+r+'m';
    } else if (type === 'Ranged') {
      var r2 = readNumber(card, 'range', 10);
      var t2 = readNumber(card, 'targets', 1);
      var extCost = cfg.range_extension_cost || {add_cost:0, energy_cost:0, step:2};
      var targetCost = cfg.target_cost || {add_cost:0, energy_cost:0};
      if (r2 > 10) { var inc = Math.floor((r2-10)/(extCost.step||2)); build += inc * extCost.add_cost; energy += inc * extCost.energy_cost; lines.push('Ranged range extension +'+inc+' (add '+(inc*extCost.add_cost)+', energy '+(inc*extCost.energy_cost)+')'); }
      if (t2 > 1) { build += (t2-1) * targetCost.add_cost; energy += (t2-1) * targetCost.energy_cost; lines.push('Targets +'+(t2-1)+' (add '+((t2-1)*targetCost.add_cost)+', energy '+((t2-1)*targetCost.energy_cost)+')'); }
      var notVisible = findPerk(cfg.perks, 'not_visible');
      if (notVisible && readBool(card, 'visible'))       { build += notVisible.add_cost; energy += notVisible.energy_cost; lines.push('Target may be invisible (add '+notVisible.add_cost+', energy '+notVisible.energy_cost+')'); }
      var obstructed = findPerk(cfg.perks, 'obstructed');
      if (obstructed && readBool(card, 'obstructed'))    { build += obstructed.add_cost; energy += obstructed.energy_cost; lines.push('Target may be obstructed (add '+obstructed.add_cost+', energy '+obstructed.energy_cost+')'); }
      var removePenalty = findPerk(cfg.perks, 'remove_penalty');
      if (removePenalty && readBool(card, 'remove_penalty')){ build += removePenalty.add_cost; energy += removePenalty.energy_cost; lines.push('Remove engagement penalty (add '+removePenalty.add_cost+', energy '+removePenalty.energy_cost+')'); }
      formula = 'Ranged, '+t2+' target(s), range '+r2+'m';
    } else if (type === 'Area') {
      var radius = readNumber(card, 'radius', 1);
      var range  = readNumber(card, 'range', 0);
      var radiusCost = cfg.radius_cost || {add_cost:0, energy_cost:0};
      var rangeCost = cfg.range_cost || {add_cost:0, energy_cost:0, step:2};
      if (radius > 1) { build += (radius-1) * radiusCost.add_cost; energy += (radius-1) * radiusCost.energy_cost; lines.push('Radius +'+(radius-1)+'m (add '+((radius-1)*radiusCost.add_cost)+', energy '+((radius-1)*radiusCost.energy_cost)+')'); }
      if (range > 0)  { var rng = Math.ceil(range/(rangeCost.step||2)); build += rng * rangeCost.add_cost; energy += rng * rangeCost.energy_cost; lines.push('Range +'+range+'m (add '+(rng*rangeCost.add_cost)+', energy '+(rng*rangeCost.energy_cost)+')'); }
      var om = (card.querySelector('[name="origin_mode"]')||{}).value;
      var orig = om === 'other' ? (card.querySelector('[name="origin_text"]')||{}).value || '(origin)' : 'Engager';
      var otherOrigin = findPerk(cfg.perks, 'other_origin');
      if (om === 'other') { build += otherOrigin ? otherOrigin.add_cost : 0; energy += otherOrigin ? otherOrigin.energy_cost : 0; lines.push('Other origin (add '+(otherOrigin?otherOrigin.add_cost:0)+', energy '+(otherOrigin?otherOrigin.energy_cost:0)+')'); }
      formula = 'Area, radius '+radius+'m, range '+range+'m, origin '+orig;
    } else if (type === 'Area of Effect') {
      var radius2 = readNumber(card, 'radius', 1);
      var range2  = readNumber(card, 'range', 0);
      var dur2    = readNumber(card, 'duration', 2);
      var radiusCost = cfg.radius_cost || {add_cost:0, energy_cost:0};
      var rangeCost = cfg.range_cost || {add_cost:0, energy_cost:0, step:2};
      var durationCost = cfg.duration_cost || {add_cost:0, energy_cost:0};
      if (radius2 > 1) { build += (radius2-1) * radiusCost.add_cost; energy += (radius2-1) * radiusCost.energy_cost; lines.push('Radius +'+(radius2-1)+' (add '+((radius2-1)*radiusCost.add_cost)+', energy '+((radius2-1)*radiusCost.energy_cost)+')'); }
      if (range2 > 0)  { var rng2 = Math.ceil(range2/(rangeCost.step||2)); build += rng2 * rangeCost.add_cost; energy += rng2 * rangeCost.energy_cost; lines.push('Range +'+range2+' (add '+(rng2*rangeCost.add_cost)+', energy '+(rng2*rangeCost.energy_cost)+')'); }
      if (dur2 > 2)    { build += (dur2-2) * durationCost.add_cost; energy += (dur2-2) * durationCost.energy_cost; lines.push('Duration +'+(dur2-2)+' (add '+((dur2-2)*durationCost.add_cost)+', energy '+((dur2-2)*durationCost.energy_cost)+')'); }
      var immune = findPerk(cfg.perks, 'immune');
      if (immune && readBool(card, 'immune')) { build += immune.add_cost; energy += immune.energy_cost; lines.push('Engager immune (add '+immune.add_cost+', energy '+immune.energy_cost+')'); }
      var om2 = (card.querySelector('[name="origin_mode"]')||{}).value;
      var orig2 = om2 === 'other' ? (card.querySelector('[name="origin_text"]')||{}).value || '(origin)' : 'Engager';
      var otherOrigin = findPerk(cfg.perks, 'other_origin');
      if (om2 === 'other') { build += otherOrigin ? otherOrigin.add_cost : 0; energy += otherOrigin ? otherOrigin.energy_cost : 0; lines.push('Other origin (add '+(otherOrigin?otherOrigin.add_cost:0)+', energy '+(otherOrigin?otherOrigin.energy_cost:0)+')'); }
      formula = 'AoE, radius '+radius2+'m, range '+range2+'m, duration '+dur2+' rounds, origin '+orig2;
    }
    var usePrev = findPerk(cfg.perks, 'use_previous');
    if (usePrev && readBool(card, 'use_previous')) { build += usePrev.add_cost; energy += usePrev.energy_cost; lines.push('Use result of previous (add '+usePrev.add_cost+', energy '+usePrev.energy_cost+')'); }

    card.dataset.build = build;
    card.dataset.cast = energy;
  }

  function calcValidation(card) {
    if (!card) return;
    var cfg = getValidationConfig();
    if (getGenericFieldsForCard(card)) {
      var values = readGenericCardValues(card);
      var res = genericCalcJS(cfg, values);
      card.dataset.build = res.build;
      card.dataset.cast = res.cast;
      return;
    }
    var lines = [];
    var build = 0, energy = 0;
    var formula = '';

    var mode = (card.querySelector('[name="engage_mode"]') || {}).value || 'trait';
    if (mode === 'trait') {
      var cat = (card.querySelector('[name="engage_trait_category"]') || {}).value || 'offense';
      var t = (card.querySelector('[name="engage_trait"]') || {}).value || '(trait)';
      lines.push('Engage roll: '+cat+' trait '+t);
      formula = t + ' vs counters';
    } else if (mode === 'generic') {
      var die = (card.querySelector('[name="engage_die"]') || {}).value || 'd6';
      var generic = findPerk(cfg.engagement.modes, 'generic');
      var genericAdd = generic ? generic.add_cost : 0;
      var genericEnergy = generic ? generic.energy_cost : 0;
      build += genericAdd;
      energy += genericEnergy;
      // Die tier cost (relative to d6) via the engage_up tier shift.
      var dieTiers = { d6: 0, d8: 1, d10: 2, d12: 3 };
      var tier = dieTiers[die] || 0;
      var engageUp = findPerk(cfg.counter.tier_shifts, 'engage_up');
      var upAdd = (tier > 0 && engageUp) ? tier * engageUp.add_cost : 0;
      var upEnergy = (tier > 0 && engageUp) ? tier * engageUp.energy_cost : 0;
      build += upAdd;
      energy += upEnergy;
      lines.push('Engage roll: generic '+die+' (add '+(genericAdd + upAdd)+', energy '+(genericEnergy + upEnergy)+')');
      formula = die + ' vs counters';
    } else if (mode === 'other') {
      var txt = (card.querySelector('[name="engage_other"]') || {}).value || '(other)';
      var other = findPerk(cfg.engagement.modes, 'other');
      build += other ? other.add_cost : 0;
      energy += other ? other.energy_cost : 0;
      lines.push('Engage roll: another roll result '+txt+' (add '+(other?other.add_cost:0)+', energy '+(other?other.energy_cost:0)+')');
      formula = '(other) vs counters';
    } else if (mode === 'previous') {
      var previous = findPerk(cfg.engagement.modes, 'previous');
      build += previous ? previous.add_cost : 0;
      energy += previous ? previous.energy_cost : 0;
      lines.push('Engage roll: use previous result (add '+(previous?previous.add_cost:0)+', energy '+(previous?previous.energy_cost:0)+')');
      formula = 'previous vs counters';
    }

    // Counter rolls
    var rows = card.querySelectorAll('[data-list="counters"] > div');
    var counterCount = rows.length;
    var counters = [];
    rows.forEach(function(row){
      var type = (row.querySelector('[name="counter_type"]') || {}).value || 'defense';
      var trait = '';
      var traitField = row.querySelector('[name="counter_trait"]');
      if (traitField) trait = traitField.value || '';
      counters.push({type:type, trait:trait});
      var counterCfg = findPerk(cfg.counter.types, type);
      if (type === 'defense') {
        // default, no extra cost
      } else if (counterCfg) {
        build += counterCfg.add_cost; energy += counterCfg.energy_cost;
        lines.push(type+' counter ('+trait+') (add '+counterCfg.add_cost+', energy '+counterCfg.energy_cost+')');
      }
    });
    var singleCounter = cfg.counter.single_counter_cost || {add_cost:0, energy_cost:0};
    var counterDiff = 2 - counterCount;
    if (counterDiff !== 0) {
      build += counterDiff * (singleCounter.add_cost || 0);
      energy += counterDiff * (singleCounter.energy_cost || 0);
      var counterDir = counterDiff > 0 ? 'Remove' : 'Add';
      lines.push(counterDir+' '+Math.abs(counterDiff)+' counter option'+(Math.abs(counterDiff)===1?'':'s')+' (add '+(counterDiff*(singleCounter.add_cost||0))+', energy '+(counterDiff*(singleCounter.energy_cost||0))+')');
    }

    formula = (formula || 'engage vs counters') + ' vs ' + (counters.map(function(c){return c.trait;}).join(' or ') || '(no counters)');

    card.dataset.build = build;
    card.dataset.cast = energy;
  }

  function updateBlockTotals(block) {
    var build = 0, energy = 0;
    var always = false;
    block.querySelectorAll('.section-card[data-section="enact"]').forEach(function(card){
      build += Number(card.dataset.build || 0);
      energy += Number(card.dataset.cast || 0);
      if (readBool(card, 'always')) always = true;
    });
    block.querySelectorAll('.section-card[data-section="interaction"]').forEach(function(card){
      build += Number(card.dataset.build || 0);
      energy += Number(card.dataset.cast || 0);
    });
    block.querySelectorAll('.section-card[data-section="validation"]').forEach(function(card){
      build += Number(card.dataset.build || 0);
      energy += Number(card.dataset.cast || 0);
    });
    setOut(block, 'resolve', always ? 'Yes' : 'No');
    setOut(block, 'build', build);
    setOut(block, 'cast', energy);
  }

  window.recalcAll = function () {
    var ability = document.querySelector('.section-card[data-section="ability-type"]');
    var acts = document.querySelectorAll('.section-card[data-section="enact"]');
    var inters = document.querySelectorAll('.section-card[data-section="interaction"]');
    var valids = document.querySelectorAll('.section-card[data-section="validation"]');

    if (ability) calcAbilityType();
    acts.forEach(calcEnact);
    inters.forEach(calcInter);
    valids.forEach(calcValidation);

    document.querySelectorAll('.enactment-block').forEach(updateBlockTotals);

    var total = 0;
    var totalCast = 0;
    document.querySelectorAll('.section-card').forEach(function(c){
      total += Number(c.dataset.build || 0);
      totalCast += Number(c.dataset.cast || 0);
    });
    var extraCost = C.additional_enactment || {add_cost:1, energy_cost:0};
    total += Math.max(0, acts.length - 1) * (extraCost.add_cost || 1); // +N per additional enactment
    totalCast += Math.max(0, acts.length - 1) * (extraCost.energy_cost || 0);
    var totalEl = document.getElementById('total-cost');
    if (totalEl) totalEl.textContent = total;
    var totalCastEl = document.getElementById('total-cast-cost');
    if (totalCastEl) totalCastEl.textContent = totalCast;
    var spent = Number(D.abilityPointsUsed || 0) + total;
    var budget = Number(D.abilityPointsBudget || 0);
    var remaining = budget - spent;
    var usedEl = document.getElementById('ability-points-used');
    if (usedEl) usedEl.textContent = remaining;
    var budgetEl = document.getElementById('ability-points-budget');
    if (budgetEl) budgetEl.textContent = budget;
    var displayEl = document.getElementById('ability-points-display');
    if (displayEl) displayEl.className = remaining < 0 ? 'text-2xl font-bold text-red-400' : 'text-2xl font-bold text-green-400';
    var warningEl = document.getElementById('ability-points-warning');
    if (warningEl) warningEl.style.display = remaining < 0 ? 'block' : 'none';
  };

  // =========================================================================
  // Form submit: rename fields with enactment indices.
  //
  // Strategy: walk every .enactment-block, take every input, and prefix its
  // name with "enact_<idx>_". The hidden ability-type select at the top has
  // its own name; we copy that into the existing field. Same for the
  // validation card fields.
  // =========================================================================

  document.addEventListener('submit', function (evt) {
    var form = evt.target;
    if (!form || form.id !== 'ability-form') return;

    // Copy visible ability-type select into the always-submitted hidden field
    var typeSel = document.getElementById('ability-type-select');
    var hiddenType = document.getElementById('hidden-ability-type');
    if (typeSel && hiddenType) hiddenType.value = typeSel.value;

    if (!document.querySelector('.section-card[data-section="interaction"][data-inter-type]')) {
      evt.preventDefault();
      alert('Add at least one interaction before saving the ability.');
      return;
    }

    var blocks = document.querySelectorAll('.enactment-block');
    blocks.forEach(function (block, idx) {
      block.dataset.index = idx; // re-number in display order
      prepareCardSubmit(block.querySelector('.enact-card-container'), 'enactType');
      prepareCardSubmit(block.querySelector('.inter-card-container'), 'interType');
      prepareCardSubmit(block.querySelector('.validation-card-container'));
      // Rename inputs inside the enact card only (rendered inside
      // .enact-card-container); inter card (inside .inter-card-container);
      // validation card (inside .validation-card-container).
      function renameIn(container, prefix) {
        var inputs = container.querySelectorAll('input,select,textarea');
        inputs.forEach(function (input) {
          if (!input.name) return;
          if (input.name.indexOf(prefix) === 0) return; // already prefixed
          input.name = prefix + input.name;
        });
      }
      renameIn(block.querySelector('.enact-card-container'),          'enact_'+idx+'_');
      renameIn(block.querySelector('.inter-card-container'),          'enact_'+idx+'_inter_');
      renameIn(block.querySelector('.validation-card-container'),     'enact_'+idx+'_valid_');
      // Block-level name and description inputs
      block.querySelectorAll('input[name="enactment_name"]').forEach(function(el){
        el.name = 'enact_'+idx+'_name';
      });
      block.querySelectorAll('textarea[name="enactment_description"]').forEach(function(el){
        el.name = 'enact_'+idx+'_description';
      });
    });

    var abilityCard = document.querySelector('.section-card[data-section="ability-type"]');
    if (abilityCard) {
      // Top-level ability type card fields are not in enact blocks. They
      // already have unique names (range, uses, energy_steps, etc.), but
      // item_dep and item_name may collide with an enactment's fields;
      // rename them too.
      abilityCard.querySelectorAll('[name="item_dep"]').forEach(function(el){el.name='ability_item_dep';});
      abilityCard.querySelectorAll('[name="item_name"]').forEach(function(el){el.name='ability_item_name';});
      // The remaining fields (energy_steps, action_steps, range, uses,
      // trigger, trigger_trait, phase_rounds, reverse_rounds, all_req,
      // reverse_knockout, knockout, hp, life, no_knockout) are kept as-is.
    }
  });

  // =========================================================================
  // Boot
  // =========================================================================

  document.addEventListener('DOMContentLoaded', function () {
    initialState = loadInitialState();
    if (initialState && initialState.ability_type) {
      onAbilityTypeChange(initialState.ability_type);
    }

    // Always start with one enactment block (for new mode) or recreate saved enactments
    if (initialState && initialState.enactments && initialState.enactments.length) {
      initialState.enactments.forEach(function (saved) {
        var block = makeEmptyEnactBlock();
        document.getElementById('enactments-container').appendChild(block);
        var nameInput = block.querySelector('input[name="enactment_name"]');
        if (nameInput) nameInput.value = saved.name || '';
        var descInput = block.querySelector('textarea[name="enactment_description"]');
        if (descInput) descInput.value = saved.description || '';
        block.querySelector('.enact-type-select').value = saved.type || '';
        block.dataset.enactType = saved.type || '';
        if (saved.type) {
          block.querySelector('.enact-card-container').innerHTML = renderEnactCard(saved.type, saved);
        }
        if (saved.interaction && saved.interaction.type) {
          block.querySelector('.inter-card-container').innerHTML = renderInterCard(saved.interaction.type, saved.interaction);
        }
        // validation card always present
        if (saved.interaction && saved.interaction.validation) {
          block.querySelector('.validation-card-container').innerHTML = renderValidationCard(saved.interaction.validation);
        } else {
          block.querySelector('.validation-card-container').innerHTML = renderValidationCard({});
        }
      });
    } else {
      var block = makeEmptyEnactBlock();
      document.getElementById('enactments-container').appendChild(block);
      block.querySelector('.validation-card-container').innerHTML = renderValidationCard({});
    }

    recalcAll();
  });
})();
