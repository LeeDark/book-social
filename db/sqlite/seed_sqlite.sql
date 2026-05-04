-- seed_sqlite.sql
-- SQLite version

PRAGMA foreign_keys = ON;

-- References
-- Genres
INSERT INTO genres(id, name, description)
VALUES (1, 'Romance', 'Stories about love, relationships, and emotions.'),
       (2, 'Fantasy', 'Books with magic, supernatural elements, or imaginary worlds.'),
       (3, 'Detective', 'Stories focused on solving a crime or a puzzle.');

-- Data
-- Authors
INSERT INTO authors(id, first_name, second_name, sur_name, description)
VALUES (1, 'Agatha', '', 'Christie',
        'Known as the "Queen of Crime" she is the best-selling novelist of all time. Her stories often feature famous
   detectives like Hercule Poirot and Miss Marple. Her plots are celebrated for their clever "whodunit" puzzles and unexpected plot twists.'),
       (2, 'Arthur', 'Conan', 'Doyle',
        'The creator of the world''s most famous consulting detective, Sherlock Holmes, and his loyal friend Dr. Watson.
   His stories are based on logical reasoning, forensic science, and brilliant observation, mostly set in Victorian
   London.');

-- Books
INSERT INTO books(title, description, book_author_id, book_genre_id)
VALUES ('Murder on the Orient Express',
        'While traveling from Istanbul to London on the luxurious Orient Express, a wealthy American businessman is found murdered in his locked compartment. A snowdrift stops the train in Yugoslavia, leaving the detective Hercule Poirot to identify the killer among a group of diverse passengers, all of whom have secrets to hide. It is famous for having one of the most ingenious endings in detective fiction.',
        1, 3),
       ('The Hound of the Baskervilles',
        ' This is perhaps the most celebrated Sherlock Holmes novel. Holmes and Dr. Watson travel to Dartmoor to investigate the mysterious death of Sir Charles Baskerville. Locals believe a supernatural, spectral hound haunts the foggy moors, cursed to kill members of the Baskerville family. Holmes must use logic to determine if the threat is truly paranormal or a cold-blooded human plot.',
        2, 3);
