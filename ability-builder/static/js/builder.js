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

  var D = window.BUILDER_DATA;
  var ABILITY_TYPES = D.abilityTypes;
  var ENACT_TYPES = D.allEnactmentTypes;
  var INTER_TYPES = D.interactionTypes;

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
    var inner = '';
    var opt = '';
    function addTraitOption(name) {
      return '<option value="'+esc(name)+'" '+selected(data.shift_trait||data.engage_trait, name)+'>'+esc(name)+'</option>';
    }

    if (type === 'Execution') {
      inner = renderExecutionCard(data);
    } else if (type === 'Reaction') {
      inner = renderReactionCard(data);
    } else if (type === 'Phase') {
      inner = renderPhaseCard(data);
    } else if (type === 'Minion') {
      inner = renderMinionCard(data);
    } else {
      inner = '<p class="text-yellow-400">Unknown ability type: '+esc(type)+'</p>';
    }

    return '<div class="section-card ability-type-card bg-gray-800 rounded-lg border border-gray-700 p-5 space-y-4" data-section="ability-type" data-ability-type="'+esc(type)+'">'+inner+'</div>';
  }

  function renderExecutionCard(d) {
    d = d || {};
    var itemName = d.item_name || '';
    var itemWrap = d.item_dep ? '' : 'hidden';
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Execution</h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        itemDepCheckbox(d.item_dep),
        stepSelect('energy_steps', 'Energy ±', [-2,-1,0,1,2], d.energy_steps || 0),
        stepSelect('action_steps', 'Action ±', [-1,0,1], d.action_steps || 0),
      '</div>',
      '<div data-wrap="item-name" '+itemWrap+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(itemName)+'" placeholder="e.g. Silver Dagger" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderReactionCard(d) {
    d = d || {};
    var triggerNeedsTrait = d.trigger === 'Target makes a trait check';
    var triggerWrap = triggerNeedsTrait ? '' : 'hidden';
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Reaction</h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
        '<div><label class="block text-xs text-gray-400 mb-1">Trigger</label>',
        '<select name="trigger" onchange="onReactionTriggerChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
          '<option value="">-- Select --</option>',
          D.reactionTriggers.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.trigger,t)+'>'+esc(t)+'</option>';}).join(''),
        '</select></div>',
        '<div data-wrap="trigger-trait" '+triggerWrap+'>',
          '<label class="block text-xs text-gray-400 mb-1">Trigger Trait</label>',
          '<select name="trigger_trait" class="bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white w-full">',
            traitOptions('defense', d.trigger_trait),
          '</select></div>',
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        intSelect('range', 'Range', 1, 6, d.reaction_range || 1),
        intSelect('uses', 'Uses', 1, 3, d.reaction_uses || 1),
        itemDepCheckbox(d.item_dep),
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" placeholder="e.g. Shield" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderPhaseCard(d) {
    d = d || {};
    var kos = d.knockouts || [];
    var showKos = !d.no_knockout;
    return [
      '<h3 class="text-md font-semibold text-indigo-400">Ability Type — Phase</h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
        intSelect('phase_rounds', 'Phase Duration', 2, 5, d.phase_duration || 2, ' rounds'),
        intSelect('reverse_rounds', 'Reverse Rounds', 1, 5, d.reverse_phase_rounds || 2, ' rounds'),
      '</div>',
      '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">',
        checkbox('all_req', 'All knockout requirements have to be met', d.all_knockouts_req),
        checkbox('reverse_knockout', 'Knockout can be used on reverse phase', d.reverse_knockout_ok),
      '</div>',
      '<div>',
        '<label class="flex items-center gap-2 text-sm text-gray-300 mb-2">',
          '<input type="checkbox" name="no_knockout" onchange="onNoKnockoutChange(this)" '+checked(d.no_knockout)+' class="rounded bg-gray-700 border-gray-600">',
          '<span><strong>No knockout possible</strong> — the phase cannot be ended by any condition (costs extra)</span>',
        '</label>',
        '<div data-wrap="knockouts" '+hiddenIf(!showKos)+'>',
          '<div class="text-xs text-gray-400 uppercase mb-1">Knockouts</div>',
          knockoutList(kos),
        '</div>',
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function renderMinionCard(d) {
    d = d || {};
    return [
      '<h3 class="text-md font-semibold text-indigo-400 flex items-center gap-2">Ability Type — Minion <span class="text-xs text-yellow-400">(WIP)</span></h3>',
      overview(false),
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
        intSelect('hp', 'Health Bonus', 0, 5, d.hp_bonus || 0),
        intSelect('life', 'Extra Lifetime', 0, 5, d.extra_lifetime || 0),
        itemDepCheckbox(d.item_dep),
      '</div>',
      '<div data-wrap="item-name" '+hiddenIf(d.item_dep)+'>',
        '<label class="block text-xs text-gray-400 mb-1">Item Name</label>',
        '<input type="text" name="item_name" id="ability_item_name" value="'+esc(d.item_name||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white text-sm">',
      '</div>',
      breakdown(),
    ].join('\n');
  }

  function knockoutList(values) {
    values = values || [];
    var rows = '';
    rows += knockoutRow(values[0]);
    if (values.length > 1) {
      for (var i = 1; i < values.length; i++) rows += knockoutRow(values[i]);
    } else {
      rows += knockoutRow(null);
    }
    return '<div data-list="knockouts" class="space-y-2">'+rows+'</div>';
  }
  function knockoutRow(value) {
    var opts = '<option value="">-- Select --</option>' +
      D.knockoutOptions.map(function(k){
        return '<option value="'+esc(k)+'" '+selected(value,k)+'>'+esc(k)+'</option>';
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
    var type = card ? card.dataset.abilityType : '';
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
  function stepSelect(name, label, options, selectedValue) {
    var opts = options.map(function(o){
      var lbl = (o > 0 ? '+' : '') + o;
      return '<option value="'+o+'" '+selected(selectedValue, o)+'>'+lbl+'</option>';
    }).join('');
    return '<div class="flex items-center gap-2">'+
      '<span class="text-sm text-gray-400 whitespace-nowrap">'+esc(label)+'</span>'+
      '<select name="'+name+'" onchange="recalcAll()" class="bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white flex-1">'+opts+'</select>'+
    '</div>';
  }
  function intSelect(name, label, min, max, selectedValue, suffix) {
    suffix = suffix || 'm';
    var opts = '';
    for (var i = min; i <= max; i++) {
      opts += '<option value="'+i+'" '+selected(selectedValue, i)+'>'+i+suffix+'</option>';
    }
    return '<div><label class="block text-xs text-gray-400 mb-1">'+esc(label)+'</label>'+
      '<select name="'+name+'" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select></div>';
  }
  function intSelectFlat(name, label, min, max, selectedValue) {
    var opts = '';
    for (var i = min; i <= max; i++) opts += '<option value="'+i+'" '+selected(selectedValue, i)+'>'+i+'</option>';
    return '<div><label class="block text-xs text-gray-400 mb-1">'+esc(label)+'</label>'+
      '<select name="'+name+'" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select></div>';
  }
  function checkbox(name, label, value) {
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="'+name+'" onchange="recalcAll()" '+checked(value)+' class="rounded bg-gray-700 border-gray-600">'+
      esc(label)+
    '</label>';
  }
  function itemDepCheckbox(value) {
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="item_dep" onchange="onItemDepChange(this)" '+checked(value)+' class="rounded bg-gray-700 border-gray-600">'+
      'Has Item Dependency'+
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

  function renderEnactCard(type, data) {
    data = data || {};
    if (type === 'Enact Damage') return renderEnactDamage(data);
    if (type === 'Enact Healing') return renderEnactHealing(data);
    if (type === 'Enact Movement') return renderEnactMovement(data);
    if (type === 'Enact Proficiency Shift') return renderEnactProfShift(data);
    if (type === 'Enact Persistent Effect') return renderEnactPersistent(data);
    return '<div class="section-card enact-card bg-gray-800 rounded border border-gray-700 p-4 text-red-400">Unknown enact type: '+esc(type)+'</div>';
  }

  function enactTopStats(d) {
    return [
      '<div class="grid grid-cols-1 md:grid-cols-4 gap-3 text-sm">',
        statCard('Always Resolve', d.always ? 'Yes' : 'No', 'resolve'),
        statCard('Build Cost', '0', 'build'),
        statCard('Cast Cost', '0', 'cast'),
        statCard('Formula', '...', 'formula'),
      '</div>',
    ].join('\n');
  }

  function sourceSelect(d, name) {
    name = name || 'source';
    var opts = D.damageDiceOptions.map(function(o){
      return '<option value="'+esc(o)+'" '+selected(d.source, o)+'>'+esc(o)+'</option>';
    }).join('');
    opts += '<option value="trait" '+selected(d.source, 'trait')+'>Trait (1d10)</option>';
    opts += '<option value="previous" '+selected(d.source, 'previous')+'>Use result of previous enactment</option>';
    opts += '<option value="other" '+selected(d.source, 'other')+'>Another roll result</option>';
    return opts;
  }

  function renderEnactDamage(d) {
    d = d || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'offense') : '');
    var traitSelectHTML = '';
    if (src === 'trait') {
      traitSelectHTML = '<div data-wrap="source-trait"><label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="source_category" value="'+esc(srcCat)+'">'+
        '<select name="source_trait" onchange="onSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select></div>';
    }
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Damage" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Damage</h3>',
        enactTopStats(d),
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve (costs extra)', d.always),
          '<div>',
            '<label class="block text-xs text-gray-400 mb-1">Source</label>',
            '<select name="source" onchange="onEnactSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
              sourceSelect(d),
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
          intSelectFlat('flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0),
          '<div><label class="block text-xs text-gray-400 mb-1">Offensive Trait (extra die)</label>',
            '<select name="offense" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              '<option value="">None</option>' +
              D.offenseTraits.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.offensive_trait,t)+'>'+esc(t)+'</option>';}).join('') +
            '</select></div>',
        '</div>',
        breakdown(),
      '</div>'
    ].join('\n');
  }

  function renderEnactHealing(d) {
    d = d || {};
    var src = d.source || 'd4';
    var srcCat = d.source_category || (src === 'trait' ? (categoryOfTrait(d.source_trait) || 'offense') : '');
    var traitSelectHTML = '';
    if (src === 'trait') {
      traitSelectHTML = '<div data-wrap="source-trait"><label class="block text-xs text-gray-400 mb-1">Trait</label>'+
        '<input type="hidden" name="source_category" value="'+esc(srcCat)+'">'+
        '<select name="source_trait" onchange="onSourceTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
          '<option value="">-- Select --</option>' + traitOptionsGrouped(d.source_trait) +
        '</select></div>';
    }
    var otherWrap = src === 'other' ? '' : 'hidden';
    var prevWrap  = src === 'previous' ? '' : 'hidden';
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Healing" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Healing</h3>',
        enactTopStats(d),
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve (costs extra)', d.always),
          '<div><label class="block text-xs text-gray-400 mb-1">Source</label>',
          '<select name="source" onchange="onEnactSourceChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
            sourceSelect(d),
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
          intSelectFlat('flat', 'Flat Bonus', 0, 20, d.flat_bonus || 0),
          '<div><label class="block text-xs text-gray-400 mb-1">Medicine</label>',
            '<select name="medicine" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              '<option value="">None</option>'+
              '<option value="Medicine" '+selected(d.medicine_trait,'Medicine')+'>Medicine (1d10)</option>'+
            '</select></div>',
        '</div>',
        breakdown(),
      '</div>'
    ].join('\n');
  }

  function renderEnactMovement(d) {
    d = d || {};
    var dirs = d.directions && d.directions.length ? d.directions : ['Away'];
    var originMode = d.origin_mode || 'engager';
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Movement" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Movement</h3>',
        enactTopStats(d),
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          checkbox('always', 'Will always resolve (costs extra)', d.always),
          '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
            '<select name="origin_mode" onchange="onOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              '<option value="engager" '+selected(originMode,'engager')+'>Engager</option>'+
              '<option value="other" '+selected(originMode,'other')+'>Other Origin</option>'+
            '</select></div>',
        '</div>',
        '<div data-wrap="origin" '+hiddenIf(originMode !== 'other')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
          '<input type="text" name="origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          intSelect('distance', 'Distance', 1, 10, d.distance || 1, 'm'),
          '<div>',
            '<div class="flex items-center justify-between mb-1"><span class="text-xs text-gray-400">Directions</span>',
              '<button type="button" onclick="addDirection(this)" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Direction</button>',
            '</div>',
            directionsList(dirs),
        '</div>',
        breakdown(),
      '</div>'
    ].join('\n');
  }
  function directionsList(dirs) {
    var rows = dirs.map(function(dir){
      var opts = D.directionOptions.map(function(o){return '<option value="'+esc(o)+'" '+selected(dir,o)+'>'+esc(o)+'</option>';}).join('');
      return '<div class="flex items-center gap-2">'+
        '<select name="direction" onchange="recalcAll()" class="flex-1 bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+opts+'</select>'+
        '<button type="button" onclick="this.parentElement.remove();recalcAll()" class="bg-red-700 hover:bg-red-600 text-white px-2 py-1 rounded text-xs">−</button>'+
      '</div>';
    }).join('');
    return '<div data-list="directions" class="space-y-1">'+rows+'</div>';
  }

  function renderEnactProfShift(d) {
    d = d || {};
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Proficiency Shift" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Proficiency Shift</h3>',
        enactTopStats(d),
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
          checkbox('always', 'Will always resolve (costs extra)', d.always),
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">',
          '<div><label class="block text-xs text-gray-400 mb-1">Trait</label>',
            '<select name="shifted_trait" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              '<option value="">-- Select --</option>' + D.allTraits.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.shifted_trait,t)+'>'+esc(t)+'</option>';}).join(''),
            '</select></div>',
          '<div><label class="block text-xs text-gray-400 mb-1">Direction</label>',
            '<select name="shift_dir" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              D.shiftDirectionOptions.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.shift_dir,t)+'>'+esc(t)+'</option>';}).join(''),
            '</select></div>',
          intSelectFlat('shift_amount', 'Amount', 1, 5, d.shift_amount || 1),
          intSelectFlat('shift_uses', 'Uses', 1, 5, d.shift_uses || 1),
        '</div>',
        breakdown(),
      '</div>'
    ].join('\n');
  }

  function renderEnactPersistent(d) {
    d = d || {};
    var sols = (d.solutions && d.solutions.length ? d.solutions : ['Dexterity', 'Constitution']);
    return [
      '<div class="section-card enact-card bg-gray-800 rounded-lg border border-indigo-700 p-5 space-y-4" data-section="enact" data-enact-type="Enact Persistent Effect" data-build="0" data-cast="0">',
        '<h3 class="text-md font-semibold text-indigo-400">Enact — Persistent Effect</h3>',
        enactTopStats(d),
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          '<div><label class="block text-xs text-gray-400 mb-1">Name</label>',
            '<input type="text" name="effect_name" value="'+esc(d.effect_name||'Burning')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white"></div>',
          checkbox('always', 'Will always resolve (costs extra)', d.always),
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">',
          '<div><label class="block text-xs text-gray-400 mb-1">Applies</label>',
            '<select name="effect_type" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              D.persistentEffectTypes.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.effect_type,t)+'>'+esc(t)+'</option>';}).join('') +
            '</select></div>',
          intSelectFlat('duration', 'Duration', 2, 8, d.duration || 2),
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
        breakdown(),
      '</div>'
    ].join('\n');
  }
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

  // =========================================================================
  // Interaction cards
  // =========================================================================

  function renderInterCard(type, data) {
    data = data || {};
    if (type === 'Self')         return renderInterSelf(data);
    if (type === 'Direct')       return renderInterDirect(data);
    if (type === 'Ranged')       return renderInterRanged(data);
    if (type === 'Area')         return renderInterArea(data);
    if (type === 'Area of Effect') return renderInterAoE(data);
    return '<div class="section-card inter-card bg-gray-800 rounded border border-gray-700 p-4 text-red-400">Unknown inter type: '+esc(type)+'</div>';
  }

  function interTopStats(d) {
    return [
      '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">',
        statCard('Build Cost', '0', 'build'),
        statCard('Cast Cost', '0', 'cast'),
        statCard('Final', '...', 'formula'),
      '</div>',
    ].join('\n');
  }

  function renderInterSelf(d) {
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Self" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Self</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTopStats(),
          '<div class="text-sm text-gray-300"><strong>Type =</strong> Self + <strong>Target =</strong> Self + <strong>Counter =</strong> d8</div>',
          breakdown(),
        '</div>',
      '</div>'
    ].join('\n');
  }

  function renderInterDirect(d) {
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Direct" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Direct</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTopStats(),
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
            intSelect('range', 'Range', 1, 10, d.range || 1, 'm'),
            intSelectFlat('targets', 'Targets', 1, 5, d.targets || 1),
          '</div>',
          '<div>'+usePrevCheck(d.use_previous)+'</div>',
          breakdown(),
        '</div>',
      '</div>'
    ].join('\n');
  }
  function renderInterRanged(d) {
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Ranged" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Ranged</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTopStats(),
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
            intSelect('range', 'Range', 10, 20, d.range || 10, 'm'),
            intSelectFlat('targets', 'Targets', 1, 5, d.targets || 1),
          '</div>',
          '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">',
            checkbox('visible', 'Target may be not visible', d.visible_ok),
            checkbox('obstructed', 'Target may be obstructed', d.obstructed_ok),
            checkbox('remove_penalty', 'Remove engagement penalty', d.remove_penalty),
          '</div>',
          '<div>'+usePrevCheck(d.use_previous)+'</div>',
          breakdown(),
        '</div>',
      '</div>'
    ].join('\n');
  }
  function renderInterArea(d) {
    d = d || {};
    var om = d.origin_mode || 'engager';
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Area" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Area</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTopStats(),
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
            intSelect('radius', 'Radius', 1, 6, d.radius || 1, 'm'),
            intSelect('range', 'Range', 0, 10, d.range || 0, 'm'),
          '</div>',
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
            '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
              '<select name="origin_mode" onchange="onOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
                '<option value="engager" '+selected(om,'engager')+'>Engager</option>'+
                '<option value="other" '+selected(om,'other')+'>Other Origin</option>'+
              '</select></div>',
            '<div data-wrap="origin" '+hiddenIf(om!=='other')+'>',
              '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
              '<input type="text" name="origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
            '</div>',
          '</div>',
          '<div>'+usePrevCheck(d.use_previous)+'</div>',
          breakdown(),
        '</div>',
      '</div>'
    ].join('\n');
  }
  function renderInterAoE(d) {
    d = d || {};
    var om = d.origin_mode || 'engager';
    return [
      '<div class="section-card inter-card bg-gray-800 rounded-lg border border-cyan-700 p-4 space-y-3" data-section="interaction" data-inter-type="Area of Effect" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-cyan-300">Interaction — Area of Effect</h4>',
          '<span class="collapse-chevron text-cyan-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
          interTopStats(),
          '<div class="grid grid-cols-1 md:grid-cols-3 gap-3">',
            intSelect('radius', 'Radius', 1, 6, d.radius || 1, 'm'),
            intSelect('range', 'Range', 0, 10, d.range || 0, 'm'),
            intSelectFlat('duration', 'Duration', 2, 6, d.duration || 2),
          '</div>',
          '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">',
            '<div><label class="block text-xs text-gray-400 mb-1">Trigger Timing</label>',
              '<select name="timing" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
                D.aoeTriggerTimings.map(function(t){return '<option value="'+esc(t)+'" '+selected(d.timing,t)+'>'+esc(t)+'</option>';}).join(''),
              '</select></div>',
            '<div><label class="block text-xs text-gray-400 mb-1">Origin</label>',
              '<select name="origin_mode" onchange="onOriginModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
                '<option value="engager" '+selected(om,'engager')+'>Engager</option>'+
                '<option value="other" '+selected(om,'other')+'>Other Origin</option>'+
              '</select></div>',
          '</div>',
          '<div data-wrap="origin" '+hiddenIf(om!=='other')+'>',
            '<label class="block text-xs text-gray-400 mb-1">Origin Text</label>',
            '<input type="text" name="origin_text" value="'+esc(d.origin_text||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
          '</div>',
          checkbox('immune', 'Engager is immune', d.immune),
          '<div>'+usePrevCheck(d.use_previous)+'</div>',
          breakdown(),
        '</div>',
      '</div>'
    ].join('\n');
  }
  function usePrevCheck(v) {
    return '<label class="flex items-center gap-2 text-sm text-gray-300">'+
      '<input type="checkbox" name="use_previous" onchange="recalcAll()" '+checked(v)+' class="rounded bg-gray-700 border-gray-600">'+
      'Use result of previous interaction/validation (costs extra)'+
    '</label>';
  }

  // =========================================================================
  // Validation card
  // =========================================================================

  function renderValidationCard(d) {
    d = d || {};
    var mode = d.engage_mode || 'trait';
    var cat = d.engage_trait_category || (d.engage_trait ? (categoryOfTrait(d.engage_trait) || 'offense') : 'offense');
    var counters = d.counter_entries || d.counter_rolls || [];
    return [
      '<div class="section-card validation-card bg-gray-800 rounded-lg border border-rose-700 p-4 space-y-3" data-section="validation" data-build="0" data-cast="0">',
        '<button type="button" class="collapse-toggle w-full flex items-center justify-between text-left" aria-expanded="true" onclick="toggleCollapse(this)">',
          '<h4 class="text-sm font-semibold text-rose-300">Validation</h4>',
          '<span class="collapse-chevron text-rose-300 text-xs">&#9660;</span>',
        '</button>',
        '<div class="collapsible-content space-y-3 mt-3">',
        '<div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">',
          statCard('Build Cost', '0', 'build'),
          statCard('Cast Cost', '0', 'cast'),
          statCard('Final', '...', 'formula'),
        '</div>',
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">',
          '<div><label class="block text-xs text-gray-400 mb-1">Engage Roll Type</label>',
            '<select name="engage_mode" onchange="onEngageModeChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
              '<option value="trait" '+selected(mode,'trait')+'>Trait Roll</option>'+
              '<option value="generic" '+selected(mode,'generic')+'>Generic Roll</option>'+
              '<option value="other" '+selected(mode,'other')+'>Another roll result</option>'+
              '<option value="previous" '+selected(mode,'previous')+'>Use result of previous interaction</option>'+
            '</select></div>',
        '</div>',
        '<div data-wrap="engage-trait" '+hiddenIf(mode!=='trait')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Trait</label>',
          '<input type="hidden" name="engage_trait_category" value="'+esc(cat)+'">'+
          '<select name="engage_trait" onchange="onEngageTraitChange(this)" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            '<option value="">-- Select --</option>' + traitOptionsGrouped(d.engage_trait) +
          '</select></div>',
        '<div data-wrap="engage-generic" '+hiddenIf(mode!=='generic')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Die</label>',
          '<select name="engage_die" onchange="recalcAll()" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">'+
            D.genericDieOptions.map(function(o){return '<option value="'+esc(o)+'" '+selected(d.engage_die,o)+'>'+esc(o)+'</option>';}).join('') +
          '</select></div>',
        '<div data-wrap="engage-other" '+hiddenIf(mode!=='other')+'>',
          '<label class="block text-xs text-gray-400 mb-1">Other Roll Text</label>',
          '<input type="text" name="engage_other" value="'+esc(d.engage_other||'')+'" class="w-full bg-gray-700 border border-gray-600 rounded px-2 py-2 text-white">',
        '</div>',
        '<div data-wrap="engage-previous" '+hiddenIf(mode!=='previous')+'>',
          '<p class="text-xs text-yellow-400">Engagement roll = result of previous interaction/validation (costs extra)</p>',
        '</div>',
        '<div>',
          '<div class="flex items-center justify-between mb-1">',
            '<span class="text-xs text-gray-400 uppercase">Counter Rolls</span>',
            '<button type="button" onclick="addCounter(this)" class="bg-indigo-600 hover:bg-indigo-500 text-white px-2 py-1 rounded text-xs">+ Counter</button>',
          '</div>',
          counterList(counters),
        '</div>',
        breakdown(),
        '</div>',
      '</div>'
    ].join('\n');
  }
  function counterList(items) {
    items = items.length ? items : [{type:'defense', trait:'Reflex'}, {type:'defense', trait:'Constitution'}];
    var rows = items.map(function(item){
      if (typeof item === 'string') item = {type:'defense', trait:item};
      return counterRow(item);
    }).join('');
    return '<div data-list="counters" class="space-y-2">'+rows+'</div>';
  }
  function counterRow(item) {
    var opts = '<option value="defense" '+selected(item.type,'defense')+'>Defensive Trait (default)</option>'+
               '<option value="general" '+selected(item.type,'general')+'>General Trait (costs extra)</option>'+
               '<option value="offense" '+selected(item.type,'offense')+'>Offensive Trait (costs extra)</option>'+
               '<option value="previous" '+selected(item.type,'previous')+'>Use result of previous (costs extra)</option>';
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
    updateInterOptions(block);
    if (!val) return;
    host.innerHTML = renderEnactCard(val, cur);
    recalcAll();
  };

  function updateInterOptions(block) {
    var select = block.querySelector('.inter-type-select');
    select.innerHTML = '<option value="">-- Select --</option>' +
      INTER_TYPES.map(function(t){return '<option value="'+esc(t)+'">'+esc(t)+'</option>';}).join('');
  }

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

  window.onEnactSourceChange = function (sel) {
    var c = sel.closest('.section-card');
    var v = sel.value;
    setWrap(c, 'source-trait', v === 'trait');
    setWrap(c, 'source-other', v === 'other');
    setWrap(c, 'source-previous', v === 'previous');
  };

  window.onSourceTraitChange = function (sel) {
    var c = sel.closest('.section-card');
    var hidden = c.querySelector('[name="source_category"]');
    if (hidden) hidden.value = categoryOfTrait(sel.value);
    recalcAll();
  };

  window.onOriginModeChange = function (sel) {
    var c = sel.closest('.section-card');
    setWrap(c, 'origin', sel.value === 'other');
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

  window.addDirection = function (btn) {
    var list = btn.parentElement.parentElement.querySelector('[data-list="directions"]');
    var first = list.querySelector('select[name="direction"]');
    var val = first ? first.value : 'Away';
    var opts = D.directionOptions.map(function(o){return '<option value="'+esc(o)+'" '+selected(val,o)+'>'+esc(o)+'</option>';}).join('');
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
    var row = document.createElement('div');
    row.innerHTML = counterRow({type:'defense', trait:''});
    list.appendChild(row);
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
      '<div class="collapsible-content space-y-3">',
        '<div class="flex items-center gap-3 flex-wrap">',
          '<label class="text-xs text-gray-400 flex items-center gap-1">Interaction: ',
            '<select onchange="onInterTypeChange(this)" class="inter-type-select bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white">',
              '<option value="">-- Select --</option>',
            '</select>',
          '</label>',
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
        '<div class="inter-card-container space-y-2"></div>',
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
      var triggerTrait = card.querySelector('[data-wrap="trigger-trait"]') && !card.querySelector('[data-wrap="trigger-trait"]').hidden
        ? card.querySelector('[name="trigger_trait"]').value : '';
      setOut(card, 'formula', 'Reaction, '+uses+' uses/round, range '+range+'m, trigger: '+trigger+(triggerTrait?(' of type '+triggerTrait):''));
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
      var kos = Array.from(card.querySelectorAll('[name="knockout"]')).map(function(s){return s.value;}).filter(function(v){return v && v !== 'None';});
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

  function calcEnact(card) {
    if (!card) return;
    var cfg = getEnactConfig(card.dataset.enactType);
    var lines = [];
    var build = 0, energy = cfg.base_cost ? (cfg.base_cost.energy_cost || 0) : 0;
    var formula = '';

    var always = findPerk(cfg.perks, 'always_resolve');
    if (always && readBool(card, 'always')) {
      build += always.add_cost || 0;
      energy += always.energy_cost || 0;
      lines.push('Always resolve (add '+(always.add_cost||0)+', energy '+(always.energy_cost||0)+')');
    }

    var srcEl = card.querySelector('[name="source"]');
    if (srcEl) {
      var s = srcEl.value;
      var tiers = cfg.dice_tiers || {d4:0,d6:1,d8:2,d10:3,d12:4};
      var tierCost = cfg.dice_tier_cost || {add_cost:0, energy_cost:0};
      if (tiers[s] !== undefined) {
        var tier = tiers[s] || 0;
        formula = '1'+s;
        build += tier * tierCost.add_cost;
        energy += tier * tierCost.energy_cost;
        lines.push('Source 1'+s+' (add '+(tier*tierCost.add_cost)+', energy '+(tier*tierCost.energy_cost)+')');
      } else if (s === 'trait') {
        var cat = (card.querySelector('[name="source_category"]') || {}).value || 'offense';
        var tName = (card.querySelector('[name="source_trait"]') || {}).value || '(trait)';
        formula = '1d10 ('+cat+' trait)';
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
        formula = 'previous enactment result';
        build += prevCost ? prevCost.add_cost : 0;
        energy += prevCost ? prevCost.energy_cost : 0;
        lines.push('Use result of previous enactment (add '+(prevCost?prevCost.add_cost:0)+', energy '+(prevCost?prevCost.energy_cost:0)+')');
      } else if (s === 'other') {
        var txt = (card.querySelector('[name="other"]')||{}).value || '';
        formula = txt || '(other roll result)';
        var otherCost = findPerk(cfg.perks, 'use_previous'); // use previous perk for other rolls
        build += otherCost ? otherCost.add_cost : 0;
        energy += otherCost ? otherCost.energy_cost : 0;
        lines.push('Use another roll result (add '+(otherCost?otherCost.add_cost:0)+', energy '+(otherCost?otherCost.energy_cost:0)+')');
      }
    }

    var flatEl = card.querySelector('[name="flat"]');
    if (flatEl) {
      var flat = readNumber(card, 'flat', 0);
      var flatCost = findPerk(cfg.perks, 'flat_bonus');
      var flatAdd = flatCost ? flatCost.add_cost : 0;
      var flatEnergy = flatCost ? flatCost.energy_cost : 0;
      if (flat > 0) { formula += ' + '+flat; build += flat * flatAdd; energy += flat * flatEnergy; lines.push('Flat +'+flat+' (add '+(flat*flatAdd)+', energy '+(flat*flatEnergy)+')'); }
    }

    var offenseEl = card.querySelector('[name="offense"]');
    if (offenseEl && offenseEl.value) {
      var off = findPerk(cfg.perks, 'offensive_trait');
      build += off ? off.add_cost : 0;
      energy += off ? off.energy_cost : 0;
      lines.push('Offensive trait die ('+offenseEl.value+') (add '+(off?off.add_cost:0)+', energy '+(off?off.energy_cost:0)+')');
      formula += ' + 1d8 ('+offenseEl.value+')';
    }
    var medEl = card.querySelector('[name="medicine"]');
    if (medEl && medEl.value) {
      var med = findPerk(cfg.perks, 'medicine_trait');
      build += med ? med.add_cost : 0;
      energy += med ? med.energy_cost : 0;
      lines.push('Medicine trait (add '+(med?med.add_cost:0)+', energy '+(med?med.energy_cost:0)+')');
      formula += ' + 1d10 (Medicine)';
    }

    var distanceEl = card.querySelector('[name="distance"]');
    if (distanceEl) {
      var dist = readNumber(card, 'distance', 1);
      var dirArr = Array.from(card.querySelectorAll('[name="direction"]')).map(function(s){return s.value;}).filter(Boolean);
      var originMode = (card.querySelector('[name="origin_mode"]')||{}).value;
      var distCost = cfg.distance_cost || {add_cost:0, energy_cost:0};
      if (dist > 1) { build += (dist-1) * distCost.add_cost; energy += (dist-1) * distCost.energy_cost; lines.push('Distance +'+(dist-1)+'m (add '+((dist-1)*distCost.add_cost)+', energy '+((dist-1)*distCost.energy_cost)+')'); }
      var origin = originMode === 'other' ? (card.querySelector('[name="origin_text"]')||{}).value || '(other)' : 'Engager';
      var originCost = findPerk(cfg.perks, 'other_origin');
      if (originMode === 'other') { build += originCost ? originCost.add_cost : 0; energy += originCost ? originCost.energy_cost : 0; lines.push('Other origin (add '+(originCost?originCost.add_cost:0)+', energy '+(originCost?originCost.energy_cost:0)+')'); }
      var extraDir = findPerk(cfg.perks, 'extra_direction');
      if (dirArr.length > 1) { build += (dirArr.length-1) * (extraDir?extraDir.add_cost:0); energy += (dirArr.length-1) * (extraDir?extraDir.energy_cost:0); lines.push('Extra direction '+(dirArr.length-1)+' (add '+((dirArr.length-1)*(extraDir?extraDir.add_cost:0))+', energy '+((dirArr.length-1)*(extraDir?extraDir.energy_cost:0))+')'); }
      var freeDir = findPerk(cfg.perks, 'free_direction');
      var freeCount = dirArr.filter(function(d){return d==='Free';}).length;
      if (freeCount > 0) { build += freeCount * (freeDir?freeDir.add_cost:0); energy += freeCount * (freeDir?freeDir.energy_cost:0); lines.push('Free direction '+freeCount+'x (add '+(freeCount*(freeDir?freeDir.add_cost:0))+', energy '+(freeCount*(freeDir?freeDir.energy_cost:0))+')'); }
      formula = 'Move target '+dist+'m '+(dirArr.join(' or '))+' from '+origin;
    }

    var shiftTraitEl = card.querySelector('[name="shifted_trait"]');
    if (shiftTraitEl) {
      var trait = shiftTraitEl.value || '(trait)';
      var dir   = (card.querySelector('[name="shift_dir"]')||{}).value || 'UP';
      var amt   = readNumber(card, 'shift_amount', 1);
      var uses  = readNumber(card, 'shift_uses', 1);
      var amtCost = cfg.shift_amount_cost || {add_cost:0, energy_cost:0};
      var usesCost = cfg.shift_uses_cost || {add_cost:0, energy_cost:0};
      if (amt > 1) { build += (amt-1) * amtCost.add_cost; energy += (amt-1) * amtCost.energy_cost; lines.push('Shift amount +'+(amt-1)+' (add '+((amt-1)*amtCost.add_cost)+', energy '+((amt-1)*amtCost.energy_cost)+')'); }
      if (uses > 1) { build += (uses-1) * usesCost.add_cost; energy += (uses-1) * usesCost.energy_cost; lines.push('Shift uses +'+(uses-1)+' (add '+((uses-1)*usesCost.add_cost)+', energy '+((uses-1)*usesCost.energy_cost)+')'); }
      formula = 'Shift '+trait+' '+dir+' by '+amt+' for '+uses+' uses';
    }

    var effName = card.querySelector('[name="effect_name"]');
    if (effName) {
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
      if (sols.length === 1) { build += singleSol ? singleSol.add_cost : 0; energy += singleSol ? singleSol.energy_cost : 0; lines.push('Only one solution (add '+(singleSol?singleSol.add_cost:0)+', energy '+(singleSol?singleSol.energy_cost:0)+')'); }
      var effType = (card.querySelector('[name="effect_type"]')||{}).value || '(effect)';
      formula = (effName.value||'Effect')+' applies '+effType+' for '+dur+' rounds, solutions: '+(sols.join(' or ')||'(none)');
    }

    setOut(card, 'resolve', readBool(card, 'always') ? 'Yes' : 'No');
    setOut(card, 'build', build);
    setOut(card, 'cast', energy);
    setOut(card, 'formula', formula);
    card.dataset.build = build;
    card.dataset.cast = energy;
    fillList(card, lines);
  }

  function calcInter(card) {
    if (!card) return;
    var cfg = getInterConfig(card.dataset.interType);
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

    setOut(card, 'build', build);
    setOut(card, 'cast', energy);
    setOut(card, 'formula', formula);
    card.dataset.build = build;
    card.dataset.cast = energy;
    fillList(card, lines);
  }

  function calcValidation(card) {
    if (!card) return;
    var cfg = getValidationConfig();
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
      build += generic ? generic.add_cost : 0;
      energy += generic ? generic.energy_cost : 0;
      lines.push('Engage roll: generic '+die+' (add '+(generic?generic.add_cost:0)+', energy '+(generic?generic.energy_cost:0)+')');
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
    var singleCounter = cfg.counter.single_counter_cost;
    if (counterCount === 1) { build += singleCounter.add_cost; energy += singleCounter.energy_cost; lines.push('Only one counter option (add '+singleCounter.add_cost+', energy '+singleCounter.energy_cost+')'); }

    formula = (formula || 'engage vs counters') + ' vs ' + (counters.map(function(c){return c.trait;}).join(' or ') || '(no counters)');

    setOut(card, 'build', build);
    setOut(card, 'cast', energy);
    setOut(card, 'formula', formula);
    card.dataset.build = build;
    card.dataset.cast = energy;
    fillList(card, lines);
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

    var blocks = document.querySelectorAll('.enactment-block');
    blocks.forEach(function (block, idx) {
      block.dataset.index = idx; // re-number in display order
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
      // Block-level name input
      block.querySelectorAll('input[name="enactment_name"]').forEach(function(el){
        el.name = 'enact_'+idx+'_name';
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
        block.querySelector('.enact-type-select').value = saved.type || '';
        block.dataset.enactType = saved.type || '';
        if (saved.type) {
          block.querySelector('.enact-card-container').innerHTML = renderEnactCard(saved.type, saved);
        }
        if (saved.interaction && saved.interaction.type) {
          var interSelect = block.querySelector('.inter-type-select');
          var opts = INTER_TYPES.map(function(t){return '<option value="'+esc(t)+'" '+(t===saved.interaction.type?'selected':'')+'>'+esc(t)+'</option>';}).join('');
          interSelect.innerHTML = '<option value="">-- Select --</option>' + opts;
          interSelect.value = saved.interaction.type;
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
