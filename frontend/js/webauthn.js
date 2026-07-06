// webauthn.js — passkey (WebAuthn) ceremonies.
//
// The browser needs ArrayBuffers where the server sends base64url strings, and
// vice-versa. These helpers do that conversion around navigator.credentials.

import { Api } from "./api.js";

function b64urlToBuf(s) {
  s = String(s).replace(/-/g, "+").replace(/_/g, "/");
  const pad = s.length % 4;
  if (pad) s += "=".repeat(4 - pad);
  const bin = atob(s);
  const buf = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) buf[i] = bin.charCodeAt(i);
  return buf.buffer;
}

function bufToB64url(buf) {
  const bytes = new Uint8Array(buf);
  let bin = "";
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
  return btoa(bin).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function prepCreation(pk) {
  pk.challenge = b64urlToBuf(pk.challenge);
  pk.user.id = b64urlToBuf(pk.user.id);
  if (Array.isArray(pk.excludeCredentials)) {
    pk.excludeCredentials = pk.excludeCredentials.map((c) => ({ ...c, id: b64urlToBuf(c.id) }));
  }
  return pk;
}

function prepRequest(pk) {
  pk.challenge = b64urlToBuf(pk.challenge);
  if (Array.isArray(pk.allowCredentials)) {
    pk.allowCredentials = pk.allowCredentials.map((c) => ({ ...c, id: b64urlToBuf(c.id) }));
  }
  return pk;
}

function encodeAttestation(cred) {
  const resp = cred.response;
  const out = {
    id: cred.id,
    rawId: bufToB64url(cred.rawId),
    type: cred.type,
    clientExtensionResults: cred.getClientExtensionResults ? cred.getClientExtensionResults() : {},
    response: {
      attestationObject: bufToB64url(resp.attestationObject),
      clientDataJSON: bufToB64url(resp.clientDataJSON),
    },
  };
  if (resp.getTransports) {
    try {
      out.response.transports = resp.getTransports();
    } catch {
      /* optional */
    }
  }
  return out;
}

function encodeAssertion(cred) {
  const resp = cred.response;
  return {
    id: cred.id,
    rawId: bufToB64url(cred.rawId),
    type: cred.type,
    clientExtensionResults: cred.getClientExtensionResults ? cred.getClientExtensionResults() : {},
    response: {
      authenticatorData: bufToB64url(resp.authenticatorData),
      clientDataJSON: bufToB64url(resp.clientDataJSON),
      signature: bufToB64url(resp.signature),
      userHandle: resp.userHandle ? bufToB64url(resp.userHandle) : null,
    },
  };
}

export function supported() {
  return !!(window.PublicKeyCredential && navigator.credentials && navigator.credentials.create);
}

// registerPasskey adds a passkey to the currently authenticated user.
export async function registerPasskey() {
  const begin = await Api.passkeyRegisterBegin();
  const publicKey = prepCreation(begin.options.publicKey);
  const cred = await navigator.credentials.create({ publicKey });
  return Api.passkeyRegisterFinish(begin.flow_id, encodeAttestation(cred));
}

// loginPasskey authenticates a named user with a passkey and returns the
// session response (including the Bearer token).
export async function loginPasskey(username) {
  const begin = await Api.passkeyLoginBegin({ username });
  const publicKey = prepRequest(begin.options.publicKey);
  const cred = await navigator.credentials.get({ publicKey });
  return Api.passkeyLoginFinish(begin.flow_id, encodeAssertion(cred));
}

// signupPasskey creates a brand-new account whose only credential is a
// passkey (no password anywhere), and returns the session response. The
// account only comes into existence server-side once the authenticator has
// produced a credential; cancelling the browser prompt creates nothing.
export async function signupPasskey(username, displayName) {
  const begin = await Api.passkeySignupBegin({ username, display_name: displayName });
  const publicKey = prepCreation(begin.options.publicKey);
  const cred = await navigator.credentials.create({ publicKey });
  return Api.passkeySignupFinish(begin.flow_id, encodeAttestation(cred));
}
