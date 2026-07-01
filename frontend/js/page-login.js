import { highlightNav, renderAuthNav, qs, showMsg, clearMsg } from "./common.js";
import { Api, setToken } from "./api.js";
import { loginPasskey, supported } from "./webauthn.js";

async function main() {
  highlightNav();
  await renderAuthNav();

  const form = qs("#login-form");
  const msg = qs("#msg");
  const btn = qs("#submit");
  const passkeyBtn = qs("#passkey-login");

  if (!supported() && passkeyBtn) {
    passkeyBtn.disabled = true;
    passkeyBtn.title = "This browser does not support passkeys";
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearMsg(msg);
    btn.disabled = true;

    const username = qs("#username").value.trim();
    const password = qs("#password").value;
    const totp = qs("#totp").value.trim();

    try {
      const res = await Api.login({ username, password, totp });
      setToken(res.token);
      showMsg(msg, "Signed in. Redirecting…", "ok");
      setTimeout(() => (location.href = "/activity"), 500);
    } catch (err) {
      showMsg(msg, err.message || "Login failed", "error");
      btn.disabled = false;
    }
  });

  if (passkeyBtn) {
    passkeyBtn.addEventListener("click", async () => {
      clearMsg(msg);
      const username = qs("#username").value.trim();
      if (!username) {
        showMsg(msg, "Enter your username first, then use your passkey.", "info");
        return;
      }
      passkeyBtn.disabled = true;
      try {
        const res = await loginPasskey(username);
        setToken(res.token);
        showMsg(msg, "Signed in with passkey. Redirecting…", "ok");
        setTimeout(() => (location.href = "/activity"), 500);
      } catch (err) {
        showMsg(msg, err.message || "Passkey login failed", "error");
        passkeyBtn.disabled = false;
      }
    });
  }
}

main();
