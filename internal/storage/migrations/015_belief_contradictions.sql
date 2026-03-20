-- Tracks semantic contradictions between beliefs with resolution status
CREATE TABLE IF NOT EXISTS belief_contradictions (
    id                  INTEGER  PRIMARY KEY AUTOINCREMENT,
    belief_a_id         INTEGER  NOT NULL REFERENCES beliefs(id) ON DELETE CASCADE,
    belief_b_id         INTEGER  NOT NULL REFERENCES beliefs(id) ON DELETE CASCADE,
    contradiction_score REAL     NOT NULL
                                 CHECK (contradiction_score > 0.0 AND contradiction_score <= 1.0),
    resolution_status   TEXT     NOT NULL DEFAULT 'unresolved'
                                 CHECK (resolution_status IN (
                                     'unresolved',
                                     'resolved_bayesian',
                                     'resolved_temporal',
                                     'resolved_contextual',
                                     'acknowledged'
                                 )),
    resolved_at         DATETIME,
    created_at          DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE (belief_a_id, belief_b_id)
);

CREATE INDEX IF NOT EXISTS idx_contradictions_belief_a
    ON belief_contradictions (belief_a_id);
CREATE INDEX IF NOT EXISTS idx_contradictions_belief_b
    ON belief_contradictions (belief_b_id);
CREATE INDEX IF NOT EXISTS idx_contradictions_unresolved
    ON belief_contradictions (resolution_status, created_at DESC)
    WHERE resolution_status = 'unresolved';
