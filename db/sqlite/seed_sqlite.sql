-- seed_sqlite.sql
-- SQLite version

PRAGMA foreign_keys = ON;

-- References
-- Genres
INSERT INTO genres(id, name, slug, description)
VALUES (1, 'Adventure', 'adventure', 'Stories focused on journeys, exploration, danger, and exciting challenges.'),
       (2, 'Biography', 'biography', 'Books about the life of a real person.'),
       (3, 'Business', 'business', 'Books about companies, management, entrepreneurship, leadership, and markets.'),
       (4, 'Children''s Literature', 'childrens-literature', 'Books written for children, often with simple language, imagination, and clear themes.'),
       (5, 'Classic Literature', 'classic-literature', 'Influential works of fiction with lasting literary or cultural importance.'),
       (6, 'Crime', 'crime', 'Stories about criminal acts, investigations, justice, and the people involved.'),
       (7, 'Detective', 'detective','Stories focused on solving a crime or a puzzle.'),
       (8, 'Drama', 'drama', 'Stories written for performance or centered on serious conflict and character development.'),
       (9, 'Fantasy', 'fantasy', 'Books with magic, supernatural elements, or imaginary worlds.'),
       (10, 'Historical Fiction', 'historical-fiction', 'Fictional stories set in real historical periods or events.'),
       (11, 'Horror', 'horror', 'Stories designed to create fear, suspense, or a sense of dread.'),
       (12, 'Humor', 'humor', 'Books written to entertain through wit, comedy, satire, or absurd situations.'),
       (13, 'Memoir', 'memoir', 'Personal stories written from the author''s own life and experience.'),
       (14, 'Mystery', 'mystery', 'Stories centered on secrets, clues, and unanswered questions.'),
       (15, 'Nonfiction', 'nonfiction', 'Informative books based on real events, facts, research, or personal experience.'),
       (16, 'Philosophy', 'philosophy', 'Books exploring ideas about knowledge, ethics, reality, meaning, and human life.'),
       (17, 'Poetry', 'poetry', 'Books written in verse, often focused on rhythm, imagery, and emotional expression.'),
       (18, 'Romance', 'romance','Stories about love, relationships, and emotions.'),
       (19, 'Science Fiction', 'science-fiction', 'Stories based on futuristic technology, space exploration, time travel, or scientific ideas.'),
       (20, 'Self-Help', 'self-help', 'Books offering practical advice for personal growth, habits, relationships, or wellbeing.'),
       (21, 'Thriller', 'thriller', 'Fast-paced stories built around danger, tension, and high stakes.'),
       (22, 'Travel', 'travel', 'Books about places, journeys, cultures, and personal experiences while traveling.'),
       (23, 'Young Adult', 'young-adult', 'Books written primarily for teenage readers, often focused on identity, growth, and relationships.');

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
INSERT INTO books(title, slug, description, book_author_id, book_genre_id)
VALUES ('Murder on the Orient Express',
        'murder-on-the-orient-express',
        'While traveling from Istanbul to London on the luxurious Orient Express, a wealthy American businessman is found murdered in his locked compartment. A snowdrift stops the train in Yugoslavia, leaving the detective Hercule Poirot to identify the killer among a group of diverse passengers, all of whom have secrets to hide. It is famous for having one of the most ingenious endings in detective fiction.',
        1, 7),
       ('The Hound of the Baskervilles',
        'the-hound-of-the-baskervilles',
        'This is perhaps the most celebrated Sherlock Holmes novel. Holmes and Dr. Watson travel to Dartmoor to investigate the mysterious death of Sir Charles Baskerville. Locals believe a supernatural, spectral hound haunts the foggy moors, cursed to kill members of the Baskerville family. Holmes must use logic to determine if the threat is truly paranormal or a cold-blooded human plot.',
        2, 7);
