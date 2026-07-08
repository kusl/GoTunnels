// api.js — the single place browser code talks to the API.
//
// The primary session transport is a Bearer token kept in localStorage. The
// server also sets a cross-site cookie as a secondary path, so we send
// credentials too; but because third-party cookies are increasingly blocked,
// the Bearer token is what we rely on.
//
// Why localStorage (not sessionStorage)?
//   - localStorage is shared across every tab and window of this origin, so
//     opening a link in a new tab — or a second window — stays signed in.
//     sessionStorage is scoped to a single tab, which made every new tab open
//     as logged out (e.g. right-clicking a nav link -> "open in new tab" bounced
//     you to /login even though you were signed in).
//   - localStorage also survives closing and reopening the browser, so a
//     session is never dropped just because a tab or the browser was closed.
// GoTunnels only ends a session when the user explicitly logs out (via the
// settings page, here, or "everywhere") — never for inactivity, never on close.

import { loadConfig } from "./config.js";

export const TOKEN_KEY = "gotunnels_token";

// One-time migration: earlier builds stored the token in sessionStorage. If a
// tab still has one there and localStorage does not, carry it over so the
// storage switch does not sign anyone out on their next visit.
(function migrateLegacyToken() {
  try {
    if (!localStorage.getItem(TOKEN_KEY)) {
      const legacy = sessionStorage.getItem(TOKEN_KEY);
      if (legacy) {
        localStorage.setItem(TOKEN_KEY, legacy);
        sessionStorage.removeItem(TOKEN_KEY);
      }
    }
  } catch {
    // Storage can be unavailable (private mode); callers degrade gracefully.
  }
})();

export function getToken() {
  try {
    return localStorage.getItem(TOKEN_KEY) || "";
  } catch {
    return "";
  }
}
export function setToken(t) {
  if (!t) return;
  try {
    localStorage.setItem(TOKEN_KEY, t);
  } catch {
    // Quota/private mode: the in-memory fetch below still carries no token,
    // so the user simply is not "remembered" — acceptable degradation.
  }
}
export function clearToken() {
  try {
    localStorage.removeItem(TOKEN_KEY);
  } catch {
    /* storage unavailable */
  }
  try {
    sessionStorage.removeItem(TOKEN_KEY); // clean up any legacy copy too
  } catch {
    /* storage unavailable */
  }
}

// apiFetch performs a JSON request against the API base and throws an Error
// (with .status and .data) on non-2xx responses.
//
// Pass { keepalive: true } for small requests that must survive the page
// being hidden or unloaded (e.g. flushing CAPTCHA stats). We cannot use
// navigator.sendBeacon because it cannot carry the Authorization header.
export async function apiFetch(path, opts = {}) {
  const { method = "GET", body, headers = {}, keepalive = false } = opts;
  const cfg = await loadConfig();
  const base = cfg.apiBase || "";

  const h = { Accept: "application/json", ...headers };
  const token = getToken();
  if (token) h["Authorization"] = "Bearer " + token;

  let payload;
  if (body !== undefined) {
    h["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }

  const res = await fetch(base + path, {
    method,
    headers: h,
    body: payload,
    mode: "cors",
    credentials: "include",
    keepalive,
  });

  const text = await res.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { raw: text };
    }
  }

  if (!res.ok) {
    const msg = data && data.error ? data.error : "request failed (" + res.status + ")";
    const err = new Error(msg);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

// Typed-ish endpoint wrappers.
export const Api = {
  signup: (b) => apiFetch("/api/signup", { method: "POST", body: b }),
  login: (b) => apiFetch("/api/login", { method: "POST", body: b }),
  logout: () => apiFetch("/api/logout", { method: "POST" }),
  logoutAll: () => apiFetch("/api/logout-all", { method: "POST" }),
  me: () => apiFetch("/api/me"),
  activity: () => apiFetch("/api/activity"),
  info: () => apiFetch("/api/info"),
  cspRecent: (limit) =>
    apiFetch("/api/csp-reports/recent" + (limit ? "?limit=" + encodeURIComponent(limit) : "")),

  passkeyRegisterBegin: () => apiFetch("/api/passkey/register/begin", { method: "POST" }),
  passkeyRegisterFinish: (flow, body) =>
    apiFetch("/api/passkey/register/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),
  passkeyLoginBegin: (b) => apiFetch("/api/passkey/login/begin", { method: "POST", body: b }),
  passkeyLoginFinish: (flow, body) =>
    apiFetch("/api/passkey/login/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),
  passkeySignupBegin: (b) => apiFetch("/api/passkey/signup/begin", { method: "POST", body: b }),
  passkeySignupFinish: (flow, body) =>
    apiFetch("/api/passkey/signup/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),

  totpEnroll: () => apiFetch("/api/totp/enroll", { method: "POST" }),
  totpConfirm: (b) => apiFetch("/api/totp/confirm", { method: "POST", body: b }),
  totpDisable: (b) => apiFetch("/api/totp/disable", { method: "POST", body: b }),

  captchaStats: () => apiFetch("/api/captcha/stats"),
  captchaSync: (b, keepalive = false) =>
    apiFetch("/api/captcha/sync", { method: "POST", body: b, keepalive }),
  captchaReset: () => apiFetch("/api/captcha/reset", { method: "POST" }),
  captchaLeaderboard: () => apiFetch("/api/captcha/leaderboard"),

  prefGet: (key) => apiFetch("/api/prefs/" + encodeURIComponent(key)),
  prefSet: (key, value) =>
    apiFetch("/api/prefs/" + encodeURIComponent(key), { method: "PUT", body: { value } }),

  notesList: (params = {}) => {
    const q = new URLSearchParams();
    if (params.before) q.set("before", String(params.before));
    if (params.limit) q.set("limit", String(params.limit));
    if (Array.isArray(params.authors) && params.authors.length > 0) {
      q.set("authors", params.authors.join(","));
    }
    const qs = q.toString();
    return apiFetch("/api/notes" + (qs ? "?" + qs : ""));
  },
  notesAuthors: () => apiFetch("/api/notes/authors"),
  noteCreate: (body) => apiFetch("/api/notes", { method: "POST", body: { body } }),
  noteDelete: (id) => apiFetch("/api/notes/" + encodeURIComponent(id), { method: "DELETE" }),
};
