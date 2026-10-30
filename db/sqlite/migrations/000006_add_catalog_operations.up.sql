-- SQLite version

-- Reference capability only. Administrator accounts are appointed explicitly.
INSERT INTO roles (role_name, is_admin)
VALUES ('admin', 1);

CREATE TABLE catalog_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_user_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    summary TEXT NOT NULL,
    occurred_at TEXT NOT NULL,

    CONSTRAINT ck_catalog_audit_entity_type CHECK (entity_type = 'book'),
    CONSTRAINT ck_catalog_audit_action CHECK (action IN ('create', 'correct')),
    CONSTRAINT fk_catalog_audit_actor
        FOREIGN KEY (actor_user_id) REFERENCES users(id)
            ON UPDATE RESTRICT
            ON DELETE RESTRICT,
    CONSTRAINT fk_catalog_audit_book
        FOREIGN KEY (entity_id) REFERENCES books(id)
            ON UPDATE RESTRICT
            ON DELETE RESTRICT
);
