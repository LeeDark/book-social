-- 000003_add_user_role_and_sessions.up.sql
-- SQLite version

-- This is reference data required by registration, not a demo account.
INSERT INTO roles (role_name, is_admin)
VALUES ('user', 0);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash BLOB NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,

    CONSTRAINT uq_sessions_token_hash UNIQUE (token_hash),
    CONSTRAINT ck_sessions_expiry_after_created CHECK (expires_at > created_at),

    CONSTRAINT fk_sessions_user
        FOREIGN KEY (user_id) REFERENCES users(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);

CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
