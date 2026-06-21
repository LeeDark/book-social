-- schema_v0_1_sqlite.sql
-- SQLite version

PRAGMA foreign_keys = ON;

-- Roles
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_name TEXT NOT NULL,
    is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1)),
    CONSTRAINT uq_roles_role_name UNIQUE (role_name)
);

-- Users
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name     TEXT NOT NULL,
    second_name    TEXT NULL,
    sur_name       TEXT NULL,
    login          TEXT NOT NULL,
    password_hash  TEXT NOT NULL,
    email          TEXT NOT NULL,
    user_role_id   INTEGER NOT NULL,

    CONSTRAINT uq_users_login UNIQUE (login),
    CONSTRAINT uq_users_email UNIQUE (email),

    CONSTRAINT fk_users_role
       FOREIGN KEY (user_role_id) REFERENCES roles(id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT
);

CREATE INDEX idx_users_role ON users(user_role_id);

-- Authors
CREATE TABLE authors (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     first_name  TEXT NOT NULL,
     second_name TEXT NULL,
     sur_name    TEXT NULL,
     slug TEXT NOT NULL UNIQUE,
     description TEXT NULL
);

CREATE INDEX idx_authors_name ON authors(sur_name, first_name);

-- Genres
CREATE TABLE genres (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT NULL,
    CONSTRAINT uq_genres_name UNIQUE (name)
);

-- Books (v0.1: прямые FK на Author и Genre)
CREATE TABLE books (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       title TEXT NOT NULL,
       slug TEXT NOT NULL UNIQUE,
       description TEXT NULL,

       book_author_id INTEGER NULL,
       book_genre_id  INTEGER NULL,

       CONSTRAINT fk_books_author
           FOREIGN KEY (book_author_id) REFERENCES authors(id)
               ON UPDATE CASCADE
               ON DELETE SET NULL,

       CONSTRAINT fk_books_genre
           FOREIGN KEY (book_genre_id) REFERENCES genres(id)
               ON UPDATE CASCADE
               ON DELETE SET NULL
);

CREATE INDEX idx_books_author ON books(book_author_id);
CREATE INDEX idx_books_genre  ON books(book_genre_id);

-- Cover: пока неизвестно, что хранить (url? blob? path?).
-- Если 1 обложка на книгу:
-- CREATE TABLE covers (...) with UNIQUE(book_id)

-- Shelves
CREATE TABLE shelves (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     name TEXT NOT NULL,
     description TEXT NULL,
     CONSTRAINT uq_shelves_name UNIQUE (name)
);

-- Tags
CREATE TABLE tags (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT NULL,
      CONSTRAINT uq_tags_name UNIQUE (name)
);

-- Library (v0.1: shelf + book + (optional) tag)
CREATE TABLE library (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     library_shelf_id INTEGER NOT NULL,
     library_book_id  INTEGER NOT NULL,
     library_tag_id   INTEGER NULL,

     -- чтобы не плодить одинаковые записи
     CONSTRAINT uq_library_triplet UNIQUE (library_shelf_id, library_book_id, library_tag_id),

     CONSTRAINT fk_library_shelf
         FOREIGN KEY (library_shelf_id) REFERENCES shelves(id)
             ON UPDATE CASCADE
             ON DELETE RESTRICT,

     CONSTRAINT fk_library_book
         FOREIGN KEY (library_book_id) REFERENCES books(id)
             ON UPDATE CASCADE
             ON DELETE RESTRICT,

     CONSTRAINT fk_library_tag
         FOREIGN KEY (library_tag_id) REFERENCES tags(id)
             ON UPDATE CASCADE
             ON DELETE SET NULL
);

CREATE INDEX idx_library_shelf ON library(library_shelf_id);
CREATE INDEX idx_library_book  ON library(library_book_id);
CREATE INDEX idx_library_tag   ON library(library_tag_id);
