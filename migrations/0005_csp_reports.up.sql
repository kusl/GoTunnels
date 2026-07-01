-- 0005_csp_reports.up.sql
-- Content-Security-Policy violation reports. The API accepts several report
-- shapes (legacy application/csp-report, the Reporting API array, and a custom
-- JSON body posted by the in-page violation listener) and normalises them into
-- this single table. The full original payload is retained in raw.
CREATE TABLE csp_reports (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    document_uri        text NOT NULL DEFAULT '',
    referrer            text NOT NULL DEFAULT '',
    blocked_uri         text NOT NULL DEFAULT '',
    violated_directive  text NOT NULL DEFAULT '',
    effective_directive text NOT NULL DEFAULT '',
    original_policy     text NOT NULL DEFAULT '',
    disposition         text NOT NULL DEFAULT '',
    source_file         text NOT NULL DEFAULT '',
    line_number         integer NOT NULL DEFAULT 0,
    column_number       integer NOT NULL DEFAULT 0,
    status_code         integer NOT NULL DEFAULT 0,
    script_sample       text NOT NULL DEFAULT '',
    ip_hash             text NOT NULL DEFAULT '',
    user_agent          text NOT NULL DEFAULT '',
    raw                 jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX csp_reports_created_at_idx
    ON csp_reports (created_at DESC);
CREATE INDEX csp_reports_violated_directive_idx
    ON csp_reports (violated_directive);
