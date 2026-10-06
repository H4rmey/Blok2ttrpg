// Character page tabs.
//
// The character sheet is a single page with four panels (Traits, Skills, Perks,
// Extras). The active panel is reflected in the URL hash so that a server
// redirect can land the user on a specific tab: applying a condition comes back
// to #tab-skills, saving a perk to #tab-perks, and so on.
(function () {
  var nav = document.getElementById('char-tabs');
  if (!nav) return;

  // The perk list is refetched the first time the Perks tab is opened so its
  // costs are current without the user pressing refresh. htmx.ajax is called
  // directly rather than going through a custom hx-trigger, because the element
  // is inside a hidden panel at page load and the trigger would not fire.
  var perksLoaded = false;
  function loadPerksOnce() {
    if (perksLoaded) return;
    var list = document.getElementById('tab-perks-list');
    if (!list || typeof htmx === 'undefined') return;
    var url = list.getAttribute('data-load-url');
    if (!url) return;
    perksLoaded = true;
    // Append quiet=1 so the server omits the "already up to date" notice that
    // is only meaningful when the user manually presses the refresh button.
    var loadUrl = url + (url.indexOf('?') >= 0 ? '&' : '?') + 'quiet=1';
    htmx.ajax('GET', loadUrl, { target: list, swap: 'innerHTML' });
  }

  function activateById(id) {
    var btn = nav.querySelector('[data-tab="' + id + '"]');
    if (!btn) return;
    nav.querySelectorAll('.char-tab').forEach(function (b) { b.classList.remove('active'); });
    document.querySelectorAll('.char-tab-panel').forEach(function (p) { p.classList.remove('active'); });
    btn.classList.add('active');
    var panel = document.getElementById(id);
    if (panel) panel.classList.add('active');
    if (id === 'tab-perks') loadPerksOnce();
  }

  function activateFromHash() {
    var hash = window.location.hash.replace('#', '');
    if (hash && document.getElementById(hash)) activateById(hash);
  }

  nav.addEventListener('click', function (e) {
    var btn = e.target.closest('.char-tab');
    if (!btn || !btn.dataset.tab) return;
    history.replaceState(null, '', '#' + btn.dataset.tab);
    activateById(btn.dataset.tab);
  });

  activateFromHash();
  window.addEventListener('hashchange', activateFromHash);
})();
