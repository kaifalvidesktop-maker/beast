package main

const beastVersion = "0.1.0-alpha"
const beastBuildDate = "2026"

const aboutPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>About BEAST</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; min-height: 100vh; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; display: flex; flex-direction: column;
    align-items: center; justify-content: center; text-align: center; padding: 40px;
  }
  .logo {
    font-size: 52px; font-weight: 700; letter-spacing: 3px; margin-bottom: 10px;
    background: linear-gradient(90deg, #ffffff, #9b9b9b);
    -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  }
  .version { color: #666; font-size: 13px; margin-bottom: 30px; }
  .info-card {
    background: #17181a; border: 1px solid #232323; border-radius: 14px;
    padding: 24px 32px; max-width: 480px; box-shadow: 0 8px 26px rgba(0,0,0,0.4);
  }
  .info-row { display: flex; justify-content: space-between; padding: 8px 0; font-size: 13px; border-bottom: 1px solid #1f1f21; }
  .info-row:last-child { border-bottom: none; }
  .info-label { color: #888; }
  .info-value { color: #ddd; }
  .made-by { margin-top: 24px; font-size: 12px; color: #555; }
  .made-by b { color: #999; }
</style>
</head>
<body>
  <div class="logo">BEAST</div>
  <div class="version">Version ` + beastVersion + ` &middot; Built ` + beastBuildDate + `</div>

  <div class="info-card">
    <div class="info-row"><span class="info-label">Engine</span><span class="info-value">System WebView (Chromium/WebKit-based)</span></div>
    <div class="info-row"><span class="info-label">Written in</span><span class="info-value">Go</span></div>
    <div class="info-row"><span class="info-label">License</span><span class="info-value">Open Source</span></div>
    <div class="info-row"><span class="info-label">Account required</span><span class="info-value">No</span></div>
    <div class="info-row"><span class="info-label">Sync</span><span class="info-value">None (fully local)</span></div>
    <div class="info-row"><span class="info-label">History storage</span><span class="info-value">RAM only</span></div>
  </div>

  <div class="made-by">Made by <b>kaif alvi</b>; built  as a learning project.</div>
</body>
</html>
`
