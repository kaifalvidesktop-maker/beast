package main

// ---------------------------------------------------
// FIRST-RUN WELCOME / ONBOARDING PAGE
// ---------------------------------------------------

const welcomePageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Welcome to BEAST</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; min-height: 100vh;
    background: linear-gradient(180deg, #0b0b0d 0%, #121214 100%);
    color: #e8e8e8; font-family: 'Segoe UI', sans-serif;
    display: flex; flex-direction: column; align-items: center; justify-content: center;
    text-align: center; padding: 40px;
  }
  .logo {
    font-size: 54px; font-weight: 700; letter-spacing: 3px; margin-bottom: 10px;
    background: linear-gradient(90deg, #ffffff, #9b9b9b);
    -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  }
  .welcome-sub { color: #888; font-size: 15px; margin-bottom: 40px; }
  .features {
    display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
    max-width: 560px; margin-bottom: 40px;
  }
  .feature-card {
    background: #17181a; border: 1px solid #232323; border-radius: 14px;
    padding: 20px; text-align: left; box-shadow: 0 6px 20px rgba(0,0,0,0.35);
  }
  .feature-icon { font-size: 22px; margin-bottom: 10px; }
  .feature-title { font-size: 13px; font-weight: 600; color: #fff; margin-bottom: 4px; }
  .feature-desc { font-size: 11px; color: #888; }
  .start-btn {
    background: linear-gradient(90deg, #4d90fe, #6bb0ff);
    color: #000; border: none; padding: 14px 36px; border-radius: 30px;
    font-size: 14px; font-weight: 600; cursor: pointer;
    box-shadow: 0 8px 24px rgba(77,144,254,0.35);
  }
  .start-btn:hover { transform: translateY(-1px); }
</style>
</head>
<body>
  <div class="logo">BEAST</div>
  <div class="welcome-sub">Fast. Private. Yours. No account needed.</div>

  <div class="features">
    <div class="feature-card">
      <div class="feature-icon">&#128737;</div>
      <div class="feature-title">Ad & Tracker Blocking</div>
      <div class="feature-desc">Built-in shield blocks known ads and trackers automatically.</div>
    </div>
    <div class="feature-card">
      <div class="feature-icon">&#128274;</div>
      <div class="feature-title">Nothing Saved to Disk</div>
      <div class="feature-desc">History lives only in RAM and disappears when you close BEAST.</div>
    </div>
    <div class="feature-card">
      <div class="feature-icon">&#9889;</div>
      <div class="feature-title">No Extensions, No Bloat</div>
      <div class="feature-desc">A lean browser built for speed, not for tracking you.</div>
    </div>
    <div class="feature-card">
      <div class="feature-icon">&#128100;</div>
      <div class="feature-title">No Account Required</div>
      <div class="feature-desc">Nothing to sign up for. Your data stays on your device.</div>
    </div>
  </div>

  <button class="start-btn" onclick="go('beast://home')">Start Browsing</button>
</body>
</html>
`