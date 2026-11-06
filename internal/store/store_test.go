package store

import (
	"bytes"
	"encoding/json/v2"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func TestCredentialRoundTrip(t *testing.T) {
	in := webauthn.Credential{
		ID:                []byte{1, 2, 3, 4},
		PublicKey:         []byte{5, 6, 7, 8},
		AttestationType:   "basic_self",
		AttestationFormat: "packed",
		Transport:         []protocol.AuthenticatorTransport{protocol.Internal, protocol.Hybrid},
		Flags:             webauthn.CredentialFlags{UserPresent: true, UserVerified: true, BackupEligible: true},
		Authenticator: webauthn.Authenticator{
			AAGUID:       bytes.Repeat([]byte{9}, 16),
			SignCount:    42,
			CloneWarning: true,
		},
	}
	blob, err := marshalCredential(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := unmarshalCredential(blob)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	switch {
	case !bytes.Equal(out.ID, in.ID), !bytes.Equal(out.PublicKey, in.PublicKey):
		t.Fatalf("key material changed: %+v", out)
	case out.AttestationType != in.AttestationType, out.AttestationFormat != in.AttestationFormat,
		len(out.Transport) != 2, out.Transport[1] != protocol.Hybrid:
		t.Fatalf("metadata changed: %+v", out)
	case out.Flags != in.Flags:
		t.Fatalf("flags changed: %+v", out.Flags)
	case !bytes.Equal(out.Authenticator.AAGUID, in.Authenticator.AAGUID),
		out.Authenticator.SignCount != 42, !out.Authenticator.CloneWarning:
		t.Fatalf("authenticator changed: %+v", out.Authenticator)
	}
	again, err := marshalCredential(&out)
	if err != nil || !bytes.Equal(again, blob) {
		t.Fatalf("re-encoding is not stable:\n%s\n%s", blob, again)
	}
}

func TestActivityDetailEncoding(t *testing.T) {
	blob, err := marshalActivityDetail(nil)
	if err != nil || string(blob) != "{}" {
		t.Fatalf("nil detail = %q, %v; want {}", blob, err)
	}
	blob, err = marshalActivityDetail(map[string]any{"z": 1, "error": "bad \xff byte"})
	if err != nil {
		t.Fatalf("invalid UTF-8 must not fail the audit write: %v", err)
	}
	if want := "{\"error\":\"bad \ufffd byte\",\"z\":1}"; string(blob) != want {
		t.Fatalf("detail = %s, want %s", blob, want)
	}
}

func TestActivityJSONShape(t *testing.T) {
	blob, err := json.Marshal(Activity{ID: 1, Username: "a"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(blob, []byte(`"user_id"`)) || bytes.Contains(blob, []byte(`"detail"`)) {
		t.Fatalf("zero user_id/detail must be omitted: %s", blob)
	}
	uid := "u"
	blob, _ = json.Marshal(Activity{UserID: &uid, Detail: []byte(`{}`)})
	if !bytes.Contains(blob, []byte(`"user_id":"u"`)) || !bytes.Contains(blob, []byte(`"detail":{}`)) {
		t.Fatalf("present user_id/detail must be kept: %s", blob)
	}
}
