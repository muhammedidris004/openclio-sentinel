-- Directed causal edges between beliefs (from REMI architecture)
CREATE TABLE IF NOT EXISTS belief_causal_edges (
    id              INTEGER  PRIMARY KEY AUTOINCREMENT,
    from_belief_id  INTEGER  NOT NULL REFERENCES beliefs(id) ON DELETE CASCADE,
    causal_type     TEXT     NOT NULL
                             CHECK (causal_type IN (
                                 'caused_by',
                                 'enables',
                                 'prevents',
                                 'implies',
                                 'weakens_if',
                                 'strengthens_if'
                             )),
    to_belief_id    INTEGER  NOT NULL REFERENCES beliefs(id) ON DELETE CASCADE,
    edge_strength   REAL     NOT NULL DEFAULT 0.7
                             CHECK (edge_strength > 0.0 AND edge_strength <= 1.0),
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE (from_belief_id, causal_type, to_belief_id)
);

CREATE INDEX IF NOT EXISTS idx_causal_edges_from
    ON belief_causal_edges (from_belief_id);
CREATE INDEX IF NOT EXISTS idx_causal_edges_to
    ON belief_causal_edges (to_belief_id);
CREATE INDEX IF NOT EXISTS idx_causal_edges_type
    ON belief_causal_edges (causal_type);
