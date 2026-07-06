import { highlightNav, renderAuthNav, qs, showMsg, clearMsg } from "./common.js";
import { Api, setToken } from "./api.js";
import { signupPasskey, supported } from "./webauthn.js";

async function main() {
  highlightNav();
  await renderAuthNav();

  const form = qs("#signup-form");
  const msg = qs("#msg");
  const btn = qs("#submit");
  const passkeyBtn = qs("#passkey-signup");

  if (passkeyBtn && !supported()) {
    passkeyBtn.disabled = true;
    passkeyBtn.title = "This browser does not support passkeys";
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearMsg(msg);
    btn.disabled = true;

    const username = qs("#username").value.trim();
    const displayName = qs("#display_name").value.trim();
    const password = qs("#password").value;

    try {
      const res = await Api.signup({
        username,
        display_name: displayName,
        password,
      });
      setToken(res.token);
      showMsg(msg, "Account created. Redirecting…", "ok");
      setTimeout(() => (location.href = "/settings"), 600);
    } catch (err) {
      showMsg(msg, err.message || "Signup failed", "error");
      btn.disabled = false;
    }
  });

  // Passkey-first signup: same username/display-name fields, no password at
  // all. The account is only created after the authenticator produces a
  // credential, so cancelling the browser prompt creates nothing.
  if (passkeyBtn) {
    passkeyBtn.addEventListener("click", async () => {
      clearMsg(msg);
      const username = qs("#username").value.trim();
      const displayName = qs("#display_name").value.trim();
      if (!username) {
        showMsg(msg, "Pick a username first, then create your passkey.", "info");
        return;
      }
      passkeyBtn.disabled = true;
      try {
        const res = await signupPasskey(username, displayName);
        setToken(res.token);
        showMsg(msg, "Account created with a passkey. Redirecting…", "ok");
        setTimeout(() => (location.href = "/settings"), 600);
      } catch (err) {
        if (err && err.name === "NotAllowedError") {
          showMsg(msg, "Passkey prompt was cancelled — nothing was created.", "info");
        } else {
          showMsg(msg, err.message || "Passkey signup failed", "error");
        }
        passkeyBtn.disabled = false;
      }
    });
  }
}

main();
