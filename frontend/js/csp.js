// csp.js — reports Content-Security-Policy violations to the API.
//
// The CSP header carries no report-uri/report-to, because the API's public URL
// is only known at runtime. Instead we listen for the in-page violation event
// (which fires for both report-only and enforcing policies) and POST a compact
// JSON body to the API's /api/csp-reports endpoint. Importing this module for
// its side effect installs the listener.

import { loadConfig } from "./config.js";

async function report(e) {
  try {
    const cfg = await loadConfig();
    const base = cfg.apiBase || "";
    const body = {
      documentURI: e.documentURI || location.href,
      referrer: e.referrer || "",
      blockedURI: e.blockedURI || "",
      violatedDirective: e.violatedDirective || "",
      effectiveDirective: e.effectiveDirective || e.violatedDirective || "",
      originalPolicy: e.originalPolicy || "",
      disposition: e.disposition || "",
      sourceFile: e.sourceFile || "",
      lineNumber: e.lineNumber || 0,
      columnNumber: e.columnNumber || 0,
      statusCode: e.statusCode || 0,
      sample: e.sample || "",
    };
    // keepalive lets the report survive an in-flight navigation.
    fetch(base + "/api/csp-reports", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      mode: "cors",
      keepalive: true,
    }).catch(() => {});
  } catch {
    // Reporting must never break the page.
  }
}

document.addEventListener("securitypolicyviolation", report);
