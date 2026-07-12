-- 000001_create_v0_1_schema.down.sql
-- SQLite version

PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS library;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS shelves;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

PRAGMA foreign_keys = ON;
