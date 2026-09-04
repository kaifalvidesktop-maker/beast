package main

// ---------------------------------------------------
// KEYBOARD SHORTCUTS REGISTRY
// Actual key-capture happens in JS (keydown listener);
// this is the canonical list used to generate the
// beast://shortcuts help page and keep bindings consistent.
// ---------------------------------------------------

type Shortcut struct {
	Keys        string
	Description string
	Category    string
}

var shortcutList = []Shortcut{
	{"Ctrl + T", "Open a new tab", "Tabs"},
	{"Ctrl + W", "Close current tab", "Tabs"},
	{"Ctrl + Shift + T", "Reopen last closed tab", "Tabs"},
	{"Ctrl + Tab", "Switch to next tab", "Tabs"},
	{"Ctrl + Shift + Tab", "Switch to previous tab", "Tabs"},
	{"Ctrl + L", "Focus address bar", "Navigation"},
	{"Ctrl + R", "Reload current page", "Navigation"},
	{"Alt + Left", "Go back", "Navigation"},
	{"Alt + Right", "Go forward", "Navigation"},
	{"Ctrl + D", "Bookmark current page", "Bookmarks"},
	{"Ctrl + Shift + O", "Open bookmarks page", "Bookmarks"},
	{"Ctrl + H", "Open history page", "History"},
	{"Ctrl + J", "Open downloads page", "Downloads"},
	{"Ctrl + Shift + N", "Open incognito window", "Privacy"},
	{"Ctrl + Shift + Delete", "Clear browsing history", "Privacy"},
	{"Ctrl + F", "Find in page", "Page"},
	{"Ctrl + Plus", "Zoom in", "Page"},
	{"Ctrl + Minus", "Zoom out", "Page"},
	{"Ctrl + 0", "Reset zoom", "Page"},
	{"F11", "Toggle fullscreen", "Window"},
	{"Ctrl + Shift + R", "Toggle reader mode", "Page"},
	{"Ctrl + ,", "Open settings", "Settings"},
}

func getAllShortcuts() []Shortcut {
	return shortcutList
}

const shortcutsPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>BEAST Shortcuts</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0;
    background: #0b0b0d;
    color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 24px; }
  .group { margin-bottom: 26px; }
  .group-title {
    font-size: 13px;
    color: #4d90fe;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 12px;
  }
  .sc-row {
    display: flex;
    justify-content: space-between;
    padding: 10px 16px;
    background: #17181a;
    border: 1px solid #202022;
    border-radius: 8px;
    margin-bottom: 6px;
    font-size: 13px;
  }
  .keys {
    background: #232323;
    padding: 3px 10px;
    border-radius: 6px;
    font-family: monospace;
    color: #ccc;
    box-shadow: 0 2px 4px rgba(0,0,0,0.3);
  }
</style>
</head>
<body>
  <h1>Keyboard Shortcuts</h1>
  <div id="scContainer"></div>
<script>
  const shortcuts = window.__SHORTCUTS__;
  const grouped = {};
  shortcuts.forEach(function(s) {
    if (!grouped[s.Category]) grouped[s.Category] = [];
    grouped[s.Category].push(s);
  });
  const container = document.getElementById('scContainer');
  Object.keys(grouped).forEach(function(cat) {
    let html = '<div class="group"><div class="group-title">' + cat + '</div>';
    grouped[cat].forEach(function(s) {
      html += '<div class="sc-row"><span>' + s.Description + '</span><span class="keys">' + s.Keys + '</span></div>';
    });
    html += '</div>';
    container.innerHTML += html;
  });
</script>
</body>
</html>
`