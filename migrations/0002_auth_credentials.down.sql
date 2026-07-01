-- 0002_auth_credentials.down.sql
DROP TABLE IF EXISTS totp_recovery_codes;
DROP TABLE IF EXISTS totp_secrets;
DROP TABLE IF EXISTS webauthn_flows;
DROP TABLE IF EXISTS webauthn_credentials;
DROP TABLE IF EXISTS password_credentials;
