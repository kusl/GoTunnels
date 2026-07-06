// theme-boot.js — applies the saved color theme before first paint.
//
// This is deliberately a tiny CLASSIC script (not a module) loaded in <head>:
// classic scripts block parsing, so the data-theme attribute is set before the
// browser paints anything and there is no flash of the wrong theme. Everything
// else about theming (the picker UI, server sync) lives in common.js; this
// file only reads localStorage and stamps <html data-theme="...">.
//
// Keep the key and the value list in sync with THEMES in common.js.
(function () {
  "use strict";
  var KEY = "gotunnels_theme";
  var VALID = {
    system: true,
    dark: true,
    light: true,
    "solarized-dark": true,
    "solarized-light": true,
  };
  var theme = "system";
  try {
    var saved = localStorage.getItem(KEY);
    if (saved && VALID[saved] === true) theme = saved;
  } catch (e) {
    // localStorage can throw (privacy modes); the default is fine.
  }
  document.documentElement.setAttribute("data-theme", theme);
})();
