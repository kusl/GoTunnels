// config.js — learns the API base URL at runtime.
//
// The API's public Cloudflare Quick Tunnel URL is random and only known after
// the stack starts, so it cannot be baked into these static files. Instead the
// startup script writes /config.json into this container, and every page reads
// it here before making API calls.

let cachedPromise;

export function loadConfig() {
  if (!cachedPromise) {
    cachedPromise = fetch("/config.json", { cache: "no-store" })
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("config load failed"))))
      .catch(() => ({ apiBase: "", instanceId: "", note: "config.json missing" }));
  }
  return cachedPromise;
}
