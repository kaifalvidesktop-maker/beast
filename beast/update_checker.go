package main

// ---------------------------------------------------
// UPDATE CHECKER (informational page — real update
// downloading isn't wired up yet since BEAST has no
// backend server; this documents the version/changelog)
// ---------------------------------------------------

type ChangelogEntry struct {
	Version string
	Date    string
	Changes []string
}

var beastChangelog = []ChangelogEntry{
	{
		Version: "0.1.0-alpha",
		Date:    "2026",
		Changes: []string{
			"Initial BEAST build: tabs, history, bookmarks, downloads",
			"Security Shield with ad/tracker blocking",
			"Incognito mode, reader mode, zoom controls",
			"Encrypted password manager (AES-256)",
			"Autofill, cookie tracking, force dark mode",
		},
	},
}

const updatesPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>BEAST Updates</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .current {
    background: #17181a; border: 1px solid #232323; border-radius: 12px;
    padding: 18px 22px; margin-bottom: 24px; max-width: 480px;
    display: flex; align-items: center; gap: 14px;
  }
  .current .dot { width: 10px; height: 10px; border-radius: 50%; background: #3ddc84; }
  .current-text { font-size: 13px; }
  .current-text b { color: #fff; }
  .changelog-entry {
    max-width: 480px; margin-bottom: 20px;
    background: #17181a; border: 1px solid #232323; border-radius: 12px; padding: 18px 22px;
  }
  .changelog-version { font-size: 14px; color: #fff; font-weight: 600; margin-bottom: 4px; }
  .changelog-date { font-size: 11px; color: #666; margin-bottom: 12px; }
  .changelog-item { font-size: 12px; color: #bbb; padding: 4px 0; padding-left: 16px; position: relative; }
  .changelog-item:before { content: "•"; position: absolute; left: 0; color: #4d90fe; }
</style>
</head>
<body>
  <h1>BEAST Updates</h1>
  <div class="current">
    <div class="dot"></div>
    <div class="current-text">You're running <b>BEAST ` + beastVersion + `</b> — the latest version.</div>
  </div>
  <div id="changelogList"></div>

<script>
  async function loadChangelog() {
    const items = await window.getChangelog();
    document.getElementById('changelogList').innerHTML = items.map(function(entry) {
      return '<div class="changelog-entry">' +
        '<div class="changelog-version">Version ' + entry.Version + '</div>' +
        '<div class="changelog-date">' + entry.Date + '</div>' +
        entry.Changes.map(function(c) { return '<div class="changelog-item">' + c + '</div>'; }).join('') +
      '</div>';
    }).join('');
  }
  loadChangelog();
</script>
</body>
</html>
`