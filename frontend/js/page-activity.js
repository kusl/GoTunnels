import { highlightNav, renderAuthNav, requireAuth, qs, showMsg } from "./common.js";
import { Api } from "./api.js";

function fmtTime(iso) {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function td(text, cls) {
  const cell = document.createElement("td");
  if (cls) cell.className = cls;
  cell.textContent = text;
  return cell;
}

async function main() {
  highlightNav();
  await renderAuthNav();
  const user = await requireAuth();
  if (!user) return;

  const msg = qs("#msg");
  const tbody = qs("#activity-body");

  try {
    const res = await Api.activity();
    const rows = (res && res.activity) || [];
    if (rows.length === 0) {
      showMsg(msg, "No activity recorded yet.", "info");
      return;
    }
    for (const r of rows) {
      const tr = document.createElement("tr");
      tr.appendChild(td(fmtTime(r.created_at)));
      tr.appendChild(td(r.event_type));
      tr.appendChild(td(r.auth_method || "—"));

      const outcome = document.createElement("td");
      const tag = document.createElement("span");
      tag.className = "tag " + (r.outcome === "success" ? "success" : "failure");
      tag.textContent = r.outcome;
      outcome.appendChild(tag);
      tr.appendChild(outcome);

      tr.appendChild(td(r.ip_hash || "—", "hash"));
      tbody.appendChild(tr);
    }
  } catch (err) {
    showMsg(msg, err.message || "Failed to load activity", "error");
  }
}

main();
