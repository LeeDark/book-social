-- schema_v0_1.sql
-- Roles
CREATE TABLE roles (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    role_name VARCHAR(64) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id),
    UNIQUE KEY uq_roles_role_name (role_name)
) ENGINE=InnoDB;

-- Users
CREATE TABLE users (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    first_name  VARCHAR(100) NOT NULL,
    second_name VARCHAR(100) NULL,
    sur_name    VARCHAR(100) NULL,
    login       VARCHAR(64)  NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email       VARCHAR(254) NOT NULL,
    user_role_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_users_login (login),
    UNIQUE KEY uq_users_email (email),
    KEY idx_users_role (user_role_id),
    CONSTRAINT fk_users_role
       FOREIGN KEY (user_role_id) REFERENCES roles(id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT
) ENGINE=InnoDB;

-- Authors
CREATE TABLE authors (
     id INT UNSIGNED NOT NULL AUTO_INCREMENT,
     first_name  VARCHAR(100) NOT NULL,
     second_name VARCHAR(100) NULL,
     sur_name    VARCHAR(100) NULL,
     description TEXT NULL,
     PRIMARY KEY (id),
     KEY idx_authors_name (sur_name, first_name)
) ENGINE=InnoDB;

-- Genres
CREATE TABLE genres (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_genres_name (name)
) ENGINE=InnoDB;

-- Books (v0.1: прямые FK на Author и Genre)
CREATE TABLE books (
       id INT UNSIGNED NOT NULL AUTO_INCREMENT,
       title VARCHAR(255) NOT NULL,
       description TEXT NULL,

       book_author_id INT UNSIGNED NULL,
       book_genre_id  INT UNSIGNED NULL,

       PRIMARY KEY (id),
       KEY idx_books_author (book_author_id),
       KEY idx_books_genre  (book_genre_id),

       CONSTRAINT fk_books_author
           FOREIGN KEY (book_author_id) REFERENCES authors(id)
               ON UPDATE CASCADE
               ON DELETE SET NULL,

       CONSTRAINT fk_books_genre
           FOREIGN KEY (book_genre_id) REFERENCES genres(id)
               ON UPDATE CASCADE
               ON DELETE SET NULL
) ENGINE=InnoDB;

-- Cover: пока неизвестно, что хранить (url? blob? path?).
-- Если 1 обложка на книгу:
-- CREATE TABLE covers (...) with UNIQUE(book_id)

-- Shelves
CREATE TABLE shelves (
     id INT UNSIGNED NOT NULL AUTO_INCREMENT,
     name VARCHAR(120) NOT NULL,
     description TEXT NULL,
     PRIMARY KEY (id),
     UNIQUE KEY uq_shelves_name (name)
) ENGINE=InnoDB;

-- Tags
CREATE TABLE tags (
      id INT UNSIGNED NOT NULL AUTO_INCREMENT,
      name VARCHAR(120) NOT NULL,
      description TEXT NULL,
      PRIMARY KEY (id),
      UNIQUE KEY uq_tags_name (name)
) ENGINE=InnoDB;

-- Library (v0.1: shelf + book + (optional) tag)
CREATE TABLE library (
     id INT UNSIGNED NOT NULL AUTO_INCREMENT,
     library_shelf_id INT UNSIGNED NOT NULL,
     library_book_id  INT UNSIGNED NOT NULL,
     library_tag_id   INT UNSIGNED NULL,

     PRIMARY KEY (id),

     KEY idx_library_shelf (library_shelf_id),
     KEY idx_library_book  (library_book_id),
     KEY idx_library_tag   (library_tag_id),

    -- чтобы не плодить одинаковые записи
     UNIQUE KEY uq_library_triplet (library_shelf_id, library_book_id, library_tag_id),

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
) ENGINE=InnoDB;
