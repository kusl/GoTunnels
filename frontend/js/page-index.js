import { highlightNav, renderAuthNav, qs } from "./common.js";
import { Api } from "./api.js";

async function main() {
  highlightNav();
  const user = await renderAuthNav();

  const status = qs("#auth-status");
  if (status) {
    if (user) {
      status.textContent = "Signed in as " + user.username + ".";
    } else {
      status.textContent = "Not signed in.";
    }
  }

  try {
    const info = await Api.info();
    const el = qs("#instance-info");
    if (el) {
      el.textContent =
        info.service + " · instance " + info.instance_id + " · CSP: " + info.csp_mode +
        " · telemetry: " + (info.telemetry_on ? "on" : "off");
    }
  } catch {
    const el = qs("#instance-info");
    if (el) el.textContent = "API not reachable yet — is the stack up and config.json written?";
  }
}

main();
