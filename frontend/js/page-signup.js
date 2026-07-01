import { highlightNav, renderAuthNav, qs, showMsg, clearMsg } from "./common.js";
import { Api, setToken } from "./api.js";

async function main() {
  highlightNav();
  await renderAuthNav();

  const form = qs("#signup-form");
  const msg = qs("#msg");
  const btn = qs("#submit");

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
}

main();
