// AI character generator (client half).
//
// Credentials are BYOK: the user enters an OpenAI-compatible endpoint, key and
// model. They are kept only in this browser (localStorage) when the user grants
// consent, otherwise in memory for this page session. The actual provider call
// and the validation retry loop run server-side; this file streams progress and
// imports the accepted character.
(function () {
  "use strict";

  var STORAGE_KEY = "blok2-ai-settings";

  // memorySettings holds credentials for the current session when the user has
  // not consented to localStorage.
  var memorySettings = null;
  var draftYAML = "";

  // Live generation state, so progress and elapsed time can be shown and the
  // request aborted.
  var controller = null;
  var timer = null;
  var startedAt = 0;

  function $(id) {
    return document.getElementById(id);
  }

  function el(id) {
    return $(id);
  }

  // -- Settings persistence --------------------------------------------------

  function readSaved() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      var s = JSON.parse(raw);
      if (s && s.endpoint && s.apiKey && s.model) return s;
    } catch (e) {
      /* ignore malformed storage */
    }
    return null;
  }

  function writeSaved(s) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(s));
    } catch (e) {
      /* storage may be unavailable; fall back to memory only */
    }
  }

  function clearSaved() {
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch (e) {
      /* ignore */
    }
  }

  // settingsFromPanel reads the raw field values from the settings panel. The
  // text input is authoritative for the model (the select copies into it via
  // aiPickModel), so a typed id is never overridden by the fetched list.
  function settingsFromPanel() {
    var model = value(el("ai-model-text")) || value(el("ai-model-select"));
    return {
      endpoint: value(el("ai-endpoint")).trim(),
      apiKey: value(el("ai-apikey")).trim(),
      model: model.trim(),
    };
  }

  function value(node) {
    return node && typeof node.value === "string" ? node.value : "";
  }

  // persist writes the given settings either to localStorage (consented) or to
  // the in-memory session variable.
  function persist(s) {
    memorySettings = s;
    if (el("ai-consent") && el("ai-consent").checked) {
      writeSaved(s);
    } else {
      clearSaved();
    }
  }

  // effectiveSettings resolves the credentials to use for a generation.
  function effectiveSettings() {
    var saved = readSaved();
    if (saved) return saved;
    if (memorySettings) return memorySettings;
    return settingsFromPanel();
  }

  // -- Settings panel --------------------------------------------------------

  window.aiToggleSettings = function () {
    var panel = el("ai-settings-panel");
    if (!panel) return;
    panel.hidden = !panel.hidden;
    if (!panel.hidden) fillSettings();
  };

  window.aiPickModel = function () {
    var sel = el("ai-model-select");
    var txt = el("ai-model-text");
    if (sel && txt && sel.value) txt.value = sel.value;
  };

  // fillSettings loads saved/memory settings into the panel fields.
  function fillSettings() {
    var saved = readSaved();
    var s = saved || memorySettings;
    if (!s) return;
    if (el("ai-endpoint")) el("ai-endpoint").value = s.endpoint || "";
    if (el("ai-apikey")) el("ai-apikey").value = s.apiKey || "";
    if (el("ai-model-text")) el("ai-model-text").value = s.model || "";
    if (el("ai-consent") && saved) el("ai-consent").checked = true;
  }

  window.aiFetchModels = function () {
    var s = settingsFromPanel();
    var msg = el("ai-settings-msg");
    if (!s.endpoint || !s.apiKey) {
      setSettingsMsg("Enter an endpoint and API key first.");
      return;
    }
    setSettingsMsg("Fetching models...");
    fetch("/ai/models", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint: s.endpoint, apiKey: s.apiKey }),
    })
      .then(function (resp) {
        return resp.json().then(function (body) {
          return { ok: resp.ok, body: body };
        });
      })
      .then(function (res) {
        if (!res.ok) throw new Error(res.body && res.body.error ? res.body.error : "could not fetch models");
        var sel = el("ai-model-select");
        if (sel) {
          sel.innerHTML = "";
          (res.body.models || []).forEach(function (id) {
            var opt = document.createElement("option");
            opt.value = id;
            opt.textContent = id;
            sel.appendChild(opt);
          });
        }
        setSettingsMsg((res.body.models || []).length + " models loaded.");
      })
      .catch(function (err) {
        setSettingsMsg("Could not fetch models: " + err.message + " (type the model id instead).");
      });
  };

  window.aiClearCredentials = function () {
    clearSaved();
    memorySettings = null;
    if (el("ai-endpoint")) el("ai-endpoint").value = "";
    if (el("ai-apikey")) el("ai-apikey").value = "";
    if (el("ai-model-text")) el("ai-model-text").value = "";
    if (el("ai-model-select")) el("ai-model-select").innerHTML = "";
    if (el("ai-consent")) el("ai-consent").checked = false;
    setSettingsMsg("Saved credentials cleared.");
  };

  function setSettingsMsg(text) {
    var msg = el("ai-settings-msg");
    if (!msg) return;
    msg.textContent = text;
    msg.hidden = !text;
  }

  // -- Generation ------------------------------------------------------------

  function showError(text) {
    var e = el("ai-error");
    if (!e) return;
    e.textContent = text || "";
    e.hidden = !text;
  }

  function showProgress() {
    var wrap = el("ai-progress");
    if (wrap) wrap.hidden = false;
  }

  // Playful status lines cycled while the model works, so a slow generation
  // still feels alive (and human) instead of staring at a bar.
  var FUN_MESSAGES = [
    "Consulting the dice gods...",
    "Rolling initiative on your backstory...",
    "Balancing the point budget...",
    "Arguing with the rulebook...",
    "Teaching the AI what a d12 is...",
    "Spending your points responsibly...",
    "Checking the tables twice...",
    "Summoning a level-appropriate hero...",
    "Untangling the perk tree...",
    "Convincing the GM it's balanced...",
    "Inking the character sheet...",
    "Warming up the polyhedrals...",
    "Rolling for stats... (you hope it's good)",
    "Cross-referencing 400 pages of rules...",
    "Giving the perk a cool name...",
  ];
  var funIndex = 0;
  var ticks = 0;
  var overrideText = "";
  var overrideUntil = 0;

  function elapsedText() {
    var s = Math.round((Date.now() - startedAt) / 1000);
    var m = Math.floor(s / 60);
    s = s % 60;
    return m > 0 ? m + "m " + s + "s" : s + "s";
  }

  // renderStatus shows the real status when one is fresh, otherwise the
  // rotating fun message, always with the elapsed time.
  function renderStatus() {
    var t = el("ai-progress-text");
    if (!t) return;
    var msg =
      Date.now() < overrideUntil
        ? overrideText
        : FUN_MESSAGES[funIndex % FUN_MESSAGES.length];
    t.textContent = msg + "  ·  " + elapsedText();
  }

  // setStatus pins a real status line for a few seconds before the fun
  // messages resume.
  function setStatus(text, holdMs) {
    overrideText = text || "";
    overrideUntil = text ? Date.now() + (holdMs || 4000) : 0;
    renderStatus();
  }

  function startTimer() {
    stopTimer();
    timer = setInterval(tick, 1000);
  }

  function tick() {
    ticks++;
    // Rotate the fun message every three seconds.
    if (ticks % 3 === 0) funIndex++;
    renderStatus();
  }

  function setBusy(busy) {
    var btn = el("ai-generate-btn");
    if (btn) btn.disabled = busy;
    var cancel = el("ai-cancel-btn");
    if (cancel) cancel.hidden = !busy;
  }

  window.aiGenerate = function () {
    showError("");
    var desc = value(el("ai-description")).trim();
    if (!desc) {
      showError("Describe your character first.");
      return;
    }

    var panel = el("ai-settings-panel");
    // Prefer what is currently in the panel (so edits to the endpoint/key/model
    // take effect even when settings were previously saved), falling back to
    // saved session memory when the panel is empty.
    var s = settingsFromPanel();
    if (!s.endpoint || !s.apiKey || !s.model) s = effectiveSettings();
    if (!s.endpoint || !s.apiKey || !s.model) {
      if (panel) panel.hidden = false;
      fillSettings();
      setSettingsMsg("Enter your API endpoint, key and model, then Generate again.");
      return;
    }
    persist(s);

    var level = parseInt(value(el("ai-level")), 10) || 1;
    controller = typeof AbortController !== "undefined" ? new AbortController() : null;
    startedAt = Date.now();
    ticks = 0;
    funIndex = 0;
    setBusy(true);
    showProgress();
    setStatus("Generating character...");
    startTimer();

    fetch("/ai/generate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      signal: controller ? controller.signal : undefined,
      body: JSON.stringify({
        description: desc,
        level: level,
        endpoint: s.endpoint,
        apiKey: s.apiKey,
        model: s.model,
      }),
    })
      .then(function (resp) {
        if (!resp.ok || !resp.body) {
          return resp.text().then(function (t) {
            throw new Error(t || "generation failed");
          });
        }
        return readStream(resp.body);
      })
      .catch(function (err) {
        if (err && err.name === "AbortError") return;
        showError(err.message || String(err));
      })
      .then(function () {
        stopTimer();
        setBusy(false);
        controller = null;
      });
  };

  function stopTimer() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  window.aiCancel = function () {
    if (controller) controller.abort();
    stopTimer();
    setBusy(false);
    setStatus("Cancelled.", 60000);
  };

  // readStream consumes the NDJSON progress stream.
  function readStream(body) {
    var reader = body.getReader();
    var dec = new TextDecoder();
    var buf = "";
    var doneEvent = null;

    function pump() {
      return reader.read().then(function (res) {
        if (res.done) return finish();
        buf += dec.decode(res.value, { stream: true });
        var idx;
        while ((idx = buf.indexOf("\n")) >= 0) {
          var line = buf.slice(0, idx).trim();
          buf = buf.slice(idx + 1);
          if (!line) continue;
          var ev;
          try {
            ev = JSON.parse(line);
          } catch (e) {
            continue;
          }
          handleEvent(ev);
          if (ev.type === "done") doneEvent = ev;
        }
        return pump();
      });
    }

    function finish() {
      if (!doneEvent) return;
      if (doneEvent.valid) {
        return aiImport(doneEvent.yaml);
      }
      draftYAML = doneEvent.yaml || "";
      showKeepModal(doneEvent.message || "The character did not fully validate.");
    }

    return pump();
  }

  function handleEvent(ev) {
    switch (ev.type) {
      case "status":
        setStatus(ev.message);
        break;
      case "progress":
        // Progress only tells us output is arriving; the rotating messages and
        // the elapsed clock carry the "still working" signal.
        break;
      case "retry":
        setStatus(ev.message, 6000);
        break;
      case "error":
        showError(ev.message);
        stopTimer();
        setStatus("Stopped.", 60000);
        break;
      case "done":
        stopTimer();
        setStatus(ev.valid ? "Character ready." : "Finished with validation issues.", 60000);
        break;
    }
  }

  // aiImport posts YAML to the server, which repairs and stores it.
  function aiImport(yaml) {
    if (!yaml) {
      showError("No character data to import.");
      return Promise.resolve();
    }
    return fetch("/ai/import", {
      method: "POST",
      headers: { "Content-Type": "application/x-yaml" },
      body: yaml,
    })
      .then(function (resp) {
        return resp.json().then(function (body) {
          return { ok: resp.ok, body: body };
        });
      })
      .then(function (res) {
        if (!res.ok) throw new Error(res.body && res.body.error ? res.body.error : "import failed");
        var modal = el("generate-character-modal");
        if (modal) modal.close();
        window.location.href = "/characters/" + res.body.id;
      })
      .catch(function (err) {
        showError("Could not save the character: " + err.message);
      });
  }

  // -- Keep / discard --------------------------------------------------------

  function showKeepModal(summary) {
    var gen = el("generate-character-modal");
    if (gen) gen.close();
    var text = el("ai-keep-summary");
    if (text) {
      text.textContent =
        "Failed to generate character, still want to keep what was created anyways? " + summary;
    }
    var modal = el("ai-keep-modal");
    if (modal) modal.showModal();
  }

  window.aiKeepDraft = function () {
    var modal = el("ai-keep-modal");
    if (modal) modal.close();
    aiImport(draftYAML);
  };

  window.aiDiscardDraft = function () {
    var modal = el("ai-keep-modal");
    if (modal) modal.close();
    var gen = el("generate-character-modal");
    if (gen) gen.showModal();
  };

  // -- Init ------------------------------------------------------------------

  document.addEventListener("DOMContentLoaded", function () {
    var modal = el("generate-character-modal");
    if (!modal) return;
    // Closing the modal while a generation is running aborts it, so a slow
    // provider call does not keep running invisibly.
    modal.addEventListener("close", function () {
      if (controller) controller.abort();
      stopTimer();
      setBusy(false);
    });
    var s = readSaved();
    if (s) {
      memorySettings = s;
      fillSettings();
    }
  });
})();
