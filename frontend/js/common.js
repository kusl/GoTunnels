// common.js — shared helpers used by every page. Importing it also installs
// the theme picker in the top bar (theme-boot.js already applied the saved
// theme pre-paint; this module adds the UI and, when signed in, syncs the
// choice to the account). CSP violation reporting needs no JS anymore: the
// browser reports natively to the same-origin /csp-report path, which Caddy
// proxies to the API (see frontend/Caddyfile).

import { Api, getToken, TOKEN_KEY } from "./api.js";

export function qs(sel, root = document) {
  return root.querySelector(sel);
}

/* =====================================================================
   Cross-tab session sync
   ===================================================================== */

// localStorage fires a `storage` event in OTHER tabs of this origin whenever a
// key changes. When one tab logs out — here, or "everywhere" from settings —
// the token is removed; reflect that in every other open tab immediately so a
// stale tab cannot keep showing a signed-in UI. We react ONLY to the token
// being cleared (oldValue present, newValue gone), never to a login, and
// reload so each page re-runs its own auth logic: public pages re-render as
// logged-out, protected pages redirect to /login via requireAuth().
export function installSessionSync() {
  window.addEventListener("storage", (e) => {
    if (e.key !== TOKEN_KEY) return;
    if (e.newValue === null && e.oldValue) {
      location.reload();
    }
  });
}

void installSessionSync();

export function showMsg(el, text, kind = "error") {
  if (!el) return;
  el.textContent = text;
  el.className = "msg show " + kind;
}

export function clearMsg(el) {
  if (!el) return;
  el.className = "msg";
  el.textContent = "";
}

export function highlightNav() {
  let path = location.pathname.replace(/\.html$/, "");
  if (path === "" || path === "/") path = "/index";
  document.querySelectorAll("nav.mainnav a").forEach((a) => {
    let href = a.getAttribute("href").replace(/\.html$/, "");
    if (href === "/") href = "/index";
    if (href === path) a.classList.add("active");
  });
}

// resolveSession classifies the current session into one of three outcomes,
// which is what lets us honor the rule that a user is NEVER logged out unless
// they ask to be:
//   { user }            -> authenticated
//   { user: null }      -> definitely logged out: no token at all, or the API
//                          rejected the token with 401 (revoked / expired). The
//                          dead token is cleared in that case.
//   { uncertain: true } -> we could not confirm (offline, 5xx, rate-limited).
//                          The token is left UNTOUCHED so a refresh once the
//                          network recovers restores the session.
// The critical point: a transient failure must never clear the token or be
// treated as a logout.
async function resolveSession() {
  if (!getToken()) return { user: null };
  try {
    return { user: await Api.me() };
  } catch (err) {
    if (err && err.status === 401) {
      clearToken();
      return { user: null };
    }
    return { uncertain: true };
  }
}

export async function currentUser() {
  const s = await resolveSession();
  // Nav rendering only needs user|null. "Uncertain" has no user object to show,
  // but the token stays put, so this never signs anyone out.
  return s.user ?? null;
}

export async function requireAuth() {
  const s = await resolveSession();
  if (s.user) return s.user;
  if (s.uncertain) {
    // Could not verify the session (offline / server error). Do NOT bounce to
    // /login — that would be an involuntary logout, which this app refuses to
    // do. Stay on the page; a refresh once the API is reachable loads it, and
    // the token is untouched throughout.
    console.warn("session check could not reach the API; staying put");
    return null;
  }
  // Genuinely logged out (no token, or the token was rejected with 401).
  location.href = "/login";
  return null;
}

// renderAuthNav toggles elements marked with data-auth="in" | "out" based on
// whether the visitor is authenticated, and returns the current user (or null).
export async function renderAuthNav() {
  const u = await currentUser();
  document.querySelectorAll("[data-auth]").forEach((el) => {
    const need = el.getAttribute("data-auth");
    const show = (need === "in" && u) || (need === "out" && !u);
    el.classList.toggle("hidden", !show);
  });
  return u;
}

/* =====================================================================
   Theme
   ===================================================================== */

// Keep THEME_KEY and the value list in sync with theme-boot.js.
export const THEME_KEY = "gotunnels_theme";
export const THEMES = [
  { value: "system", label: "System" },
  { value: "dark", label: "Dark" },
  { value: "light", label: "Light" },
  { value: "solarized-dark", label: "Solarized dark" },
  { value: "solarized-light", label: "Solarized light" },
];

const THEME_PREF_KEY = "ui.theme";

// applyTheme stamps <html data-theme> and stores the choice locally so
// theme-boot.js restores it pre-paint on the next page load. Unknown values
// collapse to "system". Returns the theme actually applied.
export function applyTheme(value) {
  const theme = THEMES.some((t) => t.value === value) ? value : "system";
  document.documentElement.setAttribute("data-theme", theme);
  try {
    localStorage.setItem(THEME_KEY, theme);
  } catch {
    // localStorage can be unavailable; the attribute still applies for now.
  }
  return theme;
}

function currentTheme() {
  return document.documentElement.getAttribute("data-theme") || "system";
}

// initThemePicker injects a small <select> into the top bar on every page.
// Local choice applies instantly; when signed in it is mirrored to the server
// preference "ui.theme" so the theme follows the account across devices, and
// at load the server value (if any) wins over the local copy.
async function initThemePicker() {
  const bar = document.querySelector(".topbar-inner");
  if (!bar || bar.querySelector(".theme-select")) return;

  const select = document.createElement("select");
  select.className = "theme-select";
  select.setAttribute("aria-label", "Color theme");
  for (const t of THEMES) {
    const opt = document.createElement("option");
    opt.value = t.value;
    opt.textContent = t.label;
    select.appendChild(opt);
  }
  select.value = currentTheme();
  select.addEventListener("change", () => {
    const applied = applyTheme(select.value);
    select.value = applied;
    if (getToken()) {
      Api.prefSet(THEME_PREF_KEY, applied).catch(() => {});
    }
  });
  bar.appendChild(select);

  if (!getToken()) return;
  try {
    const res = await Api.prefGet(THEME_PREF_KEY);
    if (res && res.exists && res.value && res.value !== currentTheme()) {
      applyTheme(res.value);
      select.value = currentTheme();
    }
  } catch {
    // Offline or session expired: local theme stands.
  }
}

void initThemePicker();
