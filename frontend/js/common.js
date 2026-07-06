// common.js — shared helpers used by every page. Importing it also installs
// the CSP violation reporter (via ./csp.js) and the theme picker in the top
// bar (theme-boot.js already applied the saved theme pre-paint; this module
// adds the UI and, when signed in, syncs the choice to the account).

import "./csp.js";
import { Api, getToken } from "./api.js";

export function qs(sel, root = document) {
  return root.querySelector(sel);
}

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

export async function currentUser() {
  if (!getToken()) return null;
  try {
    return await Api.me();
  } catch {
    return null;
  }
}

export async function requireAuth() {
  const u = await currentUser();
  if (!u) {
    location.href = "/login";
    return null;
  }
  return u;
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
