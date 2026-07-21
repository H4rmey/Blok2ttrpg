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
        countInput.value = String(index + 1);
        if (window.htmx) window.htmx.process(node);
        applyVisibility(node);
        dispatchChange(node);
      });
  });

  container.addEventListener("click", function (e) {
    if (e.target && e.target.classList.contains("remove-enactment")) {
      var block = e.target.closest(".enactment");
      if (block) {
        block.remove();
        dispatchChange(container);
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


