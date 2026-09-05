package main

// ---------------------------------------------------
// BEAST BROWSER CHROME (toolbar + tab strip)
//
// This used to be a static "shell" page with an <iframe> where every
// other page (internal or external) was loaded. That's what broke
// search and most real browsing: sites like Google send
// X-Frame-Options / CSP headers that refuse to render inside someone
// else's iframe, and even when a site *did* allow it, JS here trying
// to reach into `frame.contentWindow` for a different-origin page is
// blocked by the browser's same-origin policy. Both together meant
// external navigation basically never worked.
//
// The fix: there is no more iframe. Every navigation (internal
// beast:// page or real external site) is a REAL top-level
// w.Navigate() call, exactly like a normal browser tab. This script
// is registered with w.Init() in main.go, so it re-runs and re-mounts
// the toolbar on top of whatever page has just loaded — internal or
// external, it doesn't matter, since it's injected directly into that
// page's own document rather than a sandboxed frame.
// ---------------------------------------------------

const chromeInjectionJS = `
(function () {
  if (window.__beastChromeBooting) return;
  window.__beastChromeBooting = true;

  function whenBodyReady(cb) {
    if (document.body) { cb(); return; }
    document.addEventListener('DOMContentLoaded', cb, { once: true });
    var t = setInterval(function () {
      if (document.body) { clearInterval(t); cb(); }
    }, 10);
  }

  function mountChrome() {
    var style = document.createElement('style');
    style.textContent =
      '#__beast_chrome, #__beast_chrome * { box-sizing: border-box; }' +
      'html { margin-top: 86px !important; }' +
      '#__beast_chrome {' +
      '  position: fixed; top: 0; left: 0; right: 0; z-index: 2147483647;' +
      '  font-family: "Segoe UI", sans-serif; background: #0b0b0d;' +
      '}' +
      '#tabstrip {' +
      '  display: flex; align-items: flex-end; background: #0f0f11;' +
      '  padding: 6px 6px 0 6px; gap: 4px; height: 38px; overflow-x: auto;' +
      '}' +
      '#tabstrip::-webkit-scrollbar { height: 3px; }' +
      '#tabstrip::-webkit-scrollbar-thumb { background: #2a2a2c; }' +
      '.tab-chip {' +
      '  display: flex; align-items: center; gap: 8px; background: #1a1b1e;' +
      '  color: #999; padding: 7px 10px 7px 14px; border-radius: 10px 10px 0 0;' +
      '  font-size: 12px; max-width: 180px; min-width: 120px; cursor: pointer;' +
      '  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;' +
      '  transition: background 0.15s, color 0.15s; position: relative; top: 1px;' +
      '}' +
      '.tab-chip.active { background: #1e1f22; color: #fff; box-shadow: 0 -3px 10px rgba(0,0,0,0.35); }' +
      '.tab-chip .tab-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }' +
      '.tab-chip .tab-close { color: #666; font-size: 13px; line-height: 1; padding: 2px 5px; border-radius: 4px; }' +
      '.tab-chip .tab-close:hover { background: #333; color: #ff6b6b; }' +
      '.tab-chip.incognito { background: #1a1420; border: 1px solid #3a2a4a; }' +
      '#new-tab-btn {' +
      '  width: 28px; height: 28px; border-radius: 8px; background: #1a1b1e; color: #888;' +
      '  display: flex; align-items: center; justify-content: center; cursor: pointer;' +
      '  font-size: 16px; flex-shrink: 0; margin-bottom: 2px;' +
      '}' +
      '#new-tab-btn:hover { background: #232427; color: #fff; }' +
      '#toolbar {' +
      '  display: flex; align-items: center; background: #1e1f22; padding: 8px 12px;' +
      '  gap: 8px; height: 48px; box-shadow: 0 4px 12px rgba(0,0,0,0.3); position: relative; z-index: 10;' +
      '}' +
      '.nav-btn {' +
      '  width: 32px; height: 32px; border-radius: 8px; background: transparent; color: #aaa;' +
      '  display: flex; align-items: center; justify-content: center; cursor: pointer;' +
      '  font-size: 15px; flex-shrink: 0; transition: 0.15s;' +
      '}' +
      '.nav-btn:hover { background: #2a2b2f; color: #fff; }' +
      '.nav-btn.disabled { opacity: 0.3; pointer-events: none; }' +
      '#address-wrap {' +
      '  flex: 1; display: flex; align-items: center; background: #17181a; border: 1px solid #2a2a2c;' +
      '  border-radius: 20px; padding: 0 8px 0 16px; height: 34px; gap: 8px; transition: border-color 0.15s;' +
      '}' +
      '#address-wrap:focus-within { border-color: #4d90fe; box-shadow: 0 0 0 3px rgba(77,144,254,0.15); }' +
      '#shield-icon { font-size: 12px; color: #3ddc84; flex-shrink: 0; cursor: pointer; }' +
      '#address-bar { flex: 1; background: transparent; border: none; outline: none; color: #eee; font-size: 13px; }' +
      '#address-bar::placeholder { color: #666; }' +
      '#star-btn { width: 26px; height: 26px; border-radius: 6px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: #666; font-size: 15px; flex-shrink: 0; }' +
      '#star-btn:hover { background: #232427; color: #ffd93d; }' +
      '#star-btn.saved { color: #ffd93d; }' +
      '#menu-btn { width: 32px; height: 32px; border-radius: 8px; background: transparent; color: #aaa; display: flex; align-items: center; justify-content: center; cursor: pointer; font-size: 16px; flex-shrink: 0; position: relative; }' +
      '#menu-btn:hover { background: #2a2b2f; color: #fff; }' +
      '#menu-dropdown {' +
      '  position: absolute; top: 46px; right: 8px; background: #1c1d20; border: 1px solid #2a2a2c;' +
      '  border-radius: 12px; box-shadow: 0 10px 30px rgba(0,0,0,0.5); width: 220px; padding: 8px; display: none; z-index: 100;' +
      '}' +
      '#menu-dropdown.open { display: block; }' +
      '.menu-item { padding: 10px 14px; border-radius: 8px; color: #ccc; font-size: 13px; cursor: pointer; display: flex; justify-content: space-between; }' +
      '.menu-item:hover { background: #2a2b2f; color: #fff; }' +
      '.menu-item .shortcut-hint { color: #555; font-size: 11px; }' +
      '.menu-divider { height: 1px; background: #2a2a2c; margin: 6px 4px; }' +
      '#incognito-badge {' +
      '  position: fixed; bottom: 14px; right: 14px; background: #1a1420; border: 1px solid #3a2a4a; color: #b58fd8;' +
      '  padding: 8px 14px; border-radius: 20px; font-size: 11px; box-shadow: 0 6px 16px rgba(0,0,0,0.4); display: none; z-index: 2147483647;' +
      '}' +
      '#incognito-badge.show { display: block; }' +
      '#suggest-box {' +
      '  position: absolute; top: 46px; left: 60px; right: 60px; background: #1c1d20; border: 1px solid #2a2a2c;' +
      '  border-radius: 12px; box-shadow: 0 12px 32px rgba(0,0,0,0.5); padding: 6px; display: none; z-index: 90;' +
      '}' +
      '#suggest-box.open { display: block; }' +
      '.suggest-item { padding: 9px 14px; border-radius: 8px; color: #ccc; font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 10px; }' +
      '.suggest-item:hover, .suggest-item.hovered { background: #2a2b2f; color: #fff; }' +
      '.suggest-icon { font-size: 12px; color: #666; width: 16px; }' +
      '#find-bar {' +
      '  position: fixed; top: 92px; right: 20px; background: #1c1d20; border: 1px solid #2a2a2c; border-radius: 10px;' +
      '  box-shadow: 0 8px 24px rgba(0,0,0,0.5); padding: 8px 10px; display: none; align-items: center; gap: 8px; z-index: 2147483647;' +
      '}' +
      '#find-bar.show { display: flex; }' +
      '#find-input { background: #17181a; border: 1px solid #2a2a2c; border-radius: 6px; padding: 6px 10px; color: #eee; font-size: 12px; outline: none; width: 160px; }' +
      '#find-count { font-size: 11px; color: #888; min-width: 50px; }' +
      '.find-btn { width: 24px; height: 24px; border-radius: 6px; display: flex; align-items: center; justify-content: center; color: #aaa; cursor: pointer; font-size: 12px; }' +
      '.find-btn:hover { background: #2a2b2f; color: #fff; }' +
      '#notif-badge { position: absolute; top: 3px; right: 3px; width: 8px; height: 8px; border-radius: 50%; background: #ff5c5c; display: none; }' +
      '#notif-badge.show { display: block; }' +
      '#notif-panel {' +
      '  position: absolute; top: 46px; right: 8px; width: 280px; max-height: 360px; overflow-y: auto; background: #1c1d20;' +
      '  border: 1px solid #2a2a2c; border-radius: 12px; box-shadow: 0 12px 32px rgba(0,0,0,0.5); padding: 8px; display: none; z-index: 100;' +
      '}' +
      '#notif-panel.open { display: block; }' +
      '.notif-item { padding: 10px 12px; border-radius: 8px; margin-bottom: 4px; background: #17181a; }' +
      '.notif-title { font-size: 12px; color: #fff; margin-bottom: 3px; }' +
      '.notif-msg { font-size: 11px; color: #999; }' +
      '.tab-chip.pinned { min-width: 42px; max-width: 42px; padding: 7px; }' +
      '.tab-chip.pinned .tab-title { display: none; }' +
      '.pin-icon { font-size: 11px; margin-right: 4px; }';
    document.head.appendChild(style);

    var root = document.createElement('div');
    root.id = '__beast_chrome';
    root.innerHTML =
      '<div id="tabstrip"></div>' +
      '<div id="toolbar">' +
      '  <div class="nav-btn" id="btn-back">&#8592;</div>' +
      '  <div class="nav-btn" id="btn-forward">&#8594;</div>' +
      '  <div class="nav-btn" id="btn-reload">&#8635;</div>' +
      '  <div class="nav-btn" id="btn-home">&#8962;</div>' +
      '  <div id="address-wrap">' +
      '    <span id="shield-icon">&#128737;</span>' +
      '    <input id="address-bar" placeholder="Search Google or type a URL">' +
      '    <div id="star-btn">&#9733;</div>' +
      '  </div>' +
      '  <div id="suggest-box"></div>' +
      '  <div class="nav-btn" id="btn-reader">&#128214;</div>' +
      '  <div class="nav-btn" id="btn-pip">&#128250;</div>' +
      '  <div class="nav-btn" id="btn-screenshot">&#128248;</div>' +
      '  <div class="nav-btn" id="btn-notif" style="position:relative;">' +
      '    &#128276;<div id="notif-badge"></div><div id="notif-panel"></div>' +
      '  </div>' +
      '  <div id="menu-btn">' +
      '    &#8942;' +
      '    <div id="menu-dropdown">' +
      '      <div class="menu-item" data-go="beast://bookmarks"><span>Bookmarks</span><span class="shortcut-hint">Ctrl+Shift+O</span></div>' +
      '      <div class="menu-item" data-go="beast://history"><span>History</span><span class="shortcut-hint">Ctrl+H</span></div>' +
      '      <div class="menu-item" data-go="beast://downloads"><span>Downloads</span><span class="shortcut-hint">Ctrl+J</span></div>' +
      '      <div class="menu-divider"></div>' +
      '      <div class="menu-item" id="menu-incognito"><span>New Incognito Tab</span><span class="shortcut-hint">Ctrl+Shift+N</span></div>' +
      '      <div class="menu-item" data-go="beast://shortcuts"><span>Keyboard Shortcuts</span></div>' +
      '      <div class="menu-item" data-go="beast://site-settings"><span>Site Settings</span></div>' +
      '      <div class="menu-item" data-go="beast://about"><span>About BEAST</span></div>' +
      '      <div class="menu-item" data-go="beast://cookies"><span>Cookies</span></div>' +
      '      <div class="menu-item" data-go="beast://autofill"><span>Autofill</span></div>' +
      '      <div class="menu-item" data-go="beast://feedback"><span>Send Feedback</span></div>' +
      '      <div class="menu-item" data-go="beast://passwords"><span>Passwords</span></div>' +
      '      <div class="menu-item" data-go="beast://updates"><span>About &amp; Updates</span></div>' +
      '      <div class="menu-item" data-go="beast://backup"><span>Backup &amp; Restore</span></div>' +
      '      <div class="menu-item" id="menu-print"><span>Print</span><span class="shortcut-hint">Ctrl+P</span></div>' +
      '      <div class="menu-divider"></div>' +
      '      <div class="menu-item" data-go="beast://settings"><span>Settings</span><span class="shortcut-hint">Ctrl+,</span></div>' +
      '    </div>' +
      '  </div>' +
      '</div>' +
      '<div id="find-bar">' +
      '  <input id="find-input" placeholder="Find in page">' +
      '  <span id="find-count">0/0</span>' +
      '  <div class="find-btn" id="find-prev">&#8593;</div>' +
      '  <div class="find-btn" id="find-next">&#8595;</div>' +
      '  <div class="find-btn" id="find-close">&times;</div>' +
      '</div>' +
      '<div id="incognito-badge">Incognito Mode Active</div>';
    document.documentElement.insertBefore(root, document.documentElement.firstChild);
    // Body content should render below the fixed chrome; putting the
    // root at documentElement level (instead of inside body) keeps it
    // immune to whatever the loaded page does to its own <body>.

    var activeTabId = null;
    var findDebounce = null;

    window.go = async function (input) {
      if (!input || input.trim() === '') return;
      await window.realNavigate(input);
    };

    window.goBack = async function () { await window.goBackNav(); };
    window.goForward = async function () { await window.goForwardNav(); };
    window.reload = function () { location.reload(); };
    window.goHome = function () { window.go('beast://home'); };

    async function navigateFromBar() {
      var val = document.getElementById('address-bar').value;
      if (val.trim() === '') return;
      await window.go(val);
    }

    async function createNewTab() {
      await window.openNewTab();
    }

    async function switchToTab(id) {
      await window.switchTab(id);
    }

    async function closeThisTab(id) {
      await window.closeTab(id);
    }

    async function renderTabs() {
      var tabs = await window.getAllTabsUI();
      var strip = document.getElementById('tabstrip');
      strip.innerHTML = '';

      tabs.forEach(function (t) {
        var chip = document.createElement('div');
        chip.className = 'tab-chip' + (t.ID === activeTabId ? ' active' : '');
        chip.onclick = function () { switchToTab(t.ID); };

        var title = document.createElement('div');
        title.className = 'tab-title';
        title.innerText = t.Title || 'New Tab';
        chip.appendChild(title);

        var closeBtn = document.createElement('div');
        closeBtn.className = 'tab-close';
        closeBtn.innerHTML = '&times;';
        closeBtn.onclick = function (e) { e.stopPropagation(); closeThisTab(t.ID); };
        chip.appendChild(closeBtn);

        strip.appendChild(chip);
      });

      var newBtn = document.createElement('div');
      newBtn.id = 'new-tab-btn';
      newBtn.innerHTML = '+';
      newBtn.onclick = createNewTab;
      strip.appendChild(newBtn);
    }

    function updateNavButtons(state) {
      document.getElementById('btn-back').className = state.canBack ? 'nav-btn' : 'nav-btn disabled';
      document.getElementById('btn-forward').className = state.canForward ? 'nav-btn' : 'nav-btn disabled';
    }

    async function refreshStar(url) {
      var starBtn = document.getElementById('star-btn');
      if (!url || url.indexOf('beast://') === 0) {
        starBtn.className = '';
        return;
      }
      var saved = await window.isBookmarked(url);
      starBtn.className = saved ? 'saved' : '';
    }

    async function toggleStar() {
      var url = document.getElementById('address-bar').value;
      if (!url || url.indexOf('beast://') === 0) return;
      var nowSaved = await window.toggleBookmark(url, url);
      document.getElementById('star-btn').className = nowSaved ? 'saved' : '';
    }

    async function updateTabTitleFromURL(url) {
      if (!url || activeTabId === null) return;
      var short = url.replace('https://', '').replace('http://', '').split('/')[0];
      if (url.indexOf('beast://') === 0) short = url.replace('beast://', 'BEAST: ');
      if (url.indexOf('data:text/html') === 0) short = 'BEAST';
      await window.updateTabTitle(activeTabId, short, url);
      renderTabs();
    }

    async function refreshChrome() {
      var state = await window.getNavState();
      activeTabId = state.tabId;
      updateNavButtons(state);
      await renderTabs();
      var addressBar = document.getElementById('address-bar');
      if (document.activeElement !== addressBar) {
        addressBar.value = state.url || '';
      }
      await refreshStar(state.url);
      await updateTabTitleFromURL(state.url);

      var incognitoOn = await window.isIncognitoActive();
      document.getElementById('incognito-badge').className = incognitoOn ? 'show' : '';

      refreshNotifBadge();
      window.injectContextMenu();
      window.injectFindInPage();
    }

    async function refreshNotifBadge() {
      var count = await window.getUnreadNotifCount();
      document.getElementById('notif-badge').className = count > 0 ? 'show' : '';
    }

    async function toggleNotifPanel(e) {
      e.stopPropagation();
      var panel = document.getElementById('notif-panel');
      if (panel.classList.contains('open')) {
        panel.classList.remove('open');
        return;
      }
      var items = await window.getNotifications();
      panel.innerHTML = items.length === 0
        ? '<div style="padding:16px;color:#666;font-size:12px;text-align:center;">No notifications</div>'
        : items.map(function (n) {
            return '<div class="notif-item"><div class="notif-title">' + n.Title +
              '</div><div class="notif-msg">' + n.Message + '</div></div>';
          }).join('');
      panel.classList.add('open');
      await window.markAllNotifsRead();
      refreshNotifBadge();
    }

    var suggestSelected = -1;
    var currentSuggestions = [];

    async function onAddressInput() {
      var val = document.getElementById('address-bar').value;
      var box = document.getElementById('suggest-box');
      if (val.trim() === '') {
        box.classList.remove('open');
        return;
      }
      currentSuggestions = await window.getSuggestions(val);
      suggestSelected = -1;
      renderSuggestions();
    }

    function renderSuggestions() {
      var box = document.getElementById('suggest-box');
      if (currentSuggestions.length === 0) {
        box.classList.remove('open');
        return;
      }
      box.innerHTML = currentSuggestions.map(function (s, i) {
        var icon = s.Source === 'bookmark' ? '&#9733;' : s.Source === 'history' ? '&#8635;' : '&#128269;';
        return '<div class="suggest-item' + (i === suggestSelected ? ' hovered' : '') + '" data-idx="' + i + '">' +
          '<span class="suggest-icon">' + icon + '</span><span>' + s.Text + '</span></div>';
      }).join('');
      Array.prototype.forEach.call(box.querySelectorAll('.suggest-item'), function (el) {
        el.onclick = function () { selectSuggestion(parseInt(el.getAttribute('data-idx'), 10)); };
      });
      box.classList.add('open');
    }

    function selectSuggestion(i) {
      var s = currentSuggestions[i];
      document.getElementById('address-bar').value = s.URL;
      document.getElementById('suggest-box').classList.remove('open');
      window.go(s.URL);
    }

    function closeSuggestions() {
      document.getElementById('suggest-box').classList.remove('open');
    }

    async function showSiteInfo(e) {
      e.stopPropagation();
      var url = document.getElementById('address-bar').value;
      if (!url || url.indexOf('beast://') === 0) return;
      var info = await window.getSiteInfo(url);
      alert(
        'Domain: ' + info.domain + '\\n' +
        'Secure (HTTPS): ' + (info.isSecure ? 'Yes' : 'No') + '\\n' +
        'Ads Blocked: ' + (info.adsBlocked ? 'On' : 'Off') + '\\n' +
        'Trackers Blocked: ' + (info.trackersBlocked ? 'On' : 'Off') + '\\n' +
        'Total Threats Blocked This Session: ' + info.threatsBlockedTotal
      );
    }

    function onFindInput() {
      clearTimeout(findDebounce);
      findDebounce = setTimeout(function () {
        var query = document.getElementById('find-input').value;
        try {
          var count = window.__beastFind.search(query);
          document.getElementById('find-count').innerText = count > 0 ? '1/' + count : '0/0';
        } catch (e) {}
      }, 200);
    }

    function findNext() {
      try {
        var idx = window.__beastFind.next();
        var total = window.__beastFind.matches.length;
        document.getElementById('find-count').innerText = (idx + 1) + '/' + total;
      } catch (e) {}
    }

    function findPrev() {
      try {
        var idx = window.__beastFind.prev();
        var total = window.__beastFind.matches.length;
        document.getElementById('find-count').innerText = (idx + 1) + '/' + total;
      } catch (e) {}
    }

    function openFindBar() {
      document.getElementById('find-bar').classList.add('show');
      document.getElementById('find-input').focus();
    }

    function closeFindBar() {
      try { window.__beastFind.clear(); } catch (e) {}
      document.getElementById('find-bar').classList.remove('show');
      document.getElementById('find-input').value = '';
      document.getElementById('find-count').innerText = '0/0';
    }

    function printCurrentPage() {
      try { window.print(); } catch (e) { alert('Cannot print this page'); }
    }

    function toggleMenu(e) {
      e.stopPropagation();
      document.getElementById('menu-dropdown').classList.toggle('open');
    }

    async function toggleReader() {
      await window.toggleReaderMode(activeTabId);
    }

    async function requestPiPForTab() {
      await window.requestPiP(activeTabId);
    }

    async function openIncognito() {
      await window.enableIncognito();
      await createNewTab();
    }

    // ---- wire up events ----
    document.getElementById('btn-back').onclick = window.goBack;
    document.getElementById('btn-forward').onclick = window.goForward;
    document.getElementById('btn-reload').onclick = window.reload;
    document.getElementById('btn-home').onclick = window.goHome;
    document.getElementById('shield-icon').onclick = showSiteInfo;
    document.getElementById('star-btn').onclick = toggleStar;
    document.getElementById('btn-reader').onclick = toggleReader;
    document.getElementById('btn-pip').onclick = requestPiPForTab;
    document.getElementById('btn-screenshot').onclick = function () { window.captureScreenshot(); };
    document.getElementById('btn-notif').onclick = toggleNotifPanel;
    document.getElementById('menu-btn').onclick = toggleMenu;
    document.getElementById('menu-incognito').onclick = openIncognito;
    document.getElementById('menu-print').onclick = printCurrentPage;
    document.getElementById('find-prev').onclick = findPrev;
    document.getElementById('find-next').onclick = findNext;
    document.getElementById('find-close').onclick = closeFindBar;
    document.getElementById('find-input').oninput = onFindInput;
    document.getElementById('find-input').onkeydown = function (e) {
      if (e.key === 'Enter') { e.shiftKey ? findPrev() : findNext(); }
      if (e.key === 'Escape') { closeFindBar(); }
    };

    Array.prototype.forEach.call(document.querySelectorAll('[data-go]'), function (el) {
      el.onclick = function () { window.go(el.getAttribute('data-go')); };
    });

    var addressBar = document.getElementById('address-bar');
    addressBar.oninput = onAddressInput;
    addressBar.onfocus = onAddressInput;
    addressBar.onkeydown = function (e) {
      if (e.key === 'Enter') { navigateFromBar(); closeSuggestions(); }
    };

    document.addEventListener('click', function () {
      document.getElementById('menu-dropdown').classList.remove('open');
      document.getElementById('suggest-box').classList.remove('open');
      document.getElementById('notif-panel').classList.remove('open');
    });

    document.addEventListener('keydown', function (e) {
      if (e.ctrlKey && e.key === 't') { e.preventDefault(); createNewTab(); }
      if (e.ctrlKey && e.key === 'w') { e.preventDefault(); closeThisTab(activeTabId); }
      if (e.ctrlKey && e.key === 'l') { e.preventDefault(); addressBar.focus(); addressBar.select(); }
      if (e.ctrlKey && e.key === 'r') { e.preventDefault(); window.reload(); }
      if (e.ctrlKey && e.key === 'd') { e.preventDefault(); toggleStar(); }
      if (e.ctrlKey && e.key === 'h') { e.preventDefault(); window.go('beast://history'); }
      if (e.ctrlKey && e.key === 'j') { e.preventDefault(); window.go('beast://downloads'); }
      if (e.ctrlKey && e.key === ',') { e.preventDefault(); window.go('beast://settings'); }
      if (e.ctrlKey && e.key === 'f') { e.preventDefault(); openFindBar(); }
      if (e.ctrlKey && e.key === 'p') { e.preventDefault(); printCurrentPage(); }
    });

    // Periodically save open tab URLs for crash recovery.
    setInterval(async function () {
      var tabs = await window.getAllTabsUI();
      var urls = tabs.map(function (t) { return t.URL; }).filter(function (u) { return u; });
      await window.updateRecoverySnapshot(urls);
    }, 10000);

    refreshChrome();
  }

  whenBodyReady(mountChrome);
})();
`
