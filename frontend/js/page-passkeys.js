// page-passkeys.js — drives the public passkeys explainer page.
//
// Two dynamic bits, both optional-feeling by design (the page reads fine even
// if the API is briefly unreachable):
//   1. The "this deployment right now" panel: fills in the hostname the
//      browser is on, which is exactly the Relying Party ID your passkeys
//      here are bound to.
//   2. The CSP transparency table: fetches the public, sanitised feed of
//      recent Content-Security-Policy violation reports and renders it —
//      with textContent only, since every field is attacker-suppliable text.

import { qs, highlightNav, renderAuthNav } from "./common.js";
import { Api } from "./api.js";
import { supported } from "./webauthn.js";

function fmtWhen(iso) {
  const d = new Date(iso);
  return Number.isFinite(d.getTime()) ? d.toLocaleString() : iso;
}

function fillDeploymentPanel() {
  const rpEl = qs("#rpIdNow");
  if (rpEl) rpEl.textContent = location.hostname;
  const supEl = qs("#passkeySupport");
  if (supEl) {
    supEl.textContent = supported()
      ? "This browser supports passkeys."
      : "This browser does not appear to support passkeys.";
  }
}

function renderCSPRow(r) {
  const tr = document.createElement("tr");
  const cells = [
    fmtWhen(r.created_at),
    r.effective_directive || r.violated_directive || "",
    r.blocked_uri || "",
    r.disposition || "",
  ];
  for (const text of cells) {
    const td = document.createElement("td");
    td.textContent = text; // never innerHTML: these strings come from reports
    tr.appendChild(td);
  }
  return tr;
}

async function loadCSPFeed() {
  const tbody = qs("#cspRows");
  const status = qs("#cspStatus");
  if (!tbody) return;
  try {
    const res = await Api.cspRecent(50);
    const rows = (res && res.reports) || [];
    tbody.textContent = "";
    if (rows.length === 0) {
      if (status) status.textContent = "No violations reported yet — a clean sheet.";
      return;
    }
    for (const r of rows) tbody.appendChild(renderCSPRow(r));
    if (status) status.textContent = "Showing the " + rows.length + " most recent reports.";
  } catch (err) {
    if (status) status.textContent = "Could not load the report feed right now.";
  }
}

async function main() {
  highlightNav();
  await renderAuthNav();
  fillDeploymentPanel();
  const reloadBtn = qs("#cspReload");
  if (reloadBtn) {
    reloadBtn.addEventListener("click", () => {
      reloadBtn.disabled = true;
      void loadCSPFeed().finally(() => { reloadBtn.disabled = false; });
    });
  }
  await loadCSPFeed();
}

void main();
