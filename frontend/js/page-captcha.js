// page-captcha.js — the Magic Solve CAPTCHA page.
//
// The puzzle itself (grid, verify, auto-solver) runs entirely in the browser,
// but the *statistics* live on the server so they survive devices and feed the
// leaderboard. Because the auto-solver can complete hundreds of puzzles per
// second, we never send one request per solve. Instead the page accumulates
// deltas locally and flushes them in small batches:
//
//   - every few seconds while anything is pending,
//   - when the auto-solver stops,
//   - and with fetch({keepalive:true}) when the tab is hidden or unloading,
//     so the last batch survives navigation. (navigator.sendBeacon cannot be
//     used here: it cannot carry the Authorization header.)
//
// Displayed solves = server base + un-flushed pending + in-flight, so the
// numbers never jump backwards while a sync is in the air.
//
// Settings split (see docs/ARCHITECTURE.md):
//   - leaderboard open/closed  -> server-side user preference (follows account)
//   - Magic Solve speed        -> localStorage (depends on this device's display)

import { highlightNav, renderAuthNav, requireAuth } from "./common.js";
import { Api } from "./api.js";

/* =====================================================================
   Local storage (device-scoped settings only)
   ===================================================================== */
const STORAGE_PREFIX = "gotunnels:captcha:";
const KEYS = {
  speed: STORAGE_PREFIX + "speed",
  // Legacy key from the frontend-only version of this page; stats now live
  // on the server, so we delete it on boot.
  legacyStats: STORAGE_PREFIX + "stats",
};
const LEADERBOARD_PREF_KEY = "captcha.leaderboard.open";

function safeGet(key) {
  try { return localStorage.getItem(key); } catch { return null; }
}
function safeSet(key, value) {
  try { localStorage.setItem(key, value); } catch { /* quota/private mode */ }
}
function safeRemove(key) {
  try { localStorage.removeItem(key); } catch { /* storage unavailable */ }
}

function defaultSpeed() { return 35; }
function loadSpeed() {
  const n = Number(safeGet(KEYS.speed));
  return Number.isFinite(n) && n >= 1 && n <= 100 ? n : defaultSpeed();
}
function saveSpeed(v) { safeSet(KEYS.speed, String(v)); }

function prefersReducedMotion() {
  return typeof matchMedia === "function" && matchMedia("(prefers-reduced-motion: reduce)").matches;
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
const gridEl = document.getElementById("grid");
const targetEl = document.getElementById("target");
const streakEl = document.getElementById("streak");
const bestEl = document.getElementById("best");
const solvesEl = document.getElementById("solves");
const statusLine = document.getElementById("statusLine");
const verifyBtn = document.getElementById("verifyBtn");
const autoBtn = document.getElementById("autoBtn");
const speedSlider = document.getElementById("speedSlider");
const speedLabel = document.getElementById("speedLabel");
const clearBtn = document.getElementById("clearBtn");
const clearDialog = document.getElementById("clearDialog");
const fpsChip = document.getElementById("fpsChip");
const syncChip = document.getElementById("syncChip");
const leaderboardEl = document.getElementById("leaderboard");
const lbRows = document.getElementById("lbRows");
const lbStatus = document.getElementById("lbStatus");
const lbMe = document.getElementById("lbMe");

/* =====================================================================
   App state (signals)
   ===================================================================== */
// Server-acknowledged values.
const baseSolvesSignal = new Signal.State(0); // manual+auto totals from server
const bestSignal = new Signal.State(0);
// Live, local values.
const streakSignal = new Signal.State(0);
const pendingSignal = new Signal.State(0); // pending + in-flight solve deltas
const speedSignal = new Signal.State(loadSpeed());
const phaseSignal = new Signal.State("idle"); // 'idle' | 'auto'
const syncStateSignal = new Signal.State("idle"); // 'idle' | 'dirty' | 'syncing' | 'error'
const solvesComputed = new Signal.Computed(() => baseSolvesSignal.get() + pendingSignal.get());
const tapsInfoComputed = new Signal.Computed(() => speedToTapsInfo(speedSignal.get()));

let me = null; // current user, set on boot

/* =====================================================================
   Server sync engine
   ===================================================================== */
const SYNC_INTERVAL_MS = 4000;

// Deltas not yet handed to a request.
let pending = { manual: 0, auto: 0 };
// Deltas currently inside an in-flight request (restored to pending on error).
let inflight = null;
let dirtyStreak = false; // streak/best changed since last successful sync
let syncTimer = null;
let syncInFlight = false;

function pendingTotal() {
  return pending.manual + pending.auto + (inflight ? inflight.manual + inflight.auto : 0);
}
function refreshPendingSignal() {
  pendingSignal.set(pendingTotal());
}

function recordSolve(mode) {
  if (mode === "auto") pending.auto += 1;
  else pending.manual += 1;
  dirtyStreak = true;
  refreshPendingSignal();
  syncStateSignal.set("dirty");
  scheduleSync();
}

function recordStreakReset() {
  dirtyStreak = true;
  syncStateSignal.set("dirty");
  scheduleSync();
}

function scheduleSync() {
  if (syncTimer !== null) return; // a flush is already coming
  syncTimer = setTimeout(() => {
    syncTimer = null;
    void flushSync();
  }, SYNC_INTERVAL_MS);
}

function hasWork() {
  return pending.manual > 0 || pending.auto > 0 || dirtyStreak;
}

function applyServer(stats) {
  baseSolvesSignal.set((stats.manual_solves || 0) + (stats.auto_solves || 0));
  bestSignal.set(Math.max(bestSignal.get(), stats.best_streak || 0));
}

// flushSync sends everything accumulated so far. `keepalive` is used on
// pagehide/visibility-hidden so the request can outlive the page. The
// returned promise is also stored in syncPromise so performReset can wait
// out an in-flight request before deleting the server row.
let syncPromise = null;

function flushSync(keepalive = false) {
  if (syncInFlight || !hasWork()) return Promise.resolve();
  syncPromise = doFlush(keepalive);
  return syncPromise;
}

async function doFlush(keepalive) {
  syncInFlight = true;

  inflight = pending;
  pending = { manual: 0, auto: 0 };
  const hadDirtyStreak = dirtyStreak;
  dirtyStreak = false;

  const body = {
    manual_delta: inflight.manual,
    auto_delta: inflight.auto,
    current_streak: streakSignal.get(),
    best_streak: bestSignal.get(),
  };

  syncStateSignal.set("syncing");
  try {
    const res = await Api.captchaSync(body, keepalive);
    inflight = null;
    refreshPendingSignal();
    if (res && res.stats) applyServer(res.stats);
    if (hasWork()) {
      syncStateSignal.set("dirty");
      scheduleSync();
    } else {
      syncStateSignal.set("idle");
    }
  } catch (err) {
    // Restore what we tried to send so nothing is lost, then retry later.
    pending.manual += inflight.manual;
    pending.auto += inflight.auto;
    inflight = null;
    dirtyStreak = dirtyStreak || hadDirtyStreak;
    refreshPendingSignal();
    syncStateSignal.set("error");
    if (err && err.status === 401) {
      location.href = "/login";
      return;
    }
    scheduleSync();
  } finally {
    syncInFlight = false;
    syncPromise = null;
  }
}

// Flush with keepalive when the page goes away; this is the reliable path for
// "user closes the tab mid-auto-solve".
document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "hidden") void flushSync(true);
});
window.addEventListener("pagehide", () => { void flushSync(true); });

/* =====================================================================
   Puzzle logic (unchanged from the frontend-only version)
   ===================================================================== */
const LETTERS = "ABCDEFGHJKLMNPQRSTUVWXYZ";

let currentLetters = [];
let target = "";
let selected = new Set();

function showStatus(text) {
  statusLine.textContent = text;
  statusLine.classList.add("show");
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
  gridEl.innerHTML = "";
  letters.forEach((letter, i) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "cell";
    btn.dataset.index = String(i);
    btn.setAttribute("aria-pressed", "false");
    btn.textContent = letter;
    btn.addEventListener("click", () => toggleCell(i));
    gridEl.appendChild(btn);
  });
}

function toggleCell(i) {
  if (phaseSignal.get() !== "idle") return;
  const btn = gridEl.children[i];
  if (!btn) return;
  if (selected.has(i)) {
    selected.delete(i);
    btn.classList.remove("selected");
    btn.setAttribute("aria-pressed", "false");
  } else {
    selected.add(i);
    btn.classList.add("selected");
    btn.setAttribute("aria-pressed", "true");
  }
}

function clearFeedbackClasses() {
  gridEl.querySelectorAll(".cell").forEach((btn) => {
    btn.classList.remove("fb-correct", "fb-missed", "fb-wrong");
  });
  statusLine.classList.remove("show", "error", "ok");
}

function newPuzzle() {
  target = pickTarget();
  currentLetters = buildGrid(target);
  selected = new Set();
  targetEl.textContent = target;
  renderGrid(currentLetters);
}

function verify() {
  if (phaseSignal.get() !== "idle") return;

  const correctIndices = getCorrectIndices();
  const isCorrect = correctIndices.size === selected.size && [...correctIndices].every((i) => selected.has(i));

  gridEl.querySelectorAll(".cell").forEach((btn, i) => {
    btn.disabled = true;
    const isTarget = correctIndices.has(i);
    const isSelected = selected.has(i);
    if (isTarget && isSelected) btn.classList.add("fb-correct");
    else if (isTarget && !isSelected) btn.classList.add("fb-missed");
    else if (!isTarget && isSelected) btn.classList.add("fb-wrong");
  });
  verifyBtn.disabled = true;

  if (isCorrect) {
    const streak = streakSignal.get() + 1;
    streakSignal.set(streak);
    bestSignal.set(Math.max(bestSignal.get(), streak));
    recordSolve("manual");
    showStatus(`Correct! Streak ${streak}.`);
    statusLine.classList.add("ok");
  } else {
    streakSignal.set(0);
    recordStreakReset();
    showStatus("Not quite — streak reset.");
    statusLine.classList.add("error");
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
  gridEl.querySelectorAll(".cell").forEach((btn) => {
    btn.classList.remove("selected");
    btn.setAttribute("aria-pressed", "false");
  });

  autoTargets = currentLetters.map((l, i) => (l === target ? i : -1)).filter((i) => i !== -1);
  autoIndex = 0;
  lastTapTime = 0;
  phaseSignal.set("auto");

  if (autoTargets.length === 0) {
    finishAutoRound();
    return;
  }
  rafId = requestAnimationFrame(autoStep);
}

function autoStep(now) {
  if (phaseSignal.get() !== "auto") return;

  const { delayMs } = speedToTapsInfo(speedSignal.get());
  if (lastTapTime === 0 || now - lastTapTime >= delayMs) {
    const idx = autoTargets[autoIndex++];
    selected.add(idx);
    const btn = gridEl.children[idx];
    if (btn) { btn.classList.add("selected"); btn.setAttribute("aria-pressed", "true"); }
    lastTapTime = now;
  }

  if (autoIndex >= autoTargets.length) {
    finishAutoRound();
    return;
  }
  rafId = requestAnimationFrame(autoStep);
}

function finishAutoRound() {
  const streak = streakSignal.get() + 1;
  streakSignal.set(streak);
  bestSignal.set(Math.max(bestSignal.get(), streak));
  recordSolve("auto");
  showStatus(`Magic Solve — streak ${streak}.`);
  statusLine.classList.add("ok");

  const delay = prefersReducedMotion() ? 150 : 320;
  setTimeout(() => {
    if (phaseSignal.get() === "auto") {
      newPuzzle();
      startAutoSolve();
    }
  }, delay);
}

// haltAutoSolve stops the animation loop without flushing; stopAutoSolve is
// the user-facing stop, which also pushes the accumulated batch.
function haltAutoSolve() {
  if (rafId) cancelAnimationFrame(rafId);
  rafId = null;
  phaseSignal.set("idle");
  selected = new Set();
  gridEl.querySelectorAll(".cell").forEach((btn) => {
    btn.classList.remove("selected");
    btn.setAttribute("aria-pressed", "false");
  });
}

function stopAutoSolve() {
  haltAutoSolve();
  // The solver may have racked up a large batch; push it promptly.
  void flushSync();
}

function syncInteractivity() {
  const solving = phaseSignal.get() === "auto";
  gridEl.querySelectorAll(".cell").forEach((btn) => { btn.disabled = solving; });
  verifyBtn.disabled = solving;
}

autoBtn.addEventListener("click", () => {
  if (phaseSignal.get() === "auto") stopAutoSolve();
  else startAutoSolve();
});

verifyBtn.addEventListener("click", verify);

/* =====================================================================
   Speed slider (device setting -> localStorage)
   ===================================================================== */
const persistSpeed = debounce((v) => saveSpeed(v), 300);
speedSlider.value = String(speedSignal.get());
speedSlider.addEventListener("input", () => {
  speedSignal.set(Number(speedSlider.value));
});

/* =====================================================================
   Leaderboard (collapsed by default; open state is a server-side pref)
   ===================================================================== */
const LB_REFRESH_MIN_MS = 5000;
let lbLastLoaded = 0;
let lbLoading = false;
let restoringPref = false; // guards the toggle listener during programmatic open

function lbCell(text, mono = false) {
  const td = document.createElement("td");
  td.textContent = text;
  if (mono) td.classList.add("mono");
  return td;
}

async function loadLeaderboard(force = false) {
  if (lbLoading) return;
  const now = Date.now();
  if (!force && now - lbLastLoaded < LB_REFRESH_MIN_MS) return;
  lbLoading = true;
  lbStatus.textContent = "loading…";
  try {
    const res = await Api.captchaLeaderboard();
    lbLastLoaded = Date.now();
    const rows = (res && res.leaderboard) || [];
    lbRows.innerHTML = "";
    if (rows.length === 0) {
      const tr = document.createElement("tr");
      const td = document.createElement("td");
      td.colSpan = 4;
      td.textContent = "No entries yet — solve a puzzle to appear here.";
      tr.appendChild(td);
      lbRows.appendChild(tr);
    }
    for (const row of rows) {
      const tr = document.createElement("tr");
      if (me && row.user_id === me.id) tr.classList.add("lb-row-me");
      tr.appendChild(lbCell(String(row.rank)));
      tr.appendChild(lbCell(row.display_name || row.username));
      tr.appendChild(lbCell(String(row.best_streak), true));
      tr.appendChild(lbCell(String(row.total_solves), true));
      lbRows.appendChild(tr);
    }
    const mine = res && res.me;
    if (mine && !rows.some((r) => me && r.user_id === me.id)) {
      lbMe.textContent = `You are #${mine.rank} with a best streak of ${mine.best_streak} (${mine.total_solves} solves).`;
      lbMe.classList.remove("hidden");
    } else {
      lbMe.classList.add("hidden");
    }
    lbStatus.textContent = "";
  } catch {
    lbStatus.textContent = "failed to load";
  } finally {
    lbLoading = false;
  }
}

leaderboardEl.addEventListener("toggle", () => {
  if (leaderboardEl.open) void loadLeaderboard(true);
  // The toggle event fires asynchronously, so the boot-time restore consumes
  // its guard flag here rather than resetting it synchronously.
  if (restoringPref) { restoringPref = false; return; }
  void Api.prefSet(LEADERBOARD_PREF_KEY, leaderboardEl.open ? "1" : "0").catch(() => {});
});

/* =====================================================================
   Reset (server + local)
   ===================================================================== */
clearBtn.addEventListener("click", () => clearDialog.showModal());

clearDialog.addEventListener("close", () => {
  if (clearDialog.returnValue === "confirm") void performReset();
});

async function performReset() {
  haltAutoSolve();

  // Discard everything queued locally and cancel the pending flush timer.
  pending = { manual: 0, auto: 0 };
  dirtyStreak = false;
  if (syncTimer !== null) { clearTimeout(syncTimer); syncTimer = null; }

  // Let any airborne sync land (or fail) before deleting the server row, so
  // a stale response cannot resurrect stats afterwards. Its error handler may
  // restore deltas into `pending`; discard those too.
  if (syncPromise) { try { await syncPromise; } catch { /* already handled */ } }
  pending = { manual: 0, auto: 0 };
  dirtyStreak = false;
  if (syncTimer !== null) { clearTimeout(syncTimer); syncTimer = null; }
  refreshPendingSignal();

  try {
    await Api.captchaReset();
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    showStatus("Reset failed — please try again.");
    statusLine.classList.add("error");
    return;
  }

  safeRemove(KEYS.speed);
  streakSignal.set(0);
  bestSignal.set(0);
  baseSolvesSignal.set(0);
  speedSignal.set(defaultSpeed());
  speedSlider.value = String(defaultSpeed());
  syncStateSignal.set("idle");
  clearFeedbackClasses();
  newPuzzle();
  syncInteractivity();
  showStatus("Your stats have been reset.");
  if (leaderboardEl.open) void loadLeaderboard(true);
}

/* =====================================================================
   Effects (signals -> DOM)
   ===================================================================== */
effect(() => {
  streakEl.textContent = String(streakSignal.get());
});
effect(() => {
  bestEl.textContent = String(bestSignal.get());
});
effect(() => {
  solvesEl.textContent = String(solvesComputed.get());
});

effect(() => {
  const { tapsPerSec } = tapsInfoComputed.get();
  speedLabel.textContent = tapsPerSec === Infinity ? "MAX ⚡ hardware limit" : `${tapsPerSec.toFixed(1)} taps/sec`;
  persistSpeed(speedSignal.get());
});

effect(() => {
  const solving = phaseSignal.get() === "auto";
  autoBtn.textContent = solving ? "🛑 Stop" : "✨ Magic Solve";
  if (solving) {
    autoBtn.classList.add("danger");
    autoBtn.classList.remove("ghost");
  } else {
    autoBtn.classList.remove("danger");
    autoBtn.classList.add("ghost");
  }
  autoBtn.setAttribute("aria-pressed", String(solving));
  syncInteractivity();
});

effect(() => {
  const state = syncStateSignal.get();
  const labels = { idle: "synced", dirty: "pending…", syncing: "syncing…", error: "sync failed" };
  syncChip.textContent = labels[state] || state;
  syncChip.classList.toggle("pill-error", state === "error");
});

// Refresh an open leaderboard after each successful sync so your own row is
// current, without hammering the endpoint (loadLeaderboard throttles itself).
effect(() => {
  if (syncStateSignal.get() === "idle" && leaderboardEl.open) {
    void loadLeaderboard(false);
  }
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
async function boot() {
  highlightNav();
  renderAuthNav();

  me = await requireAuth();
  if (!me) return; // redirected to /login

  safeRemove(KEYS.legacyStats);

  newPuzzle();
  syncInteractivity();

  // Load server stats and the leaderboard pref in parallel.
  const [statsRes, prefRes] = await Promise.allSettled([
    Api.captchaStats(),
    Api.prefGet(LEADERBOARD_PREF_KEY),
  ]);

  if (statsRes.status === "fulfilled" && statsRes.value && statsRes.value.stats) {
    const s = statsRes.value.stats;
    applyServer(s);
    // Resume the streak you left off with on any device.
    streakSignal.set(s.current_streak || 0);
    syncStateSignal.set("idle");
  } else {
    syncChip.textContent = "offline";
  }

  if (prefRes.status === "fulfilled" && prefRes.value && prefRes.value.exists && prefRes.value.value === "1") {
    restoringPref = true;
    leaderboardEl.open = true; // queues a toggle event, which consumes the flag
  }
}

void boot();
