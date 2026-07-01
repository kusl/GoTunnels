// common.js — shared helpers used by every page. Importing it also installs
// the CSP violation reporter (via ./csp.js).

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
