// api.js — the single place browser code talks to the API.
//
// The primary session transport is a Bearer token kept in sessionStorage. The
// server also sets a cross-site cookie as a secondary path, so we send
// credentials too; but because third-party cookies are increasingly blocked,
// the Bearer token is what we rely on.

import { loadConfig } from "./config.js";

const TOKEN_KEY = "gotunnels_token";

export function getToken() {
  return sessionStorage.getItem(TOKEN_KEY) || "";
}
export function setToken(t) {
  if (t) sessionStorage.setItem(TOKEN_KEY, t);
}
export function clearToken() {
  sessionStorage.removeItem(TOKEN_KEY);
}

// apiFetch performs a JSON request against the API base and throws an Error
// (with .status and .data) on non-2xx responses.
export async function apiFetch(path, opts = {}) {
  const { method = "GET", body, headers = {} } = opts;
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
  me: () => apiFetch("/api/me"),
  activity: () => apiFetch("/api/activity"),
  info: () => apiFetch("/api/info"),

  passkeyRegisterBegin: () => apiFetch("/api/passkey/register/begin", { method: "POST" }),
  passkeyRegisterFinish: (flow, body) =>
    apiFetch("/api/passkey/register/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),
  passkeyLoginBegin: (b) => apiFetch("/api/passkey/login/begin", { method: "POST", body: b }),
  passkeyLoginFinish: (flow, body) =>
    apiFetch("/api/passkey/login/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),

  totpEnroll: () => apiFetch("/api/totp/enroll", { method: "POST" }),
  totpConfirm: (b) => apiFetch("/api/totp/confirm", { method: "POST", body: b }),
  totpDisable: (b) => apiFetch("/api/totp/disable", { method: "POST", body: b }),
};
