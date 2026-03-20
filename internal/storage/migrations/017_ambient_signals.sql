-- Short-lived ambient observations. Processed into belief updates then expired.
CREATE TABLE IF NOT EXISTS ambient_signals (
    id              INTEGER  PRIMARY KEY AUTOINCREMENT,
    signal_type     TEXT     NOT NULL
                             CHECK (signal_type IN (
                                 'calendar_event',
                                 'file_created',
                                 'file_modified',
                                 'file_deleted',
                                 'file_renamed',
                                 'time_pattern'
                             )),
    payload_json    TEXT     NOT NULL,
    processed       INTEGER  NOT NULL DEFAULT 0 CHECK (processed IN (0, 1)),
    processed_at    DATETIME,
    expires_at      DATETIME NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_ambient_unprocessed
    ON ambient_signals (created_at ASC) WHERE processed = 0;
CREATE INDEX IF NOT EXISTS idx_ambient_expiry
    ON ambient_signals (expires_at) WHERE processed = 1;
