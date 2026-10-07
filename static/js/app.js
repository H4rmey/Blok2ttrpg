// ---------------------------------------------------------------------------
// Package import modal. The Import Package button loads the built-in package
// browser into the modal body via HTMX; these helpers just toggle visibility.
// ---------------------------------------------------------------------------
function openPackageModal() {
  var modal = document.getElementById("package-modal");
  if (modal) modal.hidden = false;
}
function closePackageModal() {
  var modal = document.getElementById("package-modal");
  if (modal) modal.hidden = true;
}

// The condition picker uses its own modal rather than sharing the package one,
// so a half-open package browser cannot be replaced out from under the user.
function openConditionModal() {
  var modal = document.getElementById("condition-modal");
  if (modal) modal.hidden = false;
}

function closeConditionModal() {
  var modal = document.getElementById("condition-modal");
  if (modal) modal.hidden = true;
}

// The shift picker follows the same rule - its own modal, no sharing with the
// package or condition browsers.
function openShiftModal() {
  var modal = document.getElementById("shift-modal");
  if (modal) modal.hidden = false;
}

function closeShiftModal() {
  var modal = document.getElementById("shift-modal");
  if (modal) modal.hidden = true;
}

// Perk import modal. The Import button loads the built-in perk browser
// into the modal body via HTMX; these helpers just toggle visibility.
function openPerkModal() {
  var modal = document.getElementById("perk-modal");
  if (modal) modal.hidden = false;
}
function closePerkModal() {
  var modal = document.getElementById("perk-modal");
  if (modal) modal.hidden = true;
}

// New Perk modal: picking "Passive" as the Type routes into the existing
// HTMX passive-picker flow instead of the enactment builder. This has to be
// a client-side intercept rather than a server redirect, because
// /perks/passives is only ever loaded as an HTMX fragment into
// #perk-modal-body; a plain full-page GET to it would render a bare
// fragment with no page shell. Without JavaScript the form falls through to
// its plain GET submit, which the server-side guard in handlePerks refuses
// safely (see abilities.go) rather than opening a broken builder.
//
// The handler is delegated from the document rather than bound directly to
// #new-perk-form. A direct binding runs once at script-parse time and silently
// does nothing if the form is not in the DOM yet, or if the page region holding
// it is later replaced by an HTMX swap. Delegation keeps it working in both
// cases.
//
// The picker modal is revealed before the request is issued, not after it
// resolves. Revealing it only in a promise callback meant any rejected request
// (htmx rejects on target, send, abort and timeout errors) left the New Perk
// dialog closed and the picker still hidden - an empty overlay with no
// indication of what went wrong.
document.addEventListener("submit", function (e) {
  var form = e.target && e.target.closest ? e.target.closest("#new-perk-form") : null;
  if (!form) return;

  var typeSel = document.getElementById("new-perk-type");
  if (!typeSel || typeSel.value !== "passive") return;

  e.preventDefault();

  var nameInput = document.getElementById("new-perk-name");
  var name = nameInput ? nameInput.value : "";
  var charID = form.getAttribute("data-character-id") || "";

  var newPerkModal = document.getElementById("new-perk-modal");
  if (newPerkModal && newPerkModal.close) newPerkModal.close();

  // Show the shell first so the user always gets visible feedback, even if the
  // fragment request is slow or fails outright.
  openPerkModal();

  var body = document.getElementById("perk-modal-body");
  var url = "/perks/passives?character=" + encodeURIComponent(charID) +
    "&name=" + encodeURIComponent(name);

  htmx.ajax("GET", url, { target: "#perk-modal-body", swap: "innerHTML" })
    .catch(function (err) {
      if (console && console.error) console.error("passive picker failed to load", err);
      if (body) {
        body.innerHTML =
          '<div class="pkg-library"><div class="pkg-library-head">' +
          '<h2>Passives</h2>' +
          '<button class="btn" type="button" onclick="closePerkModal()">Close</button>' +
          '</div><p class="pkg-blocked">The passive list could not be loaded. ' +
          'Check your connection and try again.</p></div>';
      }
    });
});



// ---------------------------------------------------------------------------
// Mobile navigation drawer.
//
// The top bar links do not fit beside the brand and the character name on a
// phone, so below the CSS breakpoint they collapse behind a hamburger. The
// handler is delegated from the document rather than bound to the button, so it
// keeps working on pages whose header is swapped in by HTMX. Everything below
// degrades to a plain always-visible nav if scripting is unavailable, because
// the drawer is only hidden inside the media query.
// ---------------------------------------------------------------------------
(function () {
  function links() {
    return document.getElementById("nav-links");
  }
  function toggleBtn() {
    return document.getElementById("nav-toggle");
  }
  function setOpen(open) {
    var nav = links();
    var btn = toggleBtn();
    if (!nav) return;
    nav.classList.toggle("open", open);
    if (btn) btn.setAttribute("aria-expanded", open ? "true" : "false");
  }

  document.addEventListener("click", function (e) {
    var btn = e.target && e.target.closest ? e.target.closest("#nav-toggle") : null;
    if (btn) {
      e.preventDefault();
      setOpen(!links() || !links().classList.contains("open"));
      return;
    }
    // Any click outside the drawer closes it, including a click on one of its
    // own links: navigating away should not leave the panel open behind the
    // next page's paint.
    var nav = links();
    if (nav && nav.classList.contains("open") && !nav.contains(e.target)) {
      setOpen(false);
    }
  });

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape") setOpen(false);
  });
})();

// ---------------------------------------------------------------------------
// Documentation reader: sidebar, scoped search, active-section tracking,
// heading anchors and back-to-top.
//
// This runs only on a page marked with [data-doc-page] (the rulebook and the
// changelog), and it exits immediately elsewhere, so nothing here costs anything
// on the character sheet.
//
// Search is deliberately client-side. The whole document is already in the DOM,
// so the page is its own index: there is no server endpoint to call, no index to
// build at deploy time, and no way for the index to disagree with the text. It
// is also scoped to the current page, which keeps results predictable - a query
// on the changelog never returns rulebook hits.
// ---------------------------------------------------------------------------
(function () {
  var page = document.querySelector("[data-doc-page]");
  if (!page) return;

  var body = document.getElementById("doc-body");
  var nav = document.getElementById("doc-nav");
  var input = document.getElementById("doc-search-input");
  var results = document.getElementById("doc-search-results");
  var empty = document.getElementById("doc-search-empty");
  var sidebar = document.getElementById("doc-sidebar");
  var sidebarToggle = document.getElementById("doc-sidebar-toggle");
  var toTop = document.getElementById("back-to-top");

  // -- Index -----------------------------------------------------------------
  //
  // One record per section: its heading, the chapter it belongs to, and the
  // text of everything between this heading and the next. Built once on load
  // from the rendered page.
  var index = [];

  function buildIndex() {
    if (!body) return;
    var nodes = body.querySelectorAll("h1[id], h2[id], h3[id], h4[id]");
    var chapter = "";
    nodes.forEach(function (h) {
      if (h.tagName === "H2") chapter = h.textContent.trim();
      var text = "";
      // Walk forward to the next heading of any level, collecting body text.
      var n = h.nextElementSibling;
      while (n && !/^H[1-6]$/.test(n.tagName)) {
        text += " " + (n.textContent || "");
        n = n.nextElementSibling;
      }
      index.push({
        id: h.id,
        title: h.textContent.trim(),
        chapter: h.tagName === "H2" ? "" : chapter,
        haystack: (h.textContent + " " + text).toLowerCase(),
        text: text.replace(/\s+/g, " ").trim(),
      });
    });
  }

  // snippet returns a short window of text around the first hit, so a result
  // shows why it matched rather than just that it did.
  function snippet(rec, q) {
    var i = rec.text.toLowerCase().indexOf(q);
    if (i < 0) return "";
    var start = Math.max(0, i - 40);
    var s = rec.text.slice(start, start + 160);
    return (start > 0 ? "..." : "") + s + (start + 160 < rec.text.length ? "..." : "");
  }

  function clearSearch() {
    if (results) {
      results.innerHTML = "";
      results.hidden = true;
    }
    if (empty) empty.hidden = true;
    if (nav) nav.hidden = false;
  }

  function runSearch(raw) {
    var q = raw.trim().toLowerCase();
    if (q.length < 2) {
      clearSearch();
      return;
    }
    var hits = index.filter(function (rec) {
      return rec.haystack.indexOf(q) >= 0;
    });
    // Title matches first: someone typing "energy" almost always wants the
    // section called Energy, not the first paragraph that mentions it.
    hits.sort(function (a, b) {
      var at = a.title.toLowerCase().indexOf(q) >= 0 ? 0 : 1;
      var bt = b.title.toLowerCase().indexOf(q) >= 0 ? 0 : 1;
      return at - bt;
    });
    hits = hits.slice(0, 30);

    if (nav) nav.hidden = true;
    if (!results) return;
    results.innerHTML = "";
    hits.forEach(function (rec) {
      var li = document.createElement("li");
      var a = document.createElement("a");
      a.href = "#" + rec.id;
      a.className = "doc-search-hit";
      a.dataset.target = rec.id;
      var title = document.createElement("span");
      title.className = "doc-search-hit-title";
      title.textContent = rec.chapter ? rec.chapter + " -> " + rec.title : rec.title;
      a.appendChild(title);
      var sn = snippet(rec, q);
      if (sn) {
        var p = document.createElement("span");
        p.className = "doc-search-hit-snippet";
        p.textContent = sn;
        a.appendChild(p);
      }
      li.appendChild(a);
      results.appendChild(li);
    });
    results.hidden = hits.length === 0;
    if (empty) empty.hidden = hits.length !== 0;
  }

  if (input) {
    var timer = null;
    input.addEventListener("input", function () {
      clearTimeout(timer);
      var v = input.value;
      timer = setTimeout(function () { runSearch(v); }, 120);
    });
    input.addEventListener("keydown", function (e) {
      if (e.key === "Escape") {
        input.value = "";
        clearSearch();
        input.blur();
        return;
      }
      if (e.key === "Enter") {
        // Jump straight to the first hit, so a search can be done entirely
        // from the keyboard.
        var first = results && results.querySelector("a");
        if (first) {
          e.preventDefault();
          first.click();
        }
      }
    });
  }

  // -- Active section --------------------------------------------------------
  //
  // Highlights the sidebar entry for whatever is currently on screen. Falls
  // back to doing nothing at all where IntersectionObserver is unavailable,
  // since the sidebar links work regardless.
  function markActive(id) {
    if (!sidebar) return;
    sidebar.querySelectorAll("a.active").forEach(function (a) {
      a.classList.remove("active");
    });
    var link = sidebar.querySelector('a[data-target="' + id + '"]');
    if (link) {
      link.classList.add("active");
      // Keep the highlighted entry in view in a long contents list.
      if (link.scrollIntoView) link.scrollIntoView({ block: "nearest" });
    }
  }

  if (body && window.IntersectionObserver) {
    var observed = body.querySelectorAll("h1[id], h2[id], h3[id]");
    var io = new IntersectionObserver(function (entries) {
      // The topmost intersecting heading wins, which matches what a reader
      // perceives as "where I am".
      var visible = entries
        .filter(function (e) { return e.isIntersecting; })
        .sort(function (a, b) { return a.boundingClientRect.top - b.boundingClientRect.top; });
      if (visible.length) markActive(visible[0].target.id);
    }, { rootMargin: "-80px 0px -70% 0px" });
    observed.forEach(function (h) { io.observe(h); });
  }

  // -- Sidebar collapse (narrow screens) ------------------------------------
  if (sidebarToggle && sidebar) {
    sidebarToggle.addEventListener("click", function () {
      var open = sidebar.classList.toggle("open");
      sidebarToggle.setAttribute("aria-expanded", open ? "true" : "false");
    });
    // Following a link on a phone should reveal the section, not leave the
    // contents panel covering it.
    sidebar.addEventListener("click", function (e) {
      if (e.target.closest && e.target.closest("a")) {
        sidebar.classList.remove("open");
        sidebarToggle.setAttribute("aria-expanded", "false");
      }
    });
  }

  // -- Heading anchors ------------------------------------------------------
  //
  // A clickable "#" beside each heading, so a section can be linked directly
  // into a chat or a ticket without hunting for the id in the HTML.
  if (body) {
    body.querySelectorAll("h1[id], h2[id], h3[id], h4[id]").forEach(function (h) {
      var a = document.createElement("a");
      a.className = "heading-anchor";
      a.href = "#" + h.id;
      a.textContent = "#";
      a.setAttribute("aria-label", "Link to this section");
      h.appendChild(a);
    });
  }

  // -- Back to top ----------------------------------------------------------
  if (toTop) {
    window.addEventListener("scroll", function () {
      toTop.hidden = window.scrollY < 600;
    });
    toTop.addEventListener("click", function () {
      window.scrollTo({ top: 0, behavior: "smooth" });
    });
  }

  buildIndex();
})();

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
  // Per-row cascade: a repeatable-row field carries data-row-visibility-when
  // set to a sibling row_field key (not a prefixed name), because row input
  // names are renumbered by index. Resolve the controlling input within the
  // same .row via its data-row-key so each row cascades independently.
  scope.querySelectorAll("[data-row-visibility-when]").forEach(function (el) {
    var ctrlKey = el.getAttribute("data-row-visibility-when");
    var want = el.getAttribute("data-show-when");
    var row = el.closest(".row");
    var ctrl = row ? row.querySelector('[data-row-key="' + ctrlKey + '"]') : null;
    var cur = fieldControlValue(ctrl);
    el.style.display = cur === want ? "" : "none";
  });
}


document.addEventListener("change", function () { applyVisibility(); });
document.addEventListener("input", function () { applyVisibility(); });

// ---------------------------------------------------------------------------
// Selected-option information indicator.
// A dropdown rendered with class "opt-select" lives inside a ".field" and is
// paired with a trailing ".option-info" (little "i") and a ".option-info-text"
// (below the select). When an option carries data-info:
//   - data-render-info="true"  -> render the info as text below the select
//                                  (via .option-info-text), hide the "i".
//   - otherwise                -> populate the "i" tooltip and show it.
// Options with no info clear both. The native <option> title attribute still
// provides a hover tooltip on each item in the open list.
// ---------------------------------------------------------------------------
function syncOptionInfo(sel) {
  var field = sel.closest(".field");
  if (!field) return;
  var icon = field.querySelector(".option-info");
  var text = field.querySelector(".option-info-text");
  var opt = sel.options[sel.selectedIndex];
  var info = opt ? (opt.getAttribute("data-info") || "") : "";
  var renderBelow = opt && opt.getAttribute("data-render-info") === "true";
  if (icon) { icon.hidden = true; }
  if (text) { text.hidden = true; text.textContent = ""; }
  if (!info) return;
  if (renderBelow) {
    if (text) { text.textContent = info; text.hidden = false; }
  } else if (icon) {
    icon.setAttribute("data-tooltip", info);
    icon.hidden = false;
  }
}

function syncAllOptionInfo(root) {
  var scope = root || document;
  scope.querySelectorAll("select.opt-select").forEach(function (sel) {
    syncOptionInfo(sel);
  });
}

document.addEventListener("change", function (e) {
  var t = e.target;
  if (t && t.matches && t.matches("select.opt-select")) {
    syncOptionInfo(t);
  }
});


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
// Perk builder: add/remove enactments. Each enactment is fetched from the
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

  // Recompute the whole perk cost from the current form state. The backend
  // recalculates everything from scratch on each request, so we never do any
  // incremental add/remove math on the client. We post the full form directly
  // via htmx.ajax (rather than relying on the form's change-trigger) so the
  // request always fires - including the very first auto-added enactment on
  // page load - and so add/remove stay perfectly symmetric. Deferred to the
  // next frame so the DOM mutation (append/remove + renumberEnactments) is
  // fully applied before the form is serialized.
  function recalcPerkCost() {
    var form = document.getElementById("perk-form");
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



  // Show or hide an enactment's Interaction and Validation regions. An
  // enactment owns its target when it is the first one, or when the author
  // ticked "different target than the enactment before it"; otherwise it
  // reuses the previous enactment's target and the two regions are hidden.
  // They stay in the DOM (and keep posting) so hiding them never discards
  // stored values, matching how the server renders the block.
  function applyTargetRegions(block, index) {
    var toggle = block.querySelector(".new-target-toggle");
    var owns = index === 0 || (toggle && toggle.checked);
    ["region-interaction", "region-validation"].forEach(function (cls) {
      var region = block.querySelector("." + cls);
      if (region) region.hidden = !owns;
    });
  }

  container.addEventListener("change", function (e) {
    if (e.target && e.target.classList.contains("new-target-toggle")) {
      var block = e.target.closest(".enactment");
      if (block) {
        applyTargetRegions(block, parseInt(block.getAttribute("data-index") || "0", 10));
      }
    }
  });

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
      // The first enactment always owns its target, so it never shows the
      // "different target" checkbox; later blocks do. Renumbering can turn a
      // second enactment into the first one, so this has to be reapplied here.
      var ntField = block.querySelector(".new-target-field");
      if (ntField) ntField.hidden = i === 0;
      applyTargetRegions(block, i);


      block.querySelectorAll("[name]").forEach(function (input) {
        input.name = input.name.replace(/^en\d+_/, "en" + i + "_");
      });
      block.querySelectorAll("[hx-vals]").forEach(function (el) {
        try {
          var v = JSON.parse(el.getAttribute("hx-vals"));
          v.index = String(i);
          ["name", "shift_name"].forEach(function (k) {
            if (typeof v[k] === "string") { v[k] = v[k].replace(/en\d+_/, "en" + i + "_"); }
          });
          el.setAttribute("hx-vals", JSON.stringify(v));
        } catch (err) { /* leave as-is on parse error */ }
      });
    });
    countInput.value = String(blocks.length);
  }

  addBtn.addEventListener("click", function () {
    var index = nextIndex();
    var url = window.BUILDER.enactmentEndpoint + "?index=" + encodeURIComponent(index);
    // Pass the current perk type so the new enactment's dropdowns apply the
    // perk-type allowed/blocked enactment filtering (and Feature 4 region
    // toggles) consistently with the server-rendered blocks.
    var typeSel = document.querySelector('#perk-form select[name="type"]');
    if (typeSel && typeSel.value) {
      url += "&atype=" + encodeURIComponent(typeSel.value);
    }
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
        recalcPerkCost();
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
        recalcPerkCost();
      }

    }
  });


  // When editing/importing an existing perk the enactments are rendered on
  // the server, so normalize their indices and Remove-button visibility once
  // on load. Otherwise (a brand-new perk) the first enactment is free and
  // always present: load one automatically.
  if (container.querySelectorAll(".enactment").length > 0) {
    renumberEnactments();
  } else if (nextIndex() === 0) {
    addBtn.click();
  }
})();



// ---------------------------------------------------------------------------
// Perk builder autosave. Whenever the builder form changes we POST the full
// form to /builder/autosave (debounced). The server persists the perk and
// returns its id; we write that back into the hidden perk_id input so a
// brand-new perk keeps updating the same record on subsequent saves.
// Autosave is skipped until the perk has a name (the server enforces this
// too, returning 204 No Content).
//
// Two things prevent the "save creates a duplicate" bug:
//   1. We assign a stable perk id on the client at load time, so every
//      request (autosave AND the manual Save submit) carries the same id from
//      the very first keystroke. Even if Save fires before an autosave response
//      returns, the server updates the one record instead of appending a new
//      one.
//   2. Autosaves are serialized: only one request is in flight at a time, so
//      two concurrent autosaves can never both create a record.
// ---------------------------------------------------------------------------
(function () {
  var timer = null;
  var inFlight = false;
  var pending = false;

  // Give a brand-new perk a stable id up front. Editing an existing perk
  // already has its server id in the hidden field, so we only fill it if empty.
  (function ensurePerkID() {
    var idInput = document.getElementById("perk-id");
    if (idInput && !idInput.value) {
      idInput.value = "perk-" + Date.now() + "-" + Math.floor(Math.random() * 1e9);
    }
  })();

  function scheduleAutosave() {
    var nameInput = document.getElementById("perk-name");
    if (nameInput && nameInput.value.trim().length === 0) return;
    if (!document.getElementById("perk-form")) return;
    clearTimeout(timer);
    timer = setTimeout(runAutosave, 600);
  }

  function runAutosave() {
    var form = document.getElementById("perk-form");
    if (!form) return;
    var nameInput = document.getElementById("perk-name");
    if (nameInput && nameInput.value.trim().length === 0) return;
    if (inFlight) {
      // A save is already running; run once more when it returns so the latest
      // form state is still persisted.
      pending = true;
      return;
    }
    inFlight = true;
    var data = new FormData(form);
    var body = new URLSearchParams(data).toString();
    fetch("/builder/autosave", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: body,
    })
      .then(function (r) {
        if (r.status === 204) return null;
        var hid = r.headers.get("X-Perk-ID");
        if (hid) {
          var idInput = document.getElementById("perk-id");
          if (idInput) idInput.value = hid;
        }
        // The server returns the refreshed top bar (stat_cards) in the JSON
        // payload so the perk-points figures follow the edit. Current-vital
        // inputs are never overwritten while focused or dirty, so a live
        // update cannot reset Current back to Max mid-typing.
        return r
          .json()
          .catch(function () { return null; })
          .then(function (payload) {
            if (payload && payload.stat_cards) applyStatCards(payload.stat_cards);
            return null;
          });
      })
      .catch(function () { /* ignore transient autosave errors */ })
      .then(function () {
        inFlight = false;
        if (pending) {
          pending = false;
          runAutosave();
        }
      });
  }

  // applyStatCards swaps the top-bar numbers while preserving what the player
  // is editing: focused inputs, and inputs whose live value differs from the
  // fresh markup (typed but not yet saved), keep their live value.
  function applyStatCards(html) {
    var target = document.getElementById("stat-cards");
    if (!target) return;
    var live = {};
    target.querySelectorAll("input").forEach(function (el) {
      if (!el.name) return;
      live[el.name] = { value: el.value, active: document.activeElement === el };
    });
    target.innerHTML = html;
    if (window.htmx) window.htmx.process(target);
    Object.keys(live).forEach(function (name) {
      var st = live[name];
      var fresh = target.querySelector('input[name="' + name + '"]');
      if (!fresh) return;
      if (st.active || fresh.value !== st.value) {
        fresh.value = st.value;
      }
    });
  }

  document.addEventListener("change", function (e) {
    if (e.target && e.target.closest && e.target.closest("#perk-form")) {
      scheduleAutosave();
    }
  });
  document.addEventListener("input", function (e) {
    if (e.target && e.target.closest && e.target.closest("#perk-form")) {
      scheduleAutosave();
    }
  });
})();

// Re-apply visibility after HTMX swaps (perk-type fields, enactment reloads).
document.addEventListener("htmx:afterSwap", function (e) {
  applyVisibility();
  syncAllOptionInfo(e && e.target ? e.target : document);
});

// Preserve Current-vital inputs across any #stat-cards swap (level saves,
// skill saves, builder cost previews): a refresh must update the numbers
// without resetting a value the player is editing back to Max.
var stashedVitals = {};
document.addEventListener("htmx:oobBeforeSwap", function (e) {
  if (e.target && e.target.id === "stat-cards") stashVitalInputs(e.target);
});
document.addEventListener("htmx:oobAfterSwap", function (e) {
  if (e.target && e.target.id === "stat-cards") restoreVitalInputs(e.target);
});
function stashVitalInputs(target) {
  stashedVitals = {};
  if (!target) return;
  target.querySelectorAll("input").forEach(function (el) {
    if (!el.name) return;
    stashedVitals[el.name] = { value: el.value, active: document.activeElement === el };
  });
}
function restoreVitalInputs(target) {
  if (!target) return;
  Object.keys(stashedVitals).forEach(function (name) {
    var st = stashedVitals[name];
    var fresh = target.querySelector('input[name="' + name + '"]');
    if (!fresh) return;
    if (st.active || (st.value !== "" && fresh.value !== st.value)) {
      fresh.value = st.value;
    }
  });
  stashedVitals = {};
}

document.addEventListener("DOMContentLoaded", function () {
  applyVisibility();
  syncAllOptionInfo();
});


// ---------------------------------------------------------------------------
// Character sheet: live preview of the stats bar while the level field is
// being typed. The authoritative save happens on change via the input's own
// hx-post (see character_bar.html), which persists the level and pushes the
// refreshed skills tab out of band; this preview only re-renders the numbers
// without saving, so typing never commits a half-entered value.
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

// Typing in the level box only previews (no save); committing happens on
// change through the input's hx-post. The forms themselves autosave via HTMX
// hx-trigger="change" and the server returns updated stat_cards, so no
// separate recalc is needed for field/dropdown changes inside a form.
document.addEventListener("change", function (e) {
  if (e.target && e.target.id === "character-level") {
    // Let the hx-post save run first; only fall back to a preview when htmx
    // is absent.
    if (!window.htmx) recalcCharacterStats();
  }
});
document.addEventListener("input", function (e) {
  if (e.target && e.target.id === "character-level") {
    recalcCharacterStats();
  }
});

// ---------------------------------------------------------------------------
// Name gate: the rest of the builder stays disabled until the perk has a
// name. This mirrors the original flow where naming the perk comes first.
// ---------------------------------------------------------------------------
(function () {
  var nameInput = document.getElementById("perk-name");
  var gate = document.getElementById("builder-gate");
  if (!nameInput || !gate) return;

  var hint = document.getElementById("name-hint");
  var saveBtn = document.getElementById("save-perk");

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

