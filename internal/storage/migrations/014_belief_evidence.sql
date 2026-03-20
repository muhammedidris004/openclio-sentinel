-- Full audit trail of every Bayesian update applied to each belief
CREATE TABLE IF NOT EXISTS belief_evidence (
    id                   INTEGER  PRIMARY KEY AUTOINCREMENT,
    belief_id            INTEGER  NOT NULL REFERENCES beliefs(id) ON DELETE CASCADE,
    direction            TEXT     NOT NULL CHECK (direction IN ('supporting', 'contradicting')),
    evidence_text        TEXT     NOT NULL,
    evidence_strength    REAL     NOT NULL
                                  CHECK (evidence_strength > 0.0 AND evidence_strength <= 1.0),
    provenance           TEXT     NOT NULL,
    prior_confidence     REAL     NOT NULL CHECK (prior_confidence >= 0.0 AND prior_confidence <= 1.0),
    posterior_confidence REAL     NOT NULL CHECK (posterior_confidence >= 0.0 AND posterior_confidence <= 1.0),
    source_message_id    INTEGER  REFERENCES messages(id) ON DELETE SET NULL,
    created_at           DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_evidence_belief_id
    ON belief_evidence (belief_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_evidence_direction
    ON belief_evidence (direction);
