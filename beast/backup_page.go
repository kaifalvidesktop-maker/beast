package main

const backupPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>BEAST Backup & Restore</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 24px; max-width: 480px; }
  .card {
    background: #17181a; border: 1px solid #232323; border-radius: 14px;
    padding: 24px; max-width: 480px; margin-bottom: 20px; box-shadow: 0 6px 20px rgba(0,0,0,0.35);
  }
  .card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 6px; }
  .card-desc { font-size: 12px; color: #888; margin-bottom: 16px; }
  .btn {
    background: #4d90fe; color: #000; border: none; padding: 10px 20px;
    border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer;
  }
  .btn-secondary {
    background: #232323; color: #ccc; border: 1px solid #2c2c2e;
    padding: 10px 20px; border-radius: 8px; font-size: 13px; cursor: pointer;
  }
  .status { font-size: 12px; color: #3ddc84; margin-top: 12px; display: none; }
  .status.show { display: block; }
</style>
</head>
<body>
  <h1>Backup & Restore</h1>
  <div class="sub">Move your bookmarks and settings between computers without an account.</div>

  <div class="card">
    <div class="card-title">Create Backup</div>
    <div class="card-desc">Saves your bookmarks, settings, and shortcuts to a JSON file in your Downloads folder. Passwords are not included here for security — use the separate vault export.</div>
    <button class="btn" onclick="createBackup()">Save Backup File</button>
    <div class="status" id="backupStatus"></div>
  </div>

  <div class="card">
    <div class="card-title">Restore from Backup</div>
    <div class="card-desc">Select a previously saved BEAST backup file to restore your data.</div>
    <button class="btn-secondary" onclick="document.getElementById('restoreFile').click()">Choose Backup File</button>
    <input type="file" id="restoreFile" accept=".json" style="display:none;" onchange="restoreBackup(event)">
    <div class="status" id="restoreStatus"></div>
  </div>

<script>
  async function createBackup() {
    const path = await window.saveBackupToFile();
    const statusEl = document.getElementById('backupStatus');
    statusEl.innerText = 'Saved to: ' + path;
    statusEl.classList.add('show');
  }

  function restoreBackup(event) {
    const file = event.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = async function(e) {
      const count = await window.restoreBackup(e.target.result);
      const statusEl = document.getElementById('restoreStatus');
      statusEl.innerText = count + ' bookmark(s) restored. Settings applied.';
      statusEl.classList.add('show');
    };
    reader.readAsText(file);
  }
</script>
</body>
</html>
`