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

const msgEl = qs("#msg");
const bodyEl = qs("#noteBody");
const charCount = qs("#charCount");
const postBtn = qs("#postBtn");
const feedEl = qs("#feed");
const feedStatus = qs("#feedStatus");
const newPill = qs("#newPill");
const loadOlderBtn = qs("#loadOlder");
const deleteDialog = qs("#deleteDialog");

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
   Polling with reconciliation
   ===================================================================== */
function setStatus(text) {
  feedStatus.textContent = text;
}

async function pollOnce() {
  if (polling) return;
  polling = true;
  try {
    const res = await Api.notesList({ limit: PAGE });
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
    if (err && err.status === 401) {
      location.href = "/login";
      return;
    }
    backoffMs = Math.min(backoffMs * 2, MAX_BACKOFF_MS);
    setStatus("reconnecting…");
  } finally {
    polling = false;
    scheduleNextPoll();
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
  try {
    const res = await Api.notesList({ before: oldestLoadedId, limit: PAGE });
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
function updateCharCount() {
  // Count what the server counts: unicode code points, not UTF-16 units.
  charCount.textContent = [...bodyEl.value].length + " / " + MAX_CHARS;
}
bodyEl.addEventListener("input", updateCharCount);

postBtn.addEventListener("click", async () => {
  const text = bodyEl.value.trim();
  if (!text) {
    showMsg(msgEl, "Write something first.");
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
    updateCharCount();
    window.scrollTo({ top: 0, behavior: "smooth" });
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    if (err && err.status === 429) {
      showMsg(msgEl, "You are posting too fast — wait a moment and try again.");
    } else {
      showMsg(msgEl, err.message || "Could not post the note.");
    }
  } finally {
    postBtn.disabled = false;
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
  setStatus("loading…");
  await pollOnce(); // initial load; also schedules the next poll
}

void boot();
