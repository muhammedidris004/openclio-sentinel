-- Epistemic belief store: probabilistic claims with confidence and provenance
CREATE TABLE IF NOT EXISTS beliefs (
    id                  INTEGER  PRIMARY KEY AUTOINCREMENT,
    claim               TEXT     NOT NULL,
    claim_embedding     BLOB,
    confidence          REAL     NOT NULL DEFAULT 0.5
                                 CHECK (confidence >= 0.0 AND confidence <= 1.0),
    provenance          TEXT     NOT NULL
                                 CHECK (provenance IN (
                                     'explicit_user_statement',
                                     'implicit_user_signal',
                                     'cross_session_pattern',
                                     'agent_inference',
                                     'ambient_calendar',
                                     'ambient_file',
                                     'ambient_temporal'
                                 )),
    category            TEXT     NOT NULL
                                 CHECK (category IN (
                                     'preference',
                                     'fact_static',
                                     'fact_dynamic',
                                     'skill',
                                     'deadline',
                                     'project_status',
                                     'relationship',
                                     'observation'
                                 )),
    decay_rate          REAL     NOT NULL DEFAULT 0.005
                                 CHECK (decay_rate >= 0.0 AND decay_rate <= 1.0),
    valid_from          DATETIME NOT NULL DEFAULT (datetime('now')),
    valid_until         DATETIME,
    source_message_id   INTEGER  REFERENCES messages(id) ON DELETE SET NULL,
    is_active           INTEGER  NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    access_count        INTEGER  NOT NULL DEFAULT 0,
    last_accessed       DATETIME,
    created_at          DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at          DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_beliefs_active_confidence
    ON beliefs (confidence DESC) WHERE is_active = 1;
CREATE INDEX IF NOT EXISTS idx_beliefs_category_active
    ON beliefs (category, is_active);
CREATE INDEX IF NOT EXISTS idx_beliefs_valid_until
    ON beliefs (valid_until) WHERE valid_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_beliefs_updated_at
    ON beliefs (updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_beliefs_source_msg
    ON beliefs (source_message_id) WHERE source_message_id IS NOT NULL;
