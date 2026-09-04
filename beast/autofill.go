package main

import (
	"strings"
	"sync"
)

// ---------------------------------------------------
// FORM AUTOFILL (name, email, address — NOT passwords)
// Stored in RAM only, cleared on restart like everything else
// ---------------------------------------------------

type AutofillProfile struct {
	FullName string
	Email    string
	Phone    string
	Address  string
	City     string
	ZipCode  string
	Country  string
}

type AutofillManager struct {
	mu       sync.Mutex
	Profiles []*AutofillProfile
	Enabled  bool
}

var autofillManager = &AutofillManager{
	Profiles: []*AutofillProfile{},
	Enabled:  true,
}

// SaveProfile adds or replaces the single active autofill profile
func (am *AutofillManager) SaveProfile(p AutofillProfile) *AutofillProfile {
	am.mu.Lock()
	defer am.mu.Unlock()

	profile := &p
	am.Profiles = []*AutofillProfile{profile} // single profile model, keeps it simple
	return profile
}

// GetProfile returns the current autofill profile (or nil if none set)
func (am *AutofillManager) GetProfile() *AutofillProfile {
	am.mu.Lock()
	defer am.mu.Unlock()

	if len(am.Profiles) == 0 {
		return nil
	}
	return am.Profiles[0]
}

// ClearProfile wipes all saved form data
func (am *AutofillManager) ClearProfile() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.Profiles = []*AutofillProfile{}
}

// ToggleEnabled turns autofill suggestion on/off globally
func (am *AutofillManager) ToggleEnabled() bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.Enabled = !am.Enabled
	return am.Enabled
}

// IsEnabled checks current state
func (am *AutofillManager) IsEnabled() bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	return am.Enabled
}

// Injectable JS that detects common form fields and offers to fill them
const autofillDetectJS = `
(function() {
	if (window.__beastAutofillHooked) return;
	window.__beastAutofillHooked = true;

	function looksLikeField(el, keywords) {
		var attrs = (el.name + ' ' + el.id + ' ' + (el.placeholder || '') + ' ' + (el.autocomplete || '')).toLowerCase();
		return keywords.some(function(k) { return attrs.indexOf(k) !== -1; });
	}

	document.addEventListener('focusin', async function(e) {
		var el = e.target;
		if (el.tagName !== 'INPUT') return;
		if (el.type === 'password') return; // never touch password fields

		var fieldMap = [
			{ keywords: ['name', 'fullname'], key: 'FullName' },
			{ keywords: ['email'], key: 'Email' },
			{ keywords: ['phone', 'tel'], key: 'Phone' },
			{ keywords: ['address'], key: 'Address' },
			{ keywords: ['city'], key: 'City' },
			{ keywords: ['zip', 'postal'], key: 'ZipCode' },
			{ keywords: ['country'], key: 'Country' }
		];

		for (var i = 0; i < fieldMap.length; i++) {
			if (looksLikeField(el, fieldMap[i].keywords)) {
				el.dataset.beastAutofillKey = fieldMap[i].key;
				el.style.outline = '2px solid rgba(77,144,254,0.4)';
				break;
			}
		}
	});

	document.addEventListener('focusout', function(e) {
		if (e.target.tagName === 'INPUT') {
			e.target.style.outline = '';
		}
	});

	document.addEventListener('dblclick', async function(e) {
		var el = e.target;
		if (el.tagName !== 'INPUT' || !el.dataset.beastAutofillKey) return;
		var profile = await window.getAutofillProfile();
		if (!profile) return;
		var value = profile[el.dataset.beastAutofillKey];
		if (value) {
			el.value = value;
			el.dispatchEvent(new Event('input', { bubbles: true }));
		}
	});
})();
`

const autofillPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>BEAST Autofill</title>
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
    padding: 24px; max-width: 420px; box-shadow: 0 6px 20px rgba(0,0,0,0.35);
  }
  .field { margin-bottom: 14px; }
  .field label { display: block; font-size: 11px; color: #888; margin-bottom: 6px; }
  .field input {
    width: 100%; background: #1f2023; border: 1px solid #2c2c2e; border-radius: 8px;
    padding: 9px 12px; color: #fff; font-size: 13px; outline: none;
  }
  .field input:focus { border-color: #4d90fe; }
  .btn-row { display: flex; gap: 10px; margin-top: 8px; }
  .btn {
    background: #232323; color: #ccc; border: 1px solid #2c2c2e;
    padding: 9px 18px; border-radius: 8px; font-size: 12px; cursor: pointer;
  }
  .btn:hover { background: #2a2a2c; color: #fff; }
  .btn.primary { background: #4d90fe; color: #000; border: none; font-weight: 600; }
  .btn.primary:hover { opacity: 0.9; }
  .hint { font-size: 11px; color: #666; margin-top: 16px; }
</style>
</head>
<body>
  <h1>Autofill</h1>
  <div class="sub">Saved locally in memory only, never written to disk. Double-click a detected form field on any page to fill it. Passwords are never stored here.</div>

  <div class="card">
    <div class="field"><label>Full Name</label><input id="af-name" placeholder="Kaif Ahmed"></div>
    <div class="field"><label>Email</label><input id="af-email" placeholder="you@example.com"></div>
    <div class="field"><label>Phone</label><input id="af-phone" placeholder="+880..."></div>
    <div class="field"><label>Address</label><input id="af-address" placeholder="Street address"></div>
    <div class="field"><label>City</label><input id="af-city" placeholder="Rangpur"></div>
    <div class="field"><label>Zip Code</label><input id="af-zip" placeholder="5400"></div>
    <div class="field"><label>Country</label><input id="af-country" placeholder="Bangladesh"></div>

    <div class="btn-row">
      <button class="btn primary" onclick="saveProfile()">Save</button>
      <button class="btn" onclick="clearProfile()">Clear</button>
    </div>
  </div>

  <div class="hint">Tip: this data disappears when BEAST closes, just like history.</div>

<script>
  async function loadProfile() {
    const p = await window.getAutofillProfile();
    if (!p) return;
    document.getElementById('af-name').value = p.FullName || '';
    document.getElementById('af-email').value = p.Email || '';
    document.getElementById('af-phone').value = p.Phone || '';
    document.getElementById('af-address').value = p.Address || '';
    document.getElementById('af-city').value = p.City || '';
    document.getElementById('af-zip').value = p.ZipCode || '';
    document.getElementById('af-country').value = p.Country || '';
  }

  async function saveProfile() {
    await window.saveAutofillProfile(
      document.getElementById('af-name').value,
      document.getElementById('af-email').value,
      document.getElementById('af-phone').value,
      document.getElementById('af-address').value,
      document.getElementById('af-city').value,
      document.getElementById('af-zip').value,
      document.getElementById('af-country').value
    );
    alert('Saved');
  }

  async function clearProfile() {
    if (!confirm('Clear all saved autofill data?')) return;
    await window.clearAutofillProfile();
    document.querySelectorAll('.card input').forEach(function(i) { i.value = ''; });
  }

  loadProfile();
</script>
</body>
</html>
`

func trimAll(s string) string {
	return strings.TrimSpace(s)
}