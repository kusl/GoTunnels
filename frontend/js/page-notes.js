// page-notes.js — the shared plain-text notes feed.
//
// "Auto-refresh" here is deliberate, boring polling (see
// docs/ARCHITECTURE.md § Live updates): the page fetches the newest 50 notes
// every 5 seconds while visible, pauses when the tab is hidden, refreshes
// immediately when it becomes visible again, and backs off exponentially (up
// to 60s) while the API is unreachable.
//
// Each poll is reconciled against what is already on screen, keyed by note
// id:
//   - notes we hold that fall inside the fetched window but are missing from
//     it were deleted -> removed from the DOM immediately;
//   - new notes written by *you* appear immediately;
//   - new notes by others appear immediately too, unless you have scrolled
//     down to read — then they queue behind a "New notes ↑" pill so the feed
//     does not jump under your thumb.
//
// The author filter is applied server-side (the ?authors= parameter), so the
// pagination cursor, the page size, and the deletion-detection window all
// operate on the filtered feed. Changing the filter resets the feed and
// bumps a generation counter; any poll that was in flight for the previous
// filter is discarded when it lands.
//
// Rendering is plain-text only by construction: note bodies are set with
// textContent, never innerHTML, so nothing in a note can become markup, a
// link, or a script.

import { Api } from "./api.js";
import { qs, showMsg, clearMsg, highlightNav, renderAuthNav, requireAuth } from "./common.js";

const POLL_MS = 5000;
const MAX_BACKOFF_MS = 60000;
const PAGE = 50;
const MAX_CHARS = 500;
const SCROLL_DEFER_PX = 200;

// Where the author selection is remembered. The local copy makes the choice
// instant on this device; the server preference makes it follow the account.
const AUTHORS_LS_KEY = "gotunnels_notes_authors";
const AUTHORS_PREF_KEY = "notes.authors";
const MAX_AUTHOR_FILTER = 50; // mirrors the server's cap

const msgEl = qs("#msg");
const bodyEl = qs("#noteBody");
const charCount = qs("#charCount");
const postBtn = qs("#postBtn");
const feedEl = qs("#feed");
const feedStatus = qs("#feedStatus");
const newPill = qs("#newPill");
const loadOlderBtn = qs("#loadOlder");
const deleteDialog = qs("#deleteDialog");
const authorFilter = qs("#authorFilter");
const authorFilterLabel = qs("#authorFilterLabel");
const authorAllCb = qs("#authorAll");
const authorListEl = qs("#authorList");

let me = null;

// id -> note object for everything currently rendered.
const noteMap = new Map();
// id -> DOM node.
const nodeMap = new Map();
// Notes by other people that arrived while the reader was scrolled down.
const pendingNew = new Map();

let oldestLoadedId = 0; // pagination cursor for "Load older"
let reachedEnd = false;

let pollTimer = null;
let backoffMs = POLL_MS;
let polling = false;
let initialLoaded = false; // suppress "new" highlights on the very first fill

// Empty array = everyone. Otherwise: user ids to show.
let selectedAuthors = [];
// Bumped whenever the filter changes; in-flight polls compare against it and
// discard their response if it no longer matches.
let feedGen = 0;

/* =====================================================================
   Rendering
   ===================================================================== */
function fmtRelative(iso) {
  const then = new Date(iso).getTime();
  if (!Number.isFinite(then)) return "";
  const secs = Math.max(0, Math.floor((Date.now() - then) / 1000));
  if (secs < 45) return "just now";
  if (secs < 3600) return Math.floor(secs / 60) + "m ago";
  if (secs < 86400) return Math.floor(secs / 3600) + "h ago";
  if (secs < 7 * 86400) return Math.floor(secs / 86400) + "d ago";
  return new Date(iso).toLocaleDateString();
}

async function copyText(text, btn) {
  let ok = false;
  try {
    await navigator.clipboard.writeText(text);
    ok = true;
  } catch {
    // Clipboard API can be unavailable (permissions, older browsers);
    // fall back to a hidden textarea + execCommand.
    try {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.className = "sr-only";
      document.body.appendChild(ta);
      ta.select();
      ok = document.execCommand("copy");
      ta.remove();
    } catch { ok = false; }
  }
  const prev = btn.textContent;
  btn.textContent = ok ? "Copied ✓" : "Copy failed";
  btn.disabled = true;
  setTimeout(() => { btn.textContent = prev; btn.disabled = false; }, 1200);
}

function renderNote(n) {
  const card = document.createElement("article");
  card.className = "note-card";
  card.dataset.id = String(n.id);

  const head = document.createElement("div");
  head.className = "note-head";

  const author = document.createElement("span");
  author.className = "note-author";
  author.textContent = n.display_name || n.username;

  const username = document.createElement("span");
  username.className = "note-username";
  username.textContent = "@" + n.username;

  const time = document.createElement("time");
  time.className = "note-time";
  time.dateTime = n.created_at;
  time.title = new Date(n.created_at).toLocaleString();
  time.textContent = fmtRelative(n.created_at);

  head.append(author, username, time);

  const body = document.createElement("p");
  body.className = "note-body";
  // textContent, never innerHTML: bodies stay inert plain text and links in
  // them are not clickable, by design.
  body.textContent = n.body;

  const actions = document.createElement("div");
  actions.className = "note-actions";

  const copyBtn = document.createElement("button");
  copyBtn.type = "button";
  copyBtn.className = "ghost btn-small";
  copyBtn.textContent = "Copy";
  copyBtn.addEventListener("click", () => { void copyText(n.body, copyBtn); });
  actions.appendChild(copyBtn);

  if (me && n.user_id === me.id) {
    const delBtn = document.createElement("button");
    delBtn.type = "button";
    delBtn.className = "danger ghost btn-small";
    delBtn.textContent = "Delete";
    delBtn.addEventListener("click", () => confirmDelete(n.id));
    actions.appendChild(delBtn);
  }

  card.append(head, body, actions);
  return card;
}

// syncDOM makes the on-screen order match noteMap sorted by id descending.
// appendChild moves nodes that already exist, so this is a cheap reorder.
function syncDOM() {
  const ids = [...noteMap.keys()].sort((a, b) => b - a);
  for (const id of ids) {
    let node = nodeMap.get(id);
    if (!node) {
      node = renderNote(noteMap.get(id));
      nodeMap.set(id, node);
    }
    feedEl.appendChild(node);
  }
}

function insertNote(n, highlight = false) {
  if (noteMap.has(n.id)) return;
  noteMap.set(n.id, n);
  syncDOM();
  if (highlight) {
    const node = nodeMap.get(n.id);
    if (node) node.classList.add("note-new");
  }
}

function removeNote(id) {
  noteMap.delete(id);
  pendingNew.delete(id);
  const node = nodeMap.get(id);
  if (node) node.remove();
  nodeMap.delete(id);
}

function refreshTimes() {
  for (const [id, node] of nodeMap) {
    const n = noteMap.get(id);
    if (!n) continue;
    const t = node.querySelector(".note-time");
    if (t) t.textContent = fmtRelative(n.created_at);
  }
}

function updatePill() {
  if (pendingNew.size > 0) {
    newPill.textContent = pendingNew.size === 1 ? "1 new note ↑" : pendingNew.size + " new notes ↑";
    newPill.classList.remove("hidden");
  } else {
    newPill.classList.add("hidden");
  }
}

newPill.addEventListener("click", () => {
  for (const [id, n] of pendingNew) {
    if (!noteMap.has(id)) noteMap.set(id, n);
  }
  pendingNew.clear();
  syncDOM();
  updatePill();
  window.scrollTo({ top: 0, behavior: "smooth" });
});

/* =====================================================================
   Author filter
   ===================================================================== */
function sanitizeIds(v) {
  if (!Array.isArray(v)) return [];
  const uuid = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;
  return v
    .filter((x) => typeof x === "string" && uuid.test(x))
    .map((x) => x.toLowerCase())
    .slice(0, MAX_AUTHOR_FILTER);
}

// loadSelection restores the saved author filter. The server preference (set
// on any device) wins over the local copy; the local copy is the fallback for
// a brand-new pref and keeps working offline.
async function loadSelection() {
  let local = [];
  try {
    const raw = localStorage.getItem(AUTHORS_LS_KEY);
    if (raw) local = sanitizeIds(JSON.parse(raw));
  } catch { /* corrupted or unavailable: ignore */ }
  try {
    const res = await Api.prefGet(AUTHORS_PREF_KEY);
    if (res && res.exists && typeof res.value === "string" && res.value !== "") {
      const server = sanitizeIds(JSON.parse(res.value));
      try { localStorage.setItem(AUTHORS_LS_KEY, JSON.stringify(server)); } catch { /* ok */ }
      return server;
    }
  } catch { /* offline or parse error: fall back to local */ }
  return local;
}

function persistSelection() {
  const payload = JSON.stringify(selectedAuthors);
  try { localStorage.setItem(AUTHORS_LS_KEY, payload); } catch { /* ok */ }
  Api.prefSet(AUTHORS_PREF_KEY, payload).catch(() => {});
}

function filterLabelText() {
  if (selectedAuthors.length === 0) return "Everyone";
  if (selectedAuthors.length === 1) {
    const cb = authorListEl.querySelector('input[value="' + selectedAuthors[0] + '"]');
    const name = cb && cb.dataset.name ? cb.dataset.name : "1 author";
    return name;
  }
  return selectedAuthors.length + " authors";
}

function updateFilterLabel() {
  authorFilterLabel.textContent = filterLabelText();
}

function readCheckboxes() {
  const ids = [];
  authorListEl.querySelectorAll("input[type=checkbox]").forEach((cb) => {
    if (cb.checked) ids.push(cb.value);
  });
  return ids.slice(0, MAX_AUTHOR_FILTER);
}

function syncCheckboxes() {
  const selected = new Set(selectedAuthors);
  authorListEl.querySelectorAll("input[type=checkbox]").forEach((cb) => {
    cb.checked = selected.has(cb.value);
  });
  authorAllCb.checked = selectedAuthors.length === 0;
}

function applySelection(ids) {
  const next = sanitizeIds(ids);
  const changed = JSON.stringify(next) !== JSON.stringify(selectedAuthors);
  selectedAuthors = next;
  syncCheckboxes();
  updateFilterLabel();
  if (changed) {
    persistSelection();
    resetFeed();
  }
}

function renderAuthorOptions(authors) {
  authorListEl.textContent = "";
  for (const a of authors) {
    const label = document.createElement("label");
    label.className = "author-option";
    const cb = document.createElement("input");
    cb.type = "checkbox";
    cb.value = a.user_id;
    cb.dataset.name = a.display_name || a.username;
    const name = document.createElement("span");
    name.textContent = a.display_name || a.username;
    const meta = document.createElement("span");
    meta.className = "author-meta";
    meta.textContent = "@" + a.username + " · " + a.notes;
    label.append(cb, name, meta);
    authorListEl.appendChild(label);
  }
  syncCheckboxes();
  updateFilterLabel();
}

async function refreshAuthors() {
  try {
    const res = await Api.notesAuthors();
    renderAuthorOptions((res && res.authors) || []);
  } catch {
    // The dropdown just stays as it was; the feed itself is unaffected.
  }
}

authorAllCb.addEventListener("change", () => {
  // Checking "Everyone" clears the specific selection; unchecking it with
  // nothing selected would show nobody, so snap back to everyone.
  applySelection([]);
});

authorListEl.addEventListener("change", () => {
  applySelection(readCheckboxes());
});

// Refresh the author list (it changes as people post) each time the panel
// opens, and close the panel when tapping anywhere outside it.
authorFilter.addEventListener("toggle", () => {
  if (authorFilter.open) void refreshAuthors();
});
document.addEventListener("click", (e) => {
  if (authorFilter.open && !authorFilter.contains(e.target)) {
    authorFilter.open = false;
  }
});

// resetFeed clears everything on screen and starts over against the current
// filter. The generation bump makes any in-flight poll discard itself.
function resetFeed() {
  feedGen++;
  noteMap.clear();
  nodeMap.clear();
  pendingNew.clear();
  feedEl.textContent = "";
  oldestLoadedId = 0;
  reachedEnd = false;
  initialLoaded = false;
  loadOlderBtn.classList.add("hidden");
  updatePill();
  if (pollTimer !== null) { clearTimeout(pollTimer); pollTimer = null; }
  backoffMs = POLL_MS;
  setStatus("loading…");
  void pollOnce();
}

/* =====================================================================
   Polling with reconciliation
   ===================================================================== */
function setStatus(text) {
  feedStatus.textContent = text;
}

async function pollOnce() {
  if (polling) return;
  polling = true;
  const gen = feedGen;
  try {
    const res = await Api.notesList({ limit: PAGE, authors: selectedAuthors });
    if (gen !== feedGen) return; // filter changed while in flight: discard
    const fetched = (res && res.notes) || [];
    backoffMs = POLL_MS;
    setStatus(document.visibilityState === "visible" ? "live" : "paused");

    const fetchedIds = new Set(fetched.map((n) => n.id));
    // The poll returns the newest PAGE notes. Anything we hold in that same
    // range but missing from the response was deleted. If the response is
    // empty the whole feed is empty. Notes older than the window cannot be
    // checked this way; a deletion there shows up on the next full reload.
    const windowMin = fetched.length > 0 ? fetched[fetched.length - 1].id : Infinity;
    const toRemove = [];
    for (const id of noteMap.keys()) {
      if ((fetched.length === 0 || id >= windowMin) && !fetchedIds.has(id)) toRemove.push(id);
    }
    for (const [id] of pendingNew) {
      if ((fetched.length === 0 || id >= windowMin) && !fetchedIds.has(id)) pendingNew.delete(id);
    }
    toRemove.forEach(removeNote);

    const deferOthers = window.scrollY > SCROLL_DEFER_PX;
    for (const n of fetched) {
      if (noteMap.has(n.id) || pendingNew.has(n.id)) continue;
      const mine = me && n.user_id === me.id;
      if (mine || !deferOthers || !initialLoaded) {
        insertNote(n, initialLoaded);
      } else {
        pendingNew.set(n.id, n);
      }
    }
    updatePill();
    refreshTimes();

    if (oldestLoadedId === 0 && fetched.length > 0) {
      oldestLoadedId = fetched[fetched.length - 1].id;
    }
    if (fetched.length >= PAGE && !reachedEnd) {
      loadOlderBtn.classList.remove("hidden");
    }
    initialLoaded = true;
    clearMsg(msgEl);
  } catch (err) {
    if (gen !== feedGen) return;
    if (err && err.status === 401) {
      location.href = "/login";
      return;
    }
    backoffMs = Math.min(backoffMs * 2, MAX_BACKOFF_MS);
    setStatus("reconnecting…");
  } finally {
    polling = false;
    if (gen !== feedGen) {
      // This cycle was for a stale filter; poll again right away for the
      // current one instead of waiting out a timer.
      void pollOnce();
    } else {
      scheduleNextPoll();
    }
  }
}

function scheduleNextPoll() {
  if (pollTimer !== null) clearTimeout(pollTimer);
  if (document.visibilityState !== "visible") {
    pollTimer = null;
    setStatus("paused");
    return;
  }
  pollTimer = setTimeout(() => {
    pollTimer = null;
    void pollOnce();
  }, backoffMs);
}

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible") {
    backoffMs = POLL_MS;
    void pollOnce(); // refresh immediately on return
  } else {
    if (pollTimer !== null) { clearTimeout(pollTimer); pollTimer = null; }
    setStatus("paused");
  }
});

/* =====================================================================
   Pagination ("Load older")
   ===================================================================== */
loadOlderBtn.addEventListener("click", async () => {
  loadOlderBtn.disabled = true;
  const gen = feedGen;
  try {
    const res = await Api.notesList({ before: oldestLoadedId, limit: PAGE, authors: selectedAuthors });
    if (gen !== feedGen) return; // filter changed while loading: discard
    const older = (res && res.notes) || [];
    for (const n of older) {
      if (!noteMap.has(n.id)) noteMap.set(n.id, n);
    }
    if (older.length > 0) {
      oldestLoadedId = older[older.length - 1].id;
      syncDOM();
    }
    if (older.length < PAGE) {
      reachedEnd = true;
      loadOlderBtn.classList.add("hidden");
    }
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    showMsg(msgEl, "Could not load older notes: " + err.message);
  } finally {
    loadOlderBtn.disabled = false;
  }
});

/* =====================================================================
   Composer
   ===================================================================== */
// countChars counts what the server counts: Unicode code points, not UTF-16
// units. "👨‍👩‍👧" is several code points but even more UTF-16 units, which is
// exactly why the old maxlength attribute (which counts UTF-16 units) could
// block typing while the visible counter still showed room to spare.
function countChars(s) {
  return [...s].length;
}

function updateCharCount() {
  const n = countChars(bodyEl.value);
  const over = n > MAX_CHARS;
  charCount.textContent = n + " / " + MAX_CHARS;
  charCount.classList.toggle("over", over);
  // Soft limit: typing and pasting are never blocked mid-thought; the Post
  // button simply stays off until the note fits. The server re-validates.
  postBtn.disabled = over || bodyEl.value.trim() === "";
}
bodyEl.addEventListener("input", updateCharCount);

postBtn.addEventListener("click", async () => {
  const text = bodyEl.value.trim();
  if (!text) {
    showMsg(msgEl, "Write something first.");
    return;
  }
  if (countChars(text) > MAX_CHARS) {
    showMsg(msgEl, "Notes are capped at " + MAX_CHARS + " characters — trim it a little.");
    return;
  }
  postBtn.disabled = true;
  clearMsg(msgEl);
  try {
    const res = await Api.noteCreate(text);
    if (res && res.note) {
      insertNote(res.note, true);
      if (oldestLoadedId === 0) oldestLoadedId = res.note.id;
    }
    bodyEl.value = "";
    window.scrollTo({ top: 0, behavior: "smooth" });
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    if (err && err.status === 429) {
      showMsg(msgEl, "You are posting too fast — wait a moment and try again.");
    } else {
      showMsg(msgEl, err.message || "Could not post the note.");
    }
  } finally {
    updateCharCount(); // re-enables Post exactly when the content allows it
  }
});

/* =====================================================================
   Delete (own notes only)
   ===================================================================== */
let deleteTargetId = null;

function confirmDelete(id) {
  deleteTargetId = id;
  deleteDialog.showModal();
}

deleteDialog.addEventListener("close", () => {
  const id = deleteTargetId;
  deleteTargetId = null;
  if (deleteDialog.returnValue !== "confirm" || id === null) return;
  void (async () => {
    try {
      await Api.noteDelete(id);
      removeNote(id);
      updatePill();
    } catch (err) {
      if (err && err.status === 401) { location.href = "/login"; return; }
      // 404 means it is already gone (another tab, another device).
      if (err && err.status === 404) { removeNote(id); updatePill(); return; }
      showMsg(msgEl, "Could not delete the note: " + err.message);
    }
  })();
});

/* =====================================================================
   Boot
   ===================================================================== */
async function boot() {
  highlightNav();
  renderAuthNav();

  me = await requireAuth();
  if (!me) return; // redirected to /login

  updateCharCount();

  selectedAuthors = await loadSelection();
  await refreshAuthors(); // also syncs checkboxes + label to the selection

  setStatus("loading…");
  await pollOnce(); // initial load; also schedules the next poll
}

void boot();
