free tier only 

This new page 
`/home/kushal/src/golang/GoTunnels/frontend/captcha.html`
does not quite match the look and feel of the rest of the application 
we will address the backend functionality of the page later but for now 
can you please bring this page in line with the rest of the application? 
keep all necessary javascript and css local, and remove all unnecessary javascript and css from the html 
and please give me FULL file for this `/home/kushal/src/golang/GoTunnels/frontend/captcha.html`
the header and footer and the whole page should look consistent with the rest of the site. 
We don't need a "theme changer" but we do need to make sure this page is legible and accessible 
the theme of the page should be managed by the site as a whole. 
we also need to remove any reference to the neocities urls as this site will be on trycloudflare dot com 
and needs to be independent 
we don't have the try cloudflare url when we write the code as the url is literally generated during build time 
also qa, integration, staging, production, all environments will have their own URL later 
so we either need a good solution for that or remove any such hard coding entirely 

 Conversation with Gemini

I have this activity.html

it has a problem in that it is not very mobile friendly

what is the best way with best in class engineering practices

to make the website responsive across different screen sizes

and device capabilities without adding a build step

or adding any vendor software or packages?

Please give me all the possible options

along with the proposed code implementations


```html

<!doctype html>

<html lang="en"><head>

<meta charset="utf-8">

<meta name="viewport" content="width=device-width, initial-scale=1">

<title>Activity · GoTunnels</title>

<link rel="stylesheet" href="/css/style.css">

</head>

<body>

<header class="topbar">

<div class="topbar-inner">

<span class="brand"><span class="dot"></span> GoTunnels</span>

<nav class="mainnav">

<a href="/">Home</a>

<a href="/activity" data-auth="in" class="active">Activity</a>

<a href="/settings" data-auth="in" class="">Settings</a>

<a href="/login" data-auth="out" class="hidden">Log in</a>

<a href="/signup" data-auth="out" class="hidden">Sign up</a>

</nav>

</div>

</header>


<main>

<h1>Your activity</h1>

<p class="lead">

Every sign-up and login is recorded. Your IP address is never stored —

only a salted SHA-256 hash of it, shown here exactly as stored.

</p>


<div id="msg" class="msg"></div>


<div class="card">

<table>

<thead>

<tr>

<th>When</th>

<th>Event</th>

<th>Method</th>

<th>Outcome</th>

<th>IP hash</th>

</tr>

</thead>

<tbody id="activity-body"><tr><td>7/4/2026, 8:31:51 AM</td><td>login</td><td>password</td><td><span class="tag success">success</span></td><td class="hash">1a96e1c8f68f0edbb602f8adc19b4b715e832504a17e2660a45d3ad0f87875cb</td></tr><tr><td>7/4/2026, 8:31:48 AM</td><td>logout</td><td>—</td><td><span class="tag success">success</span></td><td class="hash">1a96e1c8f68f0edbb602f8adc19b4b715e832504a17e2660a45d3ad0f87875cb</td></tr><tr><td>7/4/2026, 8:31:36 AM</td><td>signup</td><td>password</td><td><span class="tag success">success</span></td><td class="hash">1a96e1c8f68f0edbb602f8adc19b4b715e832504a17e2660a45d3ad0f87875cb</td></tr></tbody>

</table>

</div>

</main>


<footer class="foot">

GoTunnels · built with LLM assistance (see the README)

</footer>


<script type="module" src="/js/page-activity.js"></script>


</body></html>

```

Keeping the stack lightweight and adhering to a utility-first design philosophy means we can tackle this entirely with modern, vanilla CSS. The primary responsiveness bottlenecks on this page are the <table> element (specifically the 64-character IP hash column ) and the top navigation bar.  

Here are the best-in-class options for handling the layout across device capabilities, requiring zero build steps, zero vendor packages, and no changes to your underlying Go/JavaScript logic.
Option 1: The "Card View" Transformation (Recommended)

This is the standard engineering practice for complex data tables on mobile. Instead of forcing the user to scroll horizontally, a CSS media query re-flows each table row (<tr>) into a visually distinct "card," and stacks the cells (<td>) vertically.

Because your table rows are generated dynamically via page-activity.js, the most elegant way to implement this is using CSS pseudo-elements (::before) paired with :nth-of-type(). This means you do not need to touch your JavaScript or HTML.  

Implementation:
Add this to the bottom of your /css/style.css:
CSS

/* ---- Mobile Responsiveness ---- */
@media (max-width: 650px) {
  /* Fix the navigation bar */
  .topbar-inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  nav.mainnav {
    margin-left: 0;
    width: 100%;
    justify-content: space-between;
  }

  /* Transform the table into cards */
  table, thead, tbody, th, td, tr {
    display: block;
  }

  /* Visually hide the header, but keep it accessible for screen readers */
  thead tr {
    position: absolute;
    top: -9999px;
    left: -9999px;
  }

  /* Style the row as a card */
  tr {
    border-bottom: 2px solid var(--border-bright);
    margin-bottom: 16px;
    padding-bottom: 8px;
  }

  /* Style the cells as key-value pairs */
  td {
    border: none;
    border-bottom: 1px solid var(--border);
    position: relative;
    padding-left: 110px;
    text-align: right;
    min-height: 38px; /* Ensures consistent height */
  }

  td:last-child {
    border-bottom: 0;
  }

  /* Inject the column headers as labels */
  td::before {
    position: absolute;
    top: 10px;
    left: 10px;
    width: 90px;
    padding-right: 10px;
    white-space: nowrap;
    text-align: left;
    font-weight: 500;
    color: var(--text-faint);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  td:nth-of-type(1)::before { content: "When"; }
  td:nth-of-type(2)::before { content: "Event"; }
  td:nth-of-type(3)::before { content: "Method"; }
  td:nth-of-type(4)::before { content: "Outcome"; }
  td:nth-of-type(5)::before { content: "IP hash"; }

  /* Allow the hash to take up the remaining space naturally */
  td.hash {
    max-width: none; 
  }
}

Option 2: The Horizontal Scroll Wrapper (The Minimalist Approach)

If you prefer to maintain the strict tabular layout regardless of screen size, the alternative is to contain the overflow. This allows the user to swipe left and right on the table without breaking the page's overall layout.

While the UX isn't quite as seamless as Option 1, it is extremely lightweight and excellent for data-dense interfaces where columnar comparison is critical.

Implementation:
First, add a new utility class to your /css/style.css:
CSS

.table-responsive {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch; /* Smooth scrolling on iOS */
}

/* Include the navigation fix from Option 1 */
@media (max-width: 600px) {
  .topbar-inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  nav.mainnav {
    margin-left: 0;
    width: 100%;
    justify-content: space-between;
  }
}

Second, modify your activity.html by wrapping the <table> inside a <div> utilizing this new class:
HTML

<div class="card">
  <div class="table-responsive"> <table>
      <thead>
        <tr>
          <th>When</th>
          <th>Event</th>
          <th>Method</th>
          <th>Outcome</th>
          <th>IP hash</th>
        </tr>
      </thead>
      <tbody id="activity-body">
        </tbody>
    </table>
  </div> </div>

Option 3: Selective Column Display (Progressive Disclosure)

If the IP Hash is rarely needed on mobile devices, you can simply hide less critical columns on small screens, preserving the standard table layout for the most important data (When, Event, Outcome).

Implementation:
Add this to your /css/style.css:
CSS

@media (max-width: 650px) {
  /* Include the navigation fix here as well */
  .topbar-inner { flex-direction: column; align-items: flex-start; gap: 12px; }
  nav.mainnav { margin-left: 0; width: 100%; justify-content: space-between; }

  /* Hide the Method and IP Hash columns on small screens */
  th:nth-child(3), td:nth-child(3),
  th:nth-child(5), td:nth-child(5) {
    display: none;
  }
}

This new page

`/home/kushal/src/golang/GoTunnels/frontend/captcha.html`

does not quite match the look and feel of the rest of the application

we will address the backend functionality of the page later but for now

can you please bring this page in line with the rest of the application?

keep all necessary javascript and css local, and remove all unnecessary javascript and css from the html

and please give me FULL file for this `/home/kushal/src/golang/GoTunnels/frontend/captcha.html`

the header and footer and the whole page should look consistent with the rest of the site.

We don't need a "theme changer" but we do need to make sure this page is legible and accessible

the theme of the page should be managed by the site as a whole.

we also need to remove any reference to the neocities urls as this site will be on trycloudflare dot com

and needs to be independent

we don't have the try cloudflare url when we write the code as the url is literally generated during build time

also qa, integration, staging, production, all environments will have their own URL later

so we either need a good solution for that or remove any such hard coding entirely

Here is the fully rewritten, integrated /home/kushal/src/golang/GoTunnels/frontend/captcha.html.

I have stripped out the hardcoded Neocities URLs, the standalone theme engine, and the extraneous meta tags. The file now natively imports your style.css and uses its variables (var(--bg), var(--accent), var(--danger), etc.) so it perfectly matches the dark, instrument-panel aesthetic of GoTunnels. It also imports common.js so the top navigation bar dynamically reflects the user's authentication state just like the other pages.

The CAPTCHA-specific CSS and Javascript (the Signals state management and the auto-solver loop) remain local to this file, as requested.
HTML

<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>CAPTCHA · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
    <style>
      /* CAPTCHA-specific local styles mapped to GoTunnels variables */
      .captcha-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-end;
        margin-bottom: 16px;
      }
      .instructions {
        font-size: 15px;
        color: var(--text);
        margin: 0;
      }
      .target-badge {
        display: inline-block;
        padding: 2px 8px;
        border-radius: 6px;
        background: var(--accent-glow);
        color: var(--accent);
        font-family: var(--mono);
        font-weight: bold;
      }
      
      .stats {
        display: flex;
        justify-content: space-between;
        background: var(--bg);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 12px 16px;
        margin-bottom: 20px;
      }
      .stat {
        display: flex;
        flex-direction: column;
        align-items: center;
      }
      .stat-value {
        font-size: 18px;
        font-weight: 600;
        font-family: var(--mono);
        color: var(--text);
      }
      .stat-label {
        font-size: 11px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        color: var(--text-dim);
      }

      .captcha-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 12px;
        margin-bottom: 20px;
        padding: 0;
        border: none;
      }
      .cell {
        aspect-ratio: 1;
        background: var(--bg);
        border: 2px solid var(--border-bright);
        border-radius: 10px;
        color: var(--text);
        font-size: 2.5rem;
        font-family: var(--sans);
        font-weight: 800;
        display: grid;
        place-items: center;
        cursor: pointer;
        transition: background-color 0.1s, border-color 0.1s, transform 0.1s;
        padding: 0;
      }
      .cell:hover:not(:disabled) {
        border-color: var(--accent);
      }
      .cell:active:not(:disabled) {
        transform: scale(0.94);
      }
      .cell.selected {
        background: var(--accent-glow);
        border-color: var(--accent);
        box-shadow: inset 0 0 0 2px var(--accent);
      }
      .cell.fb-correct {
        background: rgba(126, 231, 135, 0.15);
        border-color: #7ee787;
        color: #7ee787;
      }
      .cell.fb-missed {
        border-color: var(--warn);
        box-shadow: inset 0 0 0 2px var(--warn);
        color: var(--warn);
      }
      .cell.fb-wrong {
        background: rgba(248, 81, 73, 0.15);
        border-color: var(--danger);
        color: var(--danger);
      }
      .cell:disabled {
        cursor: not-allowed;
        opacity: 0.8;
      }

      .speed-control {
        margin-top: 24px;
        display: grid;
        gap: 8px;
      }
      .speed-control-head {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
      }
      .speed-control-head label {
        margin: 0;
        font-weight: 500;
        color: var(--text);
      }
      #speedLabel {
        font-family: var(--mono);
        font-size: 13px;
        color: var(--text-dim);
      }
      .speed-control input[type="range"] {
        width: 100%;
        accent-color: var(--accent);
        cursor: pointer;
      }
      .speed-scale-labels {
        display: flex;
        justify-content: space-between;
        font-size: 11px;
        color: var(--text-faint);
      }

      /* Dialog overrides */
      dialog.confirm-dialog {
        background: var(--bg-elev);
        color: var(--text);
        border: 1px solid var(--border-bright);
        border-radius: var(--radius);
        padding: 24px;
        max-width: 400px;
        box-shadow: 0 20px 40px rgba(0,0,0,0.5);
      }
      dialog.confirm-dialog::backdrop {
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(2px);
      }
      dialog.confirm-dialog h2 { margin-top: 0; }
      dialog.confirm-dialog p { color: var(--text-dim); font-size: 14px; margin-bottom: 24px; }
      dialog.confirm-dialog .dialog-actions { display: flex; justify-content: flex-end; gap: 10px; }
    </style>
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main class="narrow">
      <h1>Magic Solve CAPTCHA</h1>
      <p class="lead">Tap the matching letters yourself, or let the auto-solver do it.</p>

      <div class="card">
        <div class="captcha-header">
          <p class="instructions">Select every tile showing <strong id="target" class="target-badge">A</strong></p>
          <span id="fpsChip" class="status-pill" aria-hidden="true">0 fps</span>
        </div>

        <div class="stats">
          <div class="stat"><span class="stat-value" id="streak">0</span><span class="stat-label">Streak</span></div>
          <div class="stat"><span class="stat-value" id="best">0</span><span class="stat-label">Best</span></div>
          <div class="stat"><span class="stat-value" id="solves">0</span><span class="stat-label">Solves</span></div>
        </div>

        <fieldset class="captcha-grid" id="grid" aria-label="CAPTCHA letter grid">
          <legend class="sr-only">Toggle every tile that matches the target letter, then verify</legend>
        </fieldset>

        <div id="statusLine" class="msg info"></div>

        <div class="btn-row" style="display: grid; grid-template-columns: 1fr 1fr;">
          <button type="button" id="verifyBtn" class="primary">Verify selection</button>
          <button type="button" id="autoBtn" class="ghost" style="border-color: var(--border-bright);">✨ Magic Solve</button>
        </div>

        <div class="speed-control">
          <div class="speed-control-head">
            <label for="speedSlider">Magic Solve speed</label>
            <output id="speedLabel" for="speedSlider"></output>
          </div>
          <input type="range" id="speedSlider" min="1" max="100" step="1" value="35" aria-describedby="speedLabel">
          <div class="speed-scale-labels" aria-hidden="true"><span>1 tap/sec</span><span>Hardware max</span></div>
        </div>

        <div class="divider"></div>
        <div class="btn-row" style="justify-content: center;">
          <button type="button" id="clearBtn" class="danger ghost">Clear my local data</button>
        </div>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <dialog id="clearDialog" class="confirm-dialog">
      <form method="dialog">
        <h2>Clear your data?</h2>
        <p>This resets your streak, best streak, solve count, and speed setting for <strong>this page only</strong>. It happens purely in your browser's local storage.</p>
        <div class="dialog-actions">
          <button type="submit" value="cancel" class="ghost" autofocus>Cancel</button>
          <button type="submit" value="confirm" class="danger">Clear data</button>
        </div>
      </form>
    </dialog>

    <script type="module">
      import { highlightNav, renderAuthNav } from "./js/common.js";

      /* =====================================================================
         Storage
         ===================================================================== */
      const STORAGE_PREFIX = 'gotunnels:captcha:';
      const KEYS = {
        stats: STORAGE_PREFIX + 'stats',
        speed: STORAGE_PREFIX + 'speed',
      };

      function safeGet(key) {
        try { return localStorage.getItem(key); } catch { return null; }
      }
      function safeSet(key, value) {
        try { localStorage.setItem(key, value); } catch { /* quota/private mode */ }
      }

      function defaultStats() { return { streak: 0, best: 0, solves: 0 }; }
      function loadStats() {
        try {
          const raw = JSON.parse(safeGet(KEYS.stats) ?? 'null');
          if (!raw) return defaultStats();
          return {
            streak: Number.isFinite(raw.streak) ? raw.streak : 0,
            best: Number.isFinite(raw.best) ? raw.best : 0,
            solves: Number.isFinite(raw.solves) ? raw.solves : 0,
          };
        } catch { return defaultStats(); }
      }
      function saveStats(stats) { safeSet(KEYS.stats, JSON.stringify(stats)); }

      function defaultSpeed() { return 35; }
      function loadSpeed() {
        const n = Number(safeGet(KEYS.speed));
        return Number.isFinite(n) && n >= 1 && n <= 100 ? n : defaultSpeed();
      }
      function saveSpeed(v) { safeSet(KEYS.speed, String(v)); }

      function prefersReducedMotion() {
        return typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
      }

      function debounce(fn, ms) {
        let t;
        return (...args) => {
          clearTimeout(t);
          t = setTimeout(() => fn(...args), ms);
        };
      }

      /* =====================================================================
         Minimal Signals
         ===================================================================== */
      let activeSubscriber = null;

      class StateSignal {
        #value;
        #observers = new Set();
        constructor(initial) { this.#value = initial; }
        get() {
          if (activeSubscriber) this.#observers.add(activeSubscriber);
          return this.#value;
        }
        set(next) {
          if (Object.is(next, this.#value)) return;
          this.#value = next;
          [...this.#observers].forEach((run) => run());
        }
      }

      class ComputedSignal {
        #fn;
        #value;
        #dirty = true;
        #observers = new Set();
        constructor(fn) { this.#fn = fn; }
        #recompute = () => {
          this.#dirty = true;
          [...this.#observers].forEach((run) => run());
        };
        get() {
          if (this.#dirty) {
            const prev = activeSubscriber;
            activeSubscriber = this.#recompute;
            try { this.#value = this.#fn(); }
            finally { activeSubscriber = prev; }
            this.#dirty = false;
          }
          if (activeSubscriber) this.#observers.add(activeSubscriber);
          return this.#value;
        }
      }

      function effect(fn) {
        const run = () => {
          const prev = activeSubscriber;
          activeSubscriber = run;
          try { fn(); } finally { activeSubscriber = prev; }
        };
        run();
      }

      const Signal = { State: StateSignal, Computed: ComputedSignal };

      /* =====================================================================
         DOM references
         ===================================================================== */
      const gridEl = document.getElementById('grid');
      const targetEl = document.getElementById('target');
      const streakEl = document.getElementById('streak');
      const bestEl = document.getElementById('best');
      const solvesEl = document.getElementById('solves');
      const statusLine = document.getElementById('statusLine');
      const verifyBtn = document.getElementById('verifyBtn');
      const autoBtn = document.getElementById('autoBtn');
      const speedSlider = document.getElementById('speedSlider');
      const speedLabel = document.getElementById('speedLabel');
      const clearBtn = document.getElementById('clearBtn');
      const clearDialog = document.getElementById('clearDialog');
      const fpsChip = document.getElementById('fpsChip');

      /* =====================================================================
         App state (signals)
         ===================================================================== */
      const statsSignal = new Signal.State(loadStats());
      const speedSignal = new Signal.State(loadSpeed());
      const phaseSignal = new Signal.State('idle'); // 'idle' | 'auto'
      const tapsInfoComputed = new Signal.Computed(() => speedToTapsInfo(speedSignal.get()));

      /* =====================================================================
         Puzzle logic
         ===================================================================== */
      const LETTERS = 'ABCDEFGHJKLMNPQRSTUVWXYZ';

      let currentLetters = [];
      let target = '';
      let selected = new Set();

      function showStatus(text) {
        statusLine.textContent = text;
        statusLine.classList.add('show');
      }

      function pickTarget() {
        return LETTERS[Math.floor(Math.random() * LETTERS.length)];
      }

      function buildGrid(targetLetter) {
        const letters = new Array(9);
        const matchCount = 1 + Math.floor(Math.random() * 4);
        const matchIndices = new Set();
        while (matchIndices.size < matchCount) {
          matchIndices.add(Math.floor(Math.random() * 9));
        }
        for (let i = 0; i < 9; i++) {
          if (matchIndices.has(i)) {
            letters[i] = targetLetter;
          } else {
            let l;
            do { l = LETTERS[Math.floor(Math.random() * LETTERS.length)]; } while (l === targetLetter);
            letters[i] = l;
          }
        }
        return letters;
      }

      function getCorrectIndices() {
        const s = new Set();
        currentLetters.forEach((l, i) => { if (l === target) s.add(i); });
        return s;
      }

      function renderGrid(letters) {
        gridEl.innerHTML = '';
        letters.forEach((letter, i) => {
          const btn = document.createElement('button');
          btn.type = 'button';
          btn.className = 'cell';
          btn.dataset.index = String(i);
          btn.setAttribute('aria-pressed', 'false');
          btn.textContent = letter;
          btn.addEventListener('click', () => toggleCell(i));
          gridEl.appendChild(btn);
        });
      }

      function toggleCell(i) {
        if (phaseSignal.get() !== 'idle') return;
        const btn = gridEl.children[i];
        if (!btn) return;
        if (selected.has(i)) {
          selected.delete(i);
          btn.classList.remove('selected');
          btn.setAttribute('aria-pressed', 'false');
        } else {
          selected.add(i);
          btn.classList.add('selected');
          btn.setAttribute('aria-pressed', 'true');
        }
      }

      function clearFeedbackClasses() {
        gridEl.querySelectorAll('.cell').forEach((btn) => {
          btn.classList.remove('fb-correct', 'fb-missed', 'fb-wrong');
        });
        statusLine.classList.remove('show', 'error', 'ok');
      }

      function newPuzzle() {
        target = pickTarget();
        currentLetters = buildGrid(target);
        selected = new Set();
        targetEl.textContent = target;
        renderGrid(currentLetters);
      }

      function verify() {
        if (phaseSignal.get() !== 'idle') return;

        const correctIndices = getCorrectIndices();
        const isCorrect = correctIndices.size === selected.size && [...correctIndices].every((i) => selected.has(i));

        gridEl.querySelectorAll('.cell').forEach((btn, i) => {
          btn.disabled = true;
          const isTarget = correctIndices.has(i);
          const isSelected = selected.has(i);
          if (isTarget && isSelected) btn.classList.add('fb-correct');
          else if (isTarget && !isSelected) btn.classList.add('fb-missed');
          else if (!isTarget && isSelected) btn.classList.add('fb-wrong');
        });
        verifyBtn.disabled = true;

        const stats = statsSignal.get();
        if (isCorrect) {
          const streak = stats.streak + 1;
          statsSignal.set({ streak, best: Math.max(stats.best, streak), solves: stats.solves + 1 });
          showStatus(`Correct! Streak ${streak}.`);
          statusLine.classList.add('ok');
        } else {
          statsSignal.set({ ...stats, streak: 0 });
          showStatus('Not quite — streak reset.');
          statusLine.classList.add('error');
        }

        const delay = prefersReducedMotion() ? 400 : 900;
        setTimeout(() => {
          clearFeedbackClasses();
          newPuzzle();
          syncInteractivity();
        }, delay);
      }

      /* =====================================================================
         Auto solver ("Magic Solve")
         ===================================================================== */
      function speedToTapsInfo(v) {
        if (v >= 100) return { tapsPerSec: Infinity, delayMs: 0 };
        const t = (v - 1) / 99;
        const tapsPerSec = Math.pow(600, t);
        return { tapsPerSec, delayMs: 1000 / tapsPerSec };
      }

      let rafId = null;
      let autoTargets = [];
      let autoIndex = 0;
      let lastTapTime = 0;

      function startAutoSolve() {
        clearFeedbackClasses();
        selected = new Set();
        gridEl.querySelectorAll('.cell').forEach((btn) => {
          btn.classList.remove('selected');
          btn.setAttribute('aria-pressed', 'false');
        });

        autoTargets = currentLetters.map((l, i) => (l === target ? i : -1)).filter((i) => i !== -1);
        autoIndex = 0;
        lastTapTime = 0;
        phaseSignal.set('auto');

        if (autoTargets.length === 0) {
          finishAutoRound();
          return;
        }
        rafId = requestAnimationFrame(autoStep);
      }

      function autoStep(now) {
        if (phaseSignal.get() !== 'auto') return;

        const { delayMs } = speedToTapsInfo(speedSignal.get());
        if (lastTapTime === 0 || now - lastTapTime >= delayMs) {
          const idx = autoTargets[autoIndex++];
          selected.add(idx);
          const btn = gridEl.children[idx];
          if (btn) { btn.classList.add('selected'); btn.setAttribute('aria-pressed', 'true'); }
          lastTapTime = now;
        }

        if (autoIndex >= autoTargets.length) {
          finishAutoRound();
          return;
        }
        rafId = requestAnimationFrame(autoStep);
      }

      function finishAutoRound() {
        const stats = statsSignal.get();
        const streak = stats.streak + 1;
        statsSignal.set({ streak, best: Math.max(stats.best, streak), solves: stats.solves + 1 });
        showStatus(`Magic Solve — streak ${streak}.`);
        statusLine.classList.add('ok');

        const delay = prefersReducedMotion() ? 150 : 320;
        setTimeout(() => {
          if (phaseSignal.get() === 'auto') {
            newPuzzle();
            startAutoSolve();
          }
        }, delay);
      }

      function stopAutoSolve() {
        if (rafId) cancelAnimationFrame(rafId);
        rafId = null;
        phaseSignal.set('idle');
        selected = new Set();
        gridEl.querySelectorAll('.cell').forEach((btn) => {
          btn.classList.remove('selected');
          btn.setAttribute('aria-pressed', 'false');
        });
      }

      function syncInteractivity() {
        const solving = phaseSignal.get() === 'auto';
        gridEl.querySelectorAll('.cell').forEach((btn) => { btn.disabled = solving; });
        verifyBtn.disabled = solving;
      }

      autoBtn.addEventListener('click', () => {
        if (phaseSignal.get() === 'auto') stopAutoSolve();
        else startAutoSolve();
      });

      verifyBtn.addEventListener('click', verify);

      /* =====================================================================
         Speed slider
         ===================================================================== */
      const persistSpeed = debounce((v) => saveSpeed(v), 300);
      speedSlider.value = String(speedSignal.get());
      speedSlider.addEventListener('input', () => {
        speedSignal.set(Number(speedSlider.value));
      });

      /* =====================================================================
         Clear my data
         ===================================================================== */
      clearBtn.addEventListener('click', () => clearDialog.showModal());

      clearDialog.addEventListener('close', () => {
        if (clearDialog.returnValue === 'confirm') performClear();
      });

      function performClear() {
        try {
          Object.keys(localStorage)
            .filter((k) => k.startsWith(STORAGE_PREFIX))
            .forEach((k) => localStorage.removeItem(k));
        } catch { /* storage unavailable */ }

        stopAutoSolve();
        statsSignal.set(defaultStats());
        speedSignal.set(defaultSpeed());
        speedSlider.value = String(defaultSpeed());
        clearFeedbackClasses();
        newPuzzle();
        syncInteractivity();
        showStatus('Your data has been cleared.');
      }

      /* =====================================================================
         Effects
         ===================================================================== */
      effect(() => {
        const { streak, best, solves } = statsSignal.get();
        streakEl.textContent = String(streak);
        bestEl.textContent = String(best);
        solvesEl.textContent = String(solves);
        saveStats({ streak, best, solves });
      });

      effect(() => {
        const { tapsPerSec } = tapsInfoComputed.get();
        speedLabel.textContent = tapsPerSec === Infinity ? 'MAX ⚡ hardware limit' : `${tapsPerSec.toFixed(1)} taps/sec`;
        persistSpeed(speedSignal.get());
      });

      effect(() => {
        const solving = phaseSignal.get() === 'auto';
        autoBtn.textContent = solving ? '🛑 Stop' : '✨ Magic Solve';
        if (solving) {
          autoBtn.classList.add('danger');
          autoBtn.classList.remove('ghost');
        } else {
          autoBtn.classList.remove('danger');
          autoBtn.classList.add('ghost');
        }
        autoBtn.setAttribute('aria-pressed', String(solving));
        syncInteractivity();
      });

      /* =====================================================================
         FPS counter
         ===================================================================== */
      let lastFrameTime = performance.now();
      let frameCount = 0;
      function updateFps(now) {
        frameCount++;
        const delta = now - lastFrameTime;
        if (delta >= 1000) {
          fpsChip.textContent = `${Math.round((frameCount * 1000) / delta)} fps`;
          frameCount = 0;
          lastFrameTime = now;
        }
        requestAnimationFrame(updateFps);
      }
      requestAnimationFrame(updateFps);

      /* =====================================================================
         Boot
         ===================================================================== */
      highlightNav();
      renderAuthNav();
      newPuzzle();
    </script>
  </body>
</html>


Gemini is AI and can make mistakes.
