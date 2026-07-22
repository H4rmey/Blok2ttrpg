// Theme toggle with persistence.
(function () {
  var KEY = "blok2-theme";
  var saved = localStorage.getItem(KEY);
  if (saved) {
    document.documentElement.setAttribute("data-theme", saved);
  }
  document.addEventListener("click", function (e) {
    if (e.target && e.target.id === "theme-toggle") {
      var cur = document.documentElement.getAttribute("data-theme") === "light" ? "dark" : "light";
      document.documentElement.setAttribute("data-theme", cur);
      localStorage.setItem(KEY, cur);
    }
  });
})();

// ---------------------------------------------------------------------------
// Conditional field visibility (visibility_when / show_when).
// A field carries data-visibility-when="<controlling field name>" and
// data-show-when="<value>". It is shown only when the controlling input
// currently holds that value (checkboxes use "true"/"false").
// ---------------------------------------------------------------------------
function fieldControlValue(input) {
  if (!input) return "";
  if (input.type === "checkbox") return input.checked ? "true" : "false";
  return input.value;
}

function applyVisibility(root) {
  var scope = root || document;
  scope.querySelectorAll("[data-visibility-when]").forEach(function (el) {
    var ctrlName = el.getAttribute("data-visibility-when");
    var want = el.getAttribute("data-show-when");
    // Search within the enclosing form so prefixed names resolve correctly.
    var form = el.closest("form") || document;
    var ctrl = form.querySelector('[name="' + ctrlName + '"]');
    var cur = fieldControlValue(ctrl);
    el.style.display = cur === want ? "" : "none";
  });
}

document.addEventListener("change", function () { applyVisibility(); });
document.addEventListener("input", function () { applyVisibility(); });

// ---------------------------------------------------------------------------
// Collapsible sections.
// ---------------------------------------------------------------------------
document.addEventListener("click", function (e) {
  var toggle = e.target.closest && e.target.closest(".collapse-toggle");
  if (!toggle) return;
  var expanded = toggle.getAttribute("aria-expanded") !== "false";
  toggle.setAttribute("aria-expanded", String(!expanded));
  var content = toggle.closest(".enactment, .region, section, fieldset");
  if (content) {
    var body = content.querySelector(".collapsible-content");
    if (body) body.hidden = expanded;
  }
});

// ---------------------------------------------------------------------------
// Repeatable rows (solutions / states).
// ---------------------------------------------------------------------------
function renumberRows(rowsEl) {
  var name = rowsEl.getAttribute("data-rows-name");
  var body = rowsEl.querySelector(".rows-body");
  var rows = body.querySelectorAll(".row");
  rows.forEach(function (row, i) {
    row.querySelectorAll("[data-row-key]").forEach(function (input) {
      var key = input.getAttribute("data-row-key");
      input.name = name + "_" + i + "_" + key;
    });
  });
  var count = rowsEl.querySelector(".rows-count");
  if (count) count.value = String(rows.length);
}

document.addEventListener("click", function (e) {
  var add = e.target.closest && e.target.closest(".add-row");
  if (add) {
    var rowsEl = add.closest(".rows");
    var tpl = document.getElementById(add.getAttribute("data-row-template"));
    if (rowsEl && tpl) {
      var node = tpl.content.firstElementChild.cloneNode(true);
      rowsEl.querySelector(".rows-body").appendChild(node);
      renumberRows(rowsEl);
      dispatchChange(rowsEl);
    }
    return;
  }
  var rem = e.target.closest && e.target.closest(".remove-row");
  if (rem) {
    var container = rem.closest(".rows");
    var row = rem.closest(".row");
    if (row) row.remove();
    if (container) {
      renumberRows(container);
      dispatchChange(container);
    }
  }
});

function dispatchChange(el) {
  var evt = new Event("change", { bubbles: true });
  el.dispatchEvent(evt);
}

// ---------------------------------------------------------------------------
// Ability builder: add/remove enactments. Each enactment is fetched from the
// server as an HTML partial so its fields stay config-driven.
// ---------------------------------------------------------------------------
(function () {
  var addBtn = document.getElementById("add-enactment");
  if (!addBtn || !window.BUILDER) return;

  var container = document.getElementById("enactments");
  var countInput = document.getElementById("enactment_count");

  function nextIndex() {
    return parseInt(countInput.value || "0", 10);
  }

  // Recompute the whole ability cost from the current form state. The backend
  // recalculates everything from scratch on each request, so we never do any
  // incremental add/remove math on the client. We post the full form directly
  // via htmx.ajax (rather than relying on the form's change-trigger) so the
  // request always fires - including the very first auto-added enactment on
  // page load - and so add/remove stay perfectly symmetric. Deferred to the
  // next frame so the DOM mutation (append/remove + renumberEnactments) is
  // fully applied before the form is serialized.
  function recalcAbilityCost() {
    var form = document.getElementById("ability-form");
    if (!form) return;
    requestAnimationFrame(function () {
      if (window.htmx && window.htmx.ajax) {
        window.htmx.ajax("POST", "/builder/cost", {
          source: form,
          target: "#cost-badge",
          swap: "innerHTML",
        });
      } else {
        dispatchChange(form);
      }
    });
  }



  // Renumber every enactment block so their indices are contiguous starting
  // at 0. This rewrites the "en<i>_" prefix on every named input/select and
  // the hx-vals index, then syncs enactment_count to the real block count.
  // Without this, removing a block leaves a gap (e.g. en0 removed, en1 kept)
  // and the surcharge / first-free logic keys off the wrong slot, which
  // caused removing and re-adding the first enactment to charge extra points.
  function renumberEnactments() {
    var blocks = container.querySelectorAll(".enactment");
    blocks.forEach(function (block, i) {
      block.setAttribute("data-index", String(i));
      var label = block.querySelector(".enactment-head strong");
      if (label) label.textContent = "Enactment " + i;
      // The first enactment is always present and free; it cannot be removed,
      // so hide its Remove button.
      var removeBtn = block.querySelector(".remove-enactment");
      if (removeBtn) removeBtn.style.display = i === 0 ? "none" : "";

      block.querySelectorAll("[name]").forEach(function (input) {
        input.name = input.name.replace(/^en\d+_/, "en" + i + "_");
      });
      block.querySelectorAll("[hx-vals]").forEach(function (el) {
        try {
          var v = JSON.parse(el.getAttribute("hx-vals"));
          v.index = String(i);
          el.setAttribute("hx-vals", JSON.stringify(v));
        } catch (err) { /* leave as-is on parse error */ }
      });
    });
    countInput.value = String(blocks.length);
  }

  addBtn.addEventListener("click", function () {
    var index = nextIndex();
    var url = window.BUILDER.enactmentEndpoint + "?index=" + encodeURIComponent(index);
    fetch(url)
      .then(function (r) { return r.text(); })
      .then(function (html) {
        var wrap = document.createElement("div");
        wrap.innerHTML = html.trim();
        var node = wrap.firstElementChild;
        container.appendChild(node);
        renumberEnactments();
        if (window.htmx) window.htmx.process(node);
        applyVisibility(node);
        recalcAbilityCost();
      });
  });


  container.addEventListener("click", function (e) {
    if (e.target && e.target.classList.contains("remove-enactment")) {
      var block = e.target.closest(".enactment");
      // Guard: never allow the first enactment (index 0) to be removed.
      if (block && block === container.querySelector(".enactment")) return;
      if (block) {
        block.remove();
        renumberEnactments();
        recalcAbilityCost();
      }

    }
  });


  // The first enactment is free and always present: load one automatically
  // when the builder opens with none yet.
  if (container.children.length === 0 && nextIndex() === 0) {
    addBtn.click();
  }
})();


// Re-apply visibility after HTMX swaps (ability-type fields, enactment reloads).
document.addEventListener("htmx:afterSwap", function () { applyVisibility(); });
document.addEventListener("DOMContentLoaded", function () { applyVisibility(); });

// ---------------------------------------------------------------------------
// Character sheet: recalculate the whole stats bar (trait points, ability
// points) on the server whenever any field changes. We always post the full
// form (including the level input, which lives outside the form via the
// form="" attribute) so the backend recomputes everything from scratch, just
// like the ability builder does.
// ---------------------------------------------------------------------------
function recalcCharacterStats() {
  var form = document.getElementById("character-form");
  if (!form) return;
  var url = form.getAttribute("data-stats-url");
  if (!url) return;
  var target = document.getElementById("stat-cards");
  if (!target) return;
  var levelInput = document.getElementById("character-level");

  // Serialize the full form. The level input lives outside the <form> (it is
  // linked via the form="" attribute) so we add it explicitly.
  var data = new FormData(form);
  if (levelInput) data.set("level", levelInput.value);
  // Editable vital inputs (current HP/Energy) live in the stats bar, outside
  // the form, linked via form="character-form". FormData does not include such
  // associated elements, so add them explicitly.
  document.querySelectorAll(".stats-bar [name^='current_']").forEach(function (el) {
    data.set(el.name, el.value);
  });
  var body = new URLSearchParams(data).toString();

  fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body,
  })
    .then(function (r) { return r.text(); })
    .then(function (html) { target.innerHTML = html; });
}

// Delegated listeners so recalculation fires no matter when the elements were
// added and regardless of event-bubbling quirks with the external level input.
document.addEventListener("change", function (e) {
  var t = e.target;
  if (!t) return;
  if (t.id === "character-level" || (t.closest && t.closest("#character-form"))) {
    recalcCharacterStats();
  }
});
document.addEventListener("input", function (e) {
  if (e.target && e.target.id === "character-level") {
    recalcCharacterStats();
  }
});

// ---------------------------------------------------------------------------
// Name gate: the rest of the builder stays disabled until the ability has a
// name. This mirrors the original flow where naming the ability comes first.
// ---------------------------------------------------------------------------
(function () {
  var nameInput = document.getElementById("ability-name");
  var gate = document.getElementById("builder-gate");
  if (!nameInput || !gate) return;

  var hint = document.getElementById("name-hint");
  var saveBtn = document.getElementById("save-ability");

  function syncGate() {
    var named = nameInput.value.trim().length > 0;
    gate.classList.toggle("gated", !named);
    if (hint) hint.style.display = named ? "none" : "";
    if (saveBtn) saveBtn.disabled = !named;
  }

  nameInput.addEventListener("input", syncGate);
  nameInput.addEventListener("change", syncGate);
  syncGate();
})();
