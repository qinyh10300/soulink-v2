CREATE TABLE IF NOT EXISTS interface_gate_windows (
    gate_key VARCHAR(64) PRIMARY KEY,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (ends_at > starts_at)
);