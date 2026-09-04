package main

// ---------------------------------------------------
// BEAST BROWSER SHELL
// Persistent Toolbar + Tab Strip + iframe Content Area
// This is what actually loads first — everything else
// (home, settings, bookmarks...) loads INSIDE the iframe.
// ---------------------------------------------------

const shellHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>BEAST</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }

  html, body {
    height: 100%;
    background: #0b0b0d;
    font-family: 'Segoe UI', sans-serif;
    overflow: hidden;
  }

  #shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  /* ---------------- TAB STRIP ---------------- */

  #tabstrip {
    display: flex;
    align-items: flex-end;
    background: #0f0f11;
    padding: 6px 6px 0 6px;
    gap: 4px;
    height: 38px;
    flex-shrink: 0;
    overflow-x: auto;
  }

  #tabstrip::-webkit-scrollbar { height: 3px; }
  #tabstrip::-webkit-scrollbar-thumb { background: #2a2a2c; }

  .tab-chip {
    display: flex;
    align-items: center;
    gap: 8px;
    background: #1a1b1e;
    color: #999;
    padding: 7px 10px 7px 14px;
    border-radius: 10px 10px 0 0;
    font-size: 12px;
    max-width: 180px;
    min-width: 120px;
    cursor: pointer;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: background 0.15s, color 0.15s;
    position: relative;
    top: 1px;
  }

  .tab-chip.active {
    background: #1e1f22;
    color: #fff;
    box-shadow: 0 -3px 10px rgba(0,0,0,0.35);
  }

  .tab-chip .tab-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .tab-chip .tab-close {
    color: #666;
    font-size: 13px;
    line-height: 1;
    padding: 2px 5px;
    border-radius: 4px;
  }

  .tab-chip .tab-close:hover {
    background: #333;
    color: #ff6b6b;
  }

  .tab-chip.incognito {
    background: #1a1420;
    border: 1px solid #3a2a4a;
  }

  #new-tab-btn {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    background: #1a1b1e;
    color: #888;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 16px;
    flex-shrink: 0;
    margin-bottom: 2px;
  }

  #new-tab-btn:hover {
    background: #232427;
    color: #fff;
  }

  /* ---------------- TOOLBAR ---------------- */

  #toolbar {
    display: flex;
    align-items: center;
    background: #1e1f22;
    padding: 8px 12px;
    gap: 8px;
    height: 48px;
    flex-shrink: 0;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
    z-index: 10;
  }

  .nav-btn {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: transparent;
    color: #aaa;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 15px;
    flex-shrink: 0;
    transition: 0.15s;
  }

  .nav-btn:hover {
    background: #2a2b2f;
    color: #fff;
  }

  .nav-btn.disabled {
    opacity: 0.3;
    pointer-events: none;
  }

  #address-wrap {
    flex: 1;
    display: flex;
    align-items: center;
    background: #17181a;
    border: 1px solid #2a2a2c;
    border-radius: 20px;
    padding: 0 8px 0 16px;
    height: 34px;
    gap: 8px;
    transition: border-color 0.15s;
  }

  #address-wrap:focus-within {
    border-color: #4d90fe;
    box-shadow: 0 0 0 3px rgba(77,144,254,0.15);
  }

  #shield-icon {
    font-size: 12px;
    color: #3ddc84;
    flex-shrink: 0;
  }

  #address-bar {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: #eee;
    font-size: 13px;
  }

  #address-bar::placeholder { color: #666; }

  #star-btn {
    width: 26px;
    height: 26px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: #666;
    font-size: 15px;
    flex-shrink: 0;
  }

  #star-btn:hover { background: #232427; color: #ffd93d; }
  #star-btn.saved { color: #ffd93d; }

  #menu-btn {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: transparent;
    color: #aaa;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 16px;
    flex-shrink: 0;
    position: relative;
  }

  #menu-btn:hover { background: #2a2b2f; color: #fff; }

  #menu-dropdown {
    position: absolute;
    top: 46px;
    right: 8px;
    background: #1c1d20;
    border: 1px solid #2a2a2c;
    border-radius: 12px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.5);
    width: 220px;
    padding: 8px;
    display: none;
    z-index: 100;
  }

  #menu-dropdown.open { display: block; }

  .menu-item {
    padding: 10px 14px;
    border-radius: 8px;
    color: #ccc;
    font-size: 13px;
    cursor: pointer;
    display: flex;
    justify-content: space-between;
  }

  .menu-item:hover { background: #2a2b2f; color: #fff; }
  .menu-item .shortcut-hint { color: #555; font-size: 11px; }
  .menu-divider { height: 1px; background: #2a2a2c; margin: 6px 4px; }

  /* ---------------- CONTENT ---------------- */

  #content {
    flex: 1;
    position: relative;
    background: #0b0b0d;
  }

  #webframe {
    width: 100%;
    height: 100%;
    border: none;
    background: #0b0b0d;
  }

  #blocked-banner {
    position: absolute;
    top: 10px;
    left: 50%;
    transform: translateX(-50%);
    background: #2a1414;
    border: 1px solid #4a1f1f;
    color: #ff8a8a;
    padding: 10px 20px;
    border-radius: 10px;
    font-size: 12px;
    box-shadow: 0 6px 20px rgba(0,0,0,0.4);
    display: none;
    z-index: 50;
  }

  #blocked-banner.show { display: block; }

  #incognito-badge {
    position: fixed;
    bottom: 14px;
    right: 14px;
    background: #1a1420;
    border: 1px solid #3a2a4a;
    color: #b58fd8;
    padding: 8px 14px;
    border-radius: 20px;
    font-size: 11px;
    box-shadow: 0 6px 16px rgba(0,0,0,0.4);
    display: none;
    z-index: 50;
  }

  #incognito-badge.show { display: block; }

  /* ---------------- AUTOCOMPLETE ---------------- */

  #suggest-box {
    position: absolute;
    top: 46px;
    left: 60px;
    right: 60px;
    background: #1c1d20;
    border: 1px solid #2a2a2c;
    border-radius: 12px;
    box-shadow: 0 12px 32px rgba(0,0,0,0.5);
    padding: 6px;
    display: none;
    z-index: 90;
  }

  #suggest-box.open { display: block; }

  .suggest-item {
    padding: 9px 14px;
    border-radius: 8px;
    color: #ccc;
    font-size: 13px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .suggest-item:hover, .suggest-item.hovered { background: #2a2b2f; color: #fff; }
  .suggest-icon { font-size: 12px; color: #666; width: 16px; }

  /* ---------------- FIND IN PAGE ---------------- */

  #find-bar {
    position: absolute;
    top: 8px;
    right: 20px;
    background: #1c1d20;
    border: 1px solid #2a2a2c;
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0,0,0,0.5);
    padding: 8px 10px;
    display: none;
    align-items: center;
    gap: 8px;
    z-index: 60;
  }

  #find-bar.show { display: flex; }

  #find-input {
    background: #17181a;
    border: 1px solid #2a2a2c;
    border-radius: 6px;
    padding: 6px 10px;
    color: #eee;
    font-size: 12px;
    outline: none;
    width: 160px;
  }

  #find-count { font-size: 11px; color: #888; min-width: 50px; }

  .find-btn {
    width: 24px;
    height: 24px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #aaa;
    cursor: pointer;
    font-size: 12px;
  }

  .find-btn:hover { background: #2a2b2f; color: #fff; }

  /* ---------------- NOTIFICATION BELL ---------------- */

  #notif-badge {
    position: absolute;
    top: 3px;
    right: 3px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #ff5c5c;
    display: none;
  }

  #notif-badge.show { display: block; }

  #notif-panel {
    position: absolute;
    top: 46px;
    right: 8px;
    width: 280px;
    max-height: 360px;
    overflow-y: auto;
    background: #1c1d20;
    border: 1px solid #2a2a2c;
    border-radius: 12px;
    box-shadow: 0 12px 32px rgba(0,0,0,0.5);
    padding: 8px;
    display: none;
    z-index: 100;
  }

  #notif-panel.open { display: block; }

  .notif-item {
    padding: 10px 12px;
    border-radius: 8px;
    margin-bottom: 4px;
    background: #17181a;
  }

  .notif-title { font-size: 12px; color: #fff; margin-bottom: 3px; }
  .notif-msg { font-size: 11px; color: #999; }

  /* ---------------- PIN ICON ---------------- */

  .tab-chip.pinned { min-width: 42px; max-width: 42px; padding: 7px; }
  .tab-chip.pinned .tab-title { display: none; }
  .pin-icon { font-size: 11px; margin-right: 4px; }
</style>
</head>
<body>

<div id="shell">

  <div id="tabstrip"></div>

  <div id="toolbar">
    <div class="nav-btn" id="btn-back" onclick="goBack()">&#8592;</div>
    <div class="nav-btn" id="btn-forward" onclick="goForward()">&#8594;</div>
    <div class="nav-btn" id="btn-reload" onclick="reload()">&#8635;</div>
    <div class="nav-btn" onclick="goHome()">&#8962;</div>

    <div id="address-wrap">
      <span id="shield-icon" style="cursor:pointer;" onclick="showSiteInfo(event)">&#128737;</span>
      <input id="address-bar" placeholder="Search Google or type a URL"
             oninput="onAddressInput()"
             onkeydown="if(event.key==='Enter'){ navigateFromBar(); closeSuggestions(); }"
             onfocus="onAddressInput()">
      <div id="star-btn" onclick="toggleStar()">&#9733;</div>
    </div>

    <div id="find-bar">
      <input id="find-input" placeholder="Find in page" oninput="onFindInput()"
             onkeydown="if(event.key==='Enter'){ event.shiftKey ? findPrev() : findNext() } if(event.key==='Escape'){ closeFindBar() }">
      <span id="find-count">0/0</span>
      <div class="find-btn" onclick="findPrev()">&#8593;</div>
      <div class="find-btn" onclick="findNext()">&#8595;</div>
      <div class="find-btn" onclick="closeFindBar()">&times;</div>
    </div>

    <div id="suggest-box"></div>

    <div class="nav-btn" onclick="toggleReader()">&#128214;</div>
    <div class="nav-btn" onclick="requestPiP()">&#128250;</div>
    <div class="nav-btn" onclick="window.captureScreenshot()">&#128248;</div>

    <div class="nav-btn" style="position:relative;" onclick="toggleNotifPanel(event)">
      &#128276;
      <div id="notif-badge"></div>
      <div id="notif-panel"></div>
    </div>

    <div id="menu-btn" onclick="toggleMenu(event)">
      &#8942;
      <div id="menu-dropdown">
        <div class="menu-item" onclick="go('beast://bookmarks')"><span>Bookmarks</span><span class="shortcut-hint">Ctrl+Shift+O</span></div>
        <div class="menu-item" onclick="go('beast://history')"><span>History</span><span class="shortcut-hint">Ctrl+H</span></div>
        <div class="menu-item" onclick="go('beast://downloads')"><span>Downloads</span><span class="shortcut-hint">Ctrl+J</span></div>
        <div class="menu-divider"></div>
        <div class="menu-item" onclick="openIncognito()"><span>New Incognito Tab</span><span class="shortcut-hint">Ctrl+Shift+N</span></div>
        <div class="menu-item" onclick="go('beast://shortcuts')"><span>Keyboard Shortcuts</span></div>
        <div class="menu-item" onclick="go('beast://tasks')"><span>Task Manager</span></div>
        <div class="menu-item" onclick="go('beast://site-settings')"><span>Site Settings</span></div>
        <div class="menu-item" onclick="go('beast://about')"><span>About BEAST</span></div>
        <div class="menu-item" onclick="go('beast://cookies')"><span>Cookies</span></div>
        <div class="menu-item" onclick="go('beast://autofill')"><span>Autofill</span></div>
        <div class="menu-item" onclick="go('beast://feedback')"><span>Send Feedback</span></div>
        <div class="menu-item" onclick="go('beast://passwords')"><span>Passwords</span></div>
        <div class="menu-item" onclick="go('beast://updates')"><span>About & Updates</span></div>
        <div class="menu-item" onclick="go('beast://backup')"><span>Backup & Restore</span></div>
        <div class="menu-item" onclick="printCurrentPage()"><span>Print</span><span class="shortcut-hint">Ctrl+P</span></div>
        <div class="menu-divider"></div>
        <div class="menu-item" onclick="go('beast://settings')"><span>Settings</span><span class="shortcut-hint">Ctrl+,</span></div>
      </div>
    </div>
  </div>

  <div id="content">
    <div id="blocked-banner">Blocked by Security Shield</div>
    <div id="incognito-badge">Incognito Mode Active</div>
    <iframe id="webframe" srcdoc=""></iframe>
  </div>

</div>

<script>
  let activeTabId = null;
  let history_stack = {}; // tabId -> {stack: [], pointer: int}

  let findDebounce = null;

  async function requestPiP() {
    await window.requestPiP(activeTabId);
  }

  // Periodically save open tab URLs for crash recovery
  setInterval(async function() {
    const tabs = await window.getAllTabsUI();
    const urls = tabs.map(function(t) { return t.URL; }).filter(function(u) { return u; });
    await window.updateRecoverySnapshot(urls);
  }, 10000);

  function onFindInput() {
    clearTimeout(findDebounce);
    findDebounce = setTimeout(async function() {
      const query = document.getElementById('find-input').value;
      const frame = document.getElementById('webframe');
      try {
        const count = frame.contentWindow.__beastFind.search(query);
        document.getElementById('find-count').innerText =
          count > 0 ? '1/' + count : '0/0';
      } catch (e) {}
    }, 200);
  }

  function findNext() {
    const frame = document.getElementById('webframe');
    try {
      const idx = frame.contentWindow.__beastFind.next();
      const total = frame.contentWindow.__beastFind.matches.length;
      document.getElementById('find-count').innerText = (idx + 1) + '/' + total;
    } catch (e) {}
  }

  function findPrev() {
    const frame = document.getElementById('webframe');
    try {
      const idx = frame.contentWindow.__beastFind.prev();
      const total = frame.contentWindow.__beastFind.matches.length;
      document.getElementById('find-count').innerText = (idx + 1) + '/' + total;
    } catch (e) {}
  }

  function openFindBar() {
    document.getElementById('find-bar').classList.add('show');
    document.getElementById('find-input').focus();
  }

  function closeFindBar() {
    const frame = document.getElementById('webframe');
    try { frame.contentWindow.__beastFind.clear(); } catch (e) {}
    document.getElementById('find-bar').classList.remove('show');
    document.getElementById('find-input').value = '';
    document.getElementById('find-count').innerText = '0/0';
  }

  function printCurrentPage() {
    const frame = document.getElementById('webframe');
    try { frame.contentWindow.print(); } catch (e) { alert('Cannot print this page'); }
  }

  async function showSiteInfo(e) {
    e.stopPropagation();
    const url = document.getElementById('address-bar').value;
    if (!url || url.indexOf('beast://') === 0) return;
    const info = await window.getSiteInfo(url);
    alert(
      'Domain: ' + info.domain + '\n' +
      'Secure (HTTPS): ' + (info.isSecure ? 'Yes' : 'No') + '\n' +
      'Ads Blocked: ' + (info.adsBlocked ? 'On' : 'Off') + '\n' +
      'Trackers Blocked: ' + (info.trackersBlocked ? 'On' : 'Off') + '\n' +
      'Total Threats Blocked This Session: ' + info.threatsBlockedTotal
    );
  }

  let suggestSelected = -1;
  let currentSuggestions = [];

  async function onAddressInput() {
    const val = document.getElementById('address-bar').value;
    const box = document.getElementById('suggest-box');

    if (val.trim() === '') {
      box.classList.remove('open');
      return;
    }

    currentSuggestions = await window.getSuggestions(val);
    suggestSelected = -1;
    renderSuggestions();
  }

  function renderSuggestions() {
    const box = document.getElementById('suggest-box');
    if (currentSuggestions.length === 0) {
      box.classList.remove('open');
      return;
    }
    box.innerHTML = currentSuggestions.map(function(s, i) {
      const icon = s.Source === 'bookmark' ? '&#9733;' : s.Source === 'history' ? '&#8635;' : '&#128269;';
      return '<div class="suggest-item' + (i === suggestSelected ? ' hovered' : '') + '" ' +
        'onclick="selectSuggestion(' + i + ')">' +
        '<span class="suggest-icon">' + icon + '</span>' +
        '<span>' + s.Text + '</span></div>';
    }).join('');
    box.classList.add('open');
  }

  function selectSuggestion(i) {
    const s = currentSuggestions[i];
    document.getElementById('address-bar').value = s.URL;
    document.getElementById('suggest-box').classList.remove('open');
    go(s.URL);
  }

  function closeSuggestions() {
    document.getElementById('suggest-box').classList.remove('open');
  }

  async function toggleNotifPanel(e) {
    e.stopPropagation();
    const panel = document.getElementById('notif-panel');
    if (panel.classList.contains('open')) {
      panel.classList.remove('open');
      return;
    }
    const items = await window.getNotifications();
    panel.innerHTML = items.length === 0
      ? '<div style="padding:16px;color:#666;font-size:12px;text-align:center;">No notifications</div>'
      : items.map(function(n) {
          return '<div class="notif-item"><div class="notif-title">' + n.Title +
            '</div><div class="notif-msg">' + n.Message + '</div></div>';
        }).join('');
    panel.classList.add('open');
    await window.markAllNotifsRead();
    refreshNotifBadge();
  }

  async function refreshNotifBadge() {
    const count = await window.getUnreadNotifCount();
    document.getElementById('notif-badge').className = count > 0 ? 'show' : '';
  }

  async function initShell() {
    const tabId = await window.openNewTab();
    activeTabId = tabId;
    history_stack[tabId] = { stack: [], pointer: -1 };
    await renderTabs();
    await go('beast://home');
  }

  async function renderTabs() {
    const tabs = await window.getAllTabsUI();
    const strip = document.getElementById('tabstrip');
    strip.innerHTML = '';

    tabs.forEach(function(t) {
      const chip = document.createElement('div');
      chip.className = 'tab-chip' + (t.ID === activeTabId ? ' active' : '');
      chip.onclick = function() { switchToTab(t.ID); };

      const title = document.createElement('div');
      title.className = 'tab-title';
      title.innerText = t.Title || 'New Tab';
      chip.appendChild(title);

      const closeBtn = document.createElement('div');
      closeBtn.className = 'tab-close';
      closeBtn.innerHTML = '&times;';
      closeBtn.onclick = function(e) { e.stopPropagation(); closeThisTab(t.ID); };
      chip.appendChild(closeBtn);

      strip.appendChild(chip);
    });

    const newBtn = document.createElement('div');
    newBtn.id = 'new-tab-btn';
    newBtn.innerHTML = '+';
    newBtn.onclick = createNewTab;
    strip.appendChild(newBtn);
  }

  async function createNewTab() {
    const tabId = await window.openNewTab();
    activeTabId = tabId;
    history_stack[tabId] = { stack: [], pointer: -1 };
    await renderTabs();
    await go('beast://home');
  }

  async function switchToTab(id) {
    activeTabId = id;
    await window.switchTab(id);
    await renderTabs();
    const hs = history_stack[id];
    if (hs && hs.pointer >= 0) {
      loadIntoFrame(hs.stack[hs.pointer], false);
    } else {
      go('beast://home');
    }
  }

  async function closeThisTab(id) {
    await window.closeTab(id);
    delete history_stack[id];
    const tabs = await window.getAllTabsUI();
    if (tabs.length === 0) {
      await createNewTab();
      return;
    }
    if (id === activeTabId) {
      activeTabId = tabs[tabs.length - 1].ID;
      await window.switchTab(activeTabId);
    }
    await renderTabs();
    const hs = history_stack[activeTabId];
    if (hs && hs.pointer >= 0) {
      loadIntoFrame(hs.stack[hs.pointer], false);
    }
  }

  async function go(input) {
    const result = await window.realNavigate(input);
    handleNavResult(result, input);
  }

  async function navigateFromBar() {
    const val = document.getElementById('address-bar').value;
    if (val.trim() === '') return;
    await go(val);
  }

  function handleNavResult(result, originalInput) {
    if (result.type === 'blocked') {
      showBlockedBanner();
      return;
    }

    pushHistory(result.url || originalInput);

    if (result.type === 'internal') {
      loadIntoFrame({ kind: 'html', value: result.html, url: result.url }, false);
    } else {
      loadIntoFrame({ kind: 'url', value: result.url, url: result.url }, false);
    }

    document.getElementById('address-bar').value =
      (result.url && result.url.indexOf('beast://') === 0) ? result.url : (result.url || originalInput);

    refreshStar(result.url);
    updateTabTitle(result.url);
  }

  function loadIntoFrame(entry) {
    const frame = document.getElementById('webframe');
    if (entry.kind === 'html') {
      frame.removeAttribute('src');
      frame.srcdoc = entry.value;
    } else {
      frame.removeAttribute('srcdoc');
      frame.src = entry.value;
    }
    frame.onload = function() {
      window.injectContextMenu();
      window.injectFindInPage();
      refreshNotifBadge();
    };
    updateNavButtons();
  }

  function pushHistory(url) {
    if (!url) return;
    const hs = history_stack[activeTabId];
    if (!hs) return;

    const kind = url.indexOf('beast://') === 0 ? 'html' : 'url';

    hs.stack = hs.stack.slice(0, hs.pointer + 1);
    hs.stack.push({ kind: kind === 'html' ? 'url' : 'url', value: url, url: url });
    hs.pointer++;
    updateNavButtons();
  }

  function updateNavButtons() {
    const hs = history_stack[activeTabId];
    if (!hs) return;
    document.getElementById('btn-back').className = hs.pointer > 0 ? 'nav-btn' : 'nav-btn disabled';
    document.getElementById('btn-forward').className =
      hs.pointer < hs.stack.length - 1 ? 'nav-btn' : 'nav-btn disabled';
  }

  async function goBack() {
    const hs = history_stack[activeTabId];
    if (!hs || hs.pointer <= 0) return;
    hs.pointer--;
    const entry = hs.stack[hs.pointer];
    await go(entry.url);
  }

  async function goForward() {
    const hs = history_stack[activeTabId];
    if (!hs || hs.pointer >= hs.stack.length - 1) return;
    hs.pointer++;
    const entry = hs.stack[hs.pointer];
    await go(entry.url);
  }

  function reload() {
    const frame = document.getElementById('webframe');
    if (frame.src) {
      frame.src = frame.src;
    } else {
      frame.srcdoc = frame.srcdoc;
    }
  }

  function goHome() {
    go('beast://home');
  }

  function showBlockedBanner() {
    const banner = document.getElementById('blocked-banner');
    banner.className = 'show';
    setTimeout(function() { banner.className = ''; }, 2500);
  }

  async function refreshStar(url) {
    if (!url || url.indexOf('beast://') === 0) {
      document.getElementById('star-btn').className = '';
      return;
    }
    const saved = await window.isBookmarked(url);
    document.getElementById('star-btn').className = saved ? 'saved' : '';
  }

  async function toggleStar() {
    const url = document.getElementById('address-bar').value;
    if (!url || url.indexOf('beast://') === 0) return;
    const nowSaved = await window.toggleBookmark(url, url);
    document.getElementById('star-btn').className = nowSaved ? 'saved' : '';
  }

  async function updateTabTitle(url) {
    if (!url) return;
    let short = url.replace('https://', '').replace('http://', '').split('/')[0];
    if (url.indexOf('beast://') === 0) short = url.replace('beast://', 'BEAST: ');
    await window.updateTabTitle(activeTabId, short, url);
    renderTabs();
  }

  function toggleMenu(e) {
    e.stopPropagation();
    document.getElementById('menu-dropdown').classList.toggle('open');
  }

  document.addEventListener('click', function() {
    document.getElementById('menu-dropdown').classList.remove('open');
    document.getElementById('suggest-box').classList.remove('open');
    document.getElementById('notif-panel').classList.remove('open');
  });

  async function toggleReader() {
    await window.toggleReaderMode(activeTabId);
  }

  async function openIncognito() {
    await window.enableIncognito();
    document.getElementById('incognito-badge').className = 'show';
    await createNewTab();
  }

  document.addEventListener('keydown', function(e) {
    if (e.ctrlKey && e.key === 't') { e.preventDefault(); createNewTab(); }
    if (e.ctrlKey && e.key === 'w') { e.preventDefault(); closeThisTab(activeTabId); }
    if (e.ctrlKey && e.key === 'l') { e.preventDefault(); document.getElementById('address-bar').focus(); document.getElementById('address-bar').select(); }
    if (e.ctrlKey && e.key === 'r') { e.preventDefault(); reload(); }
    if (e.ctrlKey && e.key === 'd') { e.preventDefault(); toggleStar(); }
    if (e.ctrlKey && e.key === 'h') { e.preventDefault(); go('beast://history'); }
    if (e.ctrlKey && e.key === 'j') { e.preventDefault(); go('beast://downloads'); }
    if (e.ctrlKey && e.key === ',') { e.preventDefault(); go('beast://settings'); }
    if (e.ctrlKey && e.key === 'f') { e.preventDefault(); openFindBar(); }
    if (e.ctrlKey && e.key === 'p') { e.preventDefault(); printCurrentPage(); }
    if (e.ctrlKey && e.shiftKey && e.key === 'A') { e.preventDefault(); alert('Tab search: type in address bar and results will include open tabs (full overlay UI coming next batch)'); }
  });

  initShell();
</script>

</body>
</html>
`