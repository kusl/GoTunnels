import { highlightNav, renderAuthNav, requireAuth, qs, showMsg, clearMsg } from "./common.js";
import { Api, clearToken } from "./api.js";
import { registerPasskey, supported } from "./webauthn.js";

let currentProfile = null;

function renderProfile(u) {
  currentProfile = u;
  qs("#p-username").textContent = u.username;
  qs("#p-display").textContent = u.display_name || "—";
  qs("#p-roles").textContent = (u.roles || []).join(", ") || "—";
  qs("#p-passkeys").textContent = String(u.passkeys);
  qs("#p-totp").textContent = u.totp_enabled ? "enabled" : "disabled";

  qs("#totp-enroll-section").classList.toggle("hidden", u.totp_enabled);
  qs("#totp-disable-section").classList.toggle("hidden", !u.totp_enabled);
}

async function refresh() {
  const u = await Api.me();
  renderProfile(u);
}

async function main() {
  highlightNav();
  await renderAuthNav();
  const user = await requireAuth();
  if (!user) return;
  renderProfile(user);

  const msg = qs("#msg");

  // ---- passkey registration ----
  const pkBtn = qs("#register-passkey");
  if (!supported()) {
    pkBtn.disabled = true;
    pkBtn.title = "This browser does not support passkeys";
  }
  pkBtn.addEventListener("click", async () => {
    clearMsg(msg);
    pkBtn.disabled = true;
    try {
      await registerPasskey();
      showMsg(msg, "Passkey registered.", "ok");
      await refresh();
    } catch (err) {
      showMsg(msg, err.message || "Passkey registration failed", "error");
    } finally {
      pkBtn.disabled = false;
    }
  });

  // ---- TOTP enroll ----
  const enrollBtn = qs("#totp-enroll");
  const enrollResult = qs("#totp-enroll-result");
  enrollBtn.addEventListener("click", async () => {
    clearMsg(msg);
    enrollBtn.disabled = true;
    try {
      const res = await Api.totpEnroll();
      qs("#totp-secret").textContent = res.secret;
      qs("#totp-url").textContent = res.otpauth_url;
      qs("#totp-recovery").textContent = (res.recovery_codes || []).join("\n");
      enrollResult.classList.remove("hidden");
      showMsg(msg, "Scan or enter the secret in your authenticator, then confirm below.", "info");
    } catch (err) {
      showMsg(msg, err.message || "Enrollment failed", "error");
    } finally {
      enrollBtn.disabled = false;
    }
  });

  const confirmForm = qs("#totp-confirm-form");
  confirmForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearMsg(msg);
    try {
      await Api.totpConfirm({ code: qs("#totp-confirm-code").value.trim() });
      showMsg(msg, "TOTP enabled.", "ok");
      enrollResult.classList.add("hidden");
      await refresh();
    } catch (err) {
      showMsg(msg, err.message || "Confirmation failed", "error");
    }
  });

  // ---- TOTP disable ----
  const disableForm = qs("#totp-disable-form");
  disableForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearMsg(msg);
    try {
      await Api.totpDisable({ code: qs("#totp-disable-code").value.trim() });
      showMsg(msg, "TOTP disabled.", "ok");
      await refresh();
    } catch (err) {
      showMsg(msg, err.message || "Disable failed", "error");
    }
  });

  // ---- logout (this browser) ----
  qs("#logout").addEventListener("click", async () => {
    try {
      await Api.logout();
    } catch {
      /* ignore */
    }
    clearToken();
    location.href = "/login";
  });

  // ---- logout everywhere (all devices) ----
  const logoutAllBtn = qs("#logout-all");
  if (logoutAllBtn) {
    logoutAllBtn.addEventListener("click", async () => {
      logoutAllBtn.disabled = true;
      clearMsg(msg);
      try {
        const res = await Api.logoutAll();
        const n = (res && res.sessions_revoked) || 0;
        showMsg(msg, `Signed out of ${n} session${n === 1 ? "" : "s"} everywhere. Redirecting…`, "ok");
      } catch {
        // Even if the call fails, drop the local token so this browser is
        // signed out; the redirect below sends the user to log in again.
      }
      clearToken();
      setTimeout(() => (location.href = "/login"), 600);
    });
  }
}

main();
