-- Active knowledge gaps detected by the meta-memory system
CREATE TABLE IF NOT EXISTS knowledge_gaps (
    id              INTEGER  PRIMARY KEY AUTOINCREMENT,
    topic           TEXT     NOT NULL,
    topic_embedding BLOB,
    gap_type        TEXT     NOT NULL
                             CHECK (gap_type IN (
                                 'sparse_beliefs',
                                 'low_confidence',
                                 'expired',
                                 'contradicted'
                             )),
    severity        REAL     NOT NULL DEFAULT 0.5
                             CHECK (severity >= 0.0 AND severity <= 1.0),
    first_detected  DATETIME NOT NULL DEFAULT (datetime('now')),
    last_seen       DATETIME NOT NULL DEFAULT (datetime('now')),
    filled_at       DATETIME,
    is_active       INTEGER  NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1))
);

CREATE INDEX IF NOT EXISTS idx_gaps_active_severity
    ON knowledge_gaps (severity DESC) WHERE is_active = 1;
CREATE INDEX IF NOT EXISTS idx_gaps_topic
    ON knowledge_gaps (topic) WHERE is_active = 1;
