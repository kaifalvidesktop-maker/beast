package main

const passwordsPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>BEAST Passwords</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 24px; max-width: 480px; }

  .lock-screen {
    max-width: 360px; background: #17181a; border: 1px solid #232323;
    border-radius: 14px; padding: 30px; box-shadow: 0 8px 26px rgba(0,0,0,0.4);
  }
  .lock-icon { font-size: 36px; margin-bottom: 14px; }
  .lock-screen input {
    width: 100%; background: #1f2023; border: 1px solid #2c2c2e; border-radius: 8px;
    padding: 10px 14px; color: #fff; font-size: 13px; outline: none; margin-bottom: 12px;
  }
  .btn {
    background: #4d90fe; color: #000; border: none; padding: 10px 20px;
    border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; width: 100%;
  }
  .btn-secondary {
    background: #232323; color: #ccc; border: 1px solid #2c2c2e;
    padding: 9px 16px; border-radius: 8px; font-size: 12px; cursor: pointer;
  }

  .vault-view { display: none; }
  .vault-view.show { display: block; }
  .lock-screen.hide { display: none; }

  .toolbar { display: flex; gap: 10px; margin-bottom: 20px; }
  .pw-item {
    display: flex; justify-content: space-between; align-items: center;
    background: #17181a; border: 1px solid #232323; border-radius: 10px;
    padding: 14px 18px; margin-bottom: 8px;
  }
  .pw-domain { font-size: 13px; color: #fff; margin-bottom: 3px; }
  .pw-user { font-size: 11px; color: #888; }
  .pw-value { font-family: monospace; font-size: 12px; color: #4d90fe; margin-left: 10px; }
  .row-actions { display: flex; gap: 8px; }
  .mini-btn {
    background: #232323; color: #999; border: 1px solid #2c2c2e;
    padding: 5px 10px; border-radius: 6px; font-size: 11px; cursor: pointer;
  }
  .mini-btn:hover { color: #fff; }
  .empty { color: #666; font-size: 13px; margin-top: 40px; text-align: center; }

  .add-form {
    background: #17181a; border: 1px solid #232323; border-radius: 12px;
    padding: 18px; margin-bottom: 20px; display: none;
  }
  .add-form.show { display: block; }
  .add-form input {
    width: 100%; background: #1f2023; border: 1px solid #2c2c2e; border-radius: 8px;
    padding: 9px 12px; color: #fff; font-size: 12px; outline: none; margin-bottom: 10px;
  }
</style>
</head>
<body>
  <h1>Passwords</h1>
  <div class="sub">Encrypted with AES-256, stored only in memory. Locked automatically when BEAST closes.</div>

  <div class="lock-screen" id="lockScreen">
    <div class="lock-icon">&#128274;</div>
    <input type="password" id="masterInput" placeholder="Enter master password to unlock">
    <button class="btn" onclick="unlockVault()">Unlock Vault</button>
  </div>

  <div class="vault-view" id="vaultView">
    <div class="toolbar">
      <button class="btn-secondary" onclick="toggleAddForm()">+ Add Password</button>
      <button class="btn-secondary" onclick="lockVault()">Lock Vault</button>
    </div>

    <div class="add-form" id="addForm">
      <input id="new-domain" placeholder="Website (e.g. github.com)">
      <input id="new-username" placeholder="Username or email">
      <input id="new-password" type="text" placeholder="Password">
      <button class="btn-secondary" onclick="genPassword()">Generate Strong Password</button>
      <button class="btn" style="margin-top:10px;" onclick="savePassword()">Save</button>
    </div>

    <div id="pwList"></div>
  </div>

<script>
  async function unlockVault() {
    const pw = document.getElementById('masterInput').value;
    if (!pw) return;
    await window.unlockVault(pw);
    document.getElementById('lockScreen').classList.add('hide');
    document.getElementById('vaultView').classList.add('show');
    loadPasswords();
  }

  async function lockVault() {
    await window.lockVault();
    document.getElementById('lockScreen').classList.remove('hide');
    document.getElementById('vaultView').classList.remove('show');
    document.getElementById('masterInput').value = '';
  }

  function toggleAddForm() {
    document.getElementById('addForm').classList.toggle('show');
  }

  async function genPassword() {
    const pw = await window.generatePassword(18);
    document.getElementById('new-password').value = pw;
  }

  async function savePassword() {
    const domain = document.getElementById('new-domain').value;
    const username = document.getElementById('new-username').value;
    const password = document.getElementById('new-password').value;
    if (!domain || !password) return;

    await window.savePasswordEntry(domain, username, password);
    document.getElementById('new-domain').value = '';
    document.getElementById('new-username').value = '';
    document.getElementById('new-password').value = '';
    document.getElementById('addForm').classList.remove('show');
    loadPasswords();
  }

  async function loadPasswords() {
    const items = await window.getPasswordList();
    const container = document.getElementById('pwList');

    if (!items || items.length === 0) {
      container.innerHTML = '<div class="empty">No saved passwords yet.</div>';
      return;
    }

    container.innerHTML = items.map(function(p) {
      return '<div class="pw-item">' +
        '<div><div class="pw-domain">' + p.Domain + '</div><div class="pw-user">' + p.Username + '</div></div>' +
        '<div class="row-actions">' +
          '<span class="pw-value" id="pw-val-' + p.ID + '">••••••••</span>' +
          '<div class="mini-btn" onclick="revealPassword(' + p.ID + ')">Show</div>' +
          '<div class="mini-btn" onclick="copyPassword(' + p.ID + ')">Copy</div>' +
          '<div class="mini-btn" onclick="deletePassword(' + p.ID + ')">Delete</div>' +
        '</div>' +
      '</div>';
    }).join('');
  }

  async function revealPassword(id) {
    const pw = await window.revealPasswordEntry(id);
    document.getElementById('pw-val-' + id).innerText = pw;
  }

  async function copyPassword(id) {
    const pw = await window.revealPasswordEntry(id);
    navigator.clipboard.writeText(pw);
  }

  async function deletePassword(id) {
    if (!confirm('Delete this saved password?')) return;
    await window.deletePasswordEntry(id);
    loadPasswords();
  }
</script>
</body>
</html>
`