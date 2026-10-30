# Curated Demo Catalog Provenance

## Dataset and Responsibilities

The development seed contains 109 books, 24 authors, 23 genres, 109 author links, 109 genre links,
and no covers. SQLite and PostgreSQL contain the same catalog metadata. Books receive IDs in
bytewise slug order on a fresh database; PostgreSQL uses explicit `COLLATE "C"`, matching SQLite
regardless of the database locale. Author and genre IDs are explicit. Apply seed once after migrations to a
disposable database, not as a repeatable data migration or a production import. This change does
not renumber existing databases; do not rerun the seed to repair an existing catalog.

Ordinary tests use minimal fixtures in `internal/testutil` or their own scenario data. Only
dedicated
development-seed acceptance tests load the full seed. The curated demo catalog is the development
subset, not an authoritative bibliography.

## Metadata Source and Limitations

The checked-in SQL is the immediate source for every record below. Git history shows the original
SQLite seed in commit `f9c1732`, its expanded dataset in `55c1cae`, and normalized relationships in
`e36116c`. These commits establish repository lineage, not external provenance or authorship.
No original external source or reuse basis for the descriptions was found in that history.
Author biographies and genre descriptions have the same undocumented external provenance.

**Unknown** below means original external metadata source and description authorship/reuse basis
are unverified. Titles, summaries, author attribution, and genre choices are legacy demonstration
metadata, not independently verified edition records. Some collections and translated titles may
overlap; there is no ISBN, edition, or language disambiguation. Before reusing descriptions outside
this demo, verify their source or replace them with reviewed original text.

## Record Inventory

Each row identifies a seeded book by its stable slug. The immediate metadata source is the matching
row in `db/sqlite/seed.sql` and `db/postgresql/seed.sql`; external-source status is explicit.
All rows share the limitations above.

| Book slug                                               | External metadata source | Cover URL | Cover-use basis | Limitations                                 |
|---------------------------------------------------------|--------------------------|-----------|-----------------|---------------------------------------------|
| `a-christmas-carol`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `a-connecticut-yankee-in-king-arthurs-court`            | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `a-room-of-ones-own`                                    | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `a-study-in-scarlet`                                    | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `a-tale-of-two-cities`                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `adventures-of-huckleberry-finn`                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `alices-adventures-in-wonderland`                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `and-then-there-were-none`                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `animal-farm`                                           | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `anna-karenina`                                         | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `around-the-world-in-eighty-days`                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `catriona`                                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `crime-and-punishment`                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `david-copperfield`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `de-profundis`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `death-on-the-nile`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `demons`                                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `down-and-out-in-paris-and-london`                      | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `dracula`                                               | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `draculas-guest`                                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `emma`                                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `frankenstein`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `from-the-earth-to-the-moon`                            | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `gondal-poems`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `great-expectations`                                    | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `homage-to-catalonia`                                   | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `jane-eyre`                                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `journey-to-the-center-of-the-earth`                    | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `kidnapped`                                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `life-on-the-mississippi`                               | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `lives-of-the-most-eminent-literary-and-scientific-men` | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `lodore`                                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `mansfield-park`                                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `mathilda`                                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `meditations`                                           | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `mrs-dalloway`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `murder-on-the-orient-express`                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `nineteen-eighty-four`                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `no-coward-soul-is-mine`                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `notes-from-underground`                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `oliver-twist`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `on-the-origin-of-species`                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `orlando`                                               | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `personal-recollections-of-joan-of-arc`                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `persuasion`                                            | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `phantasmagoria-and-other-poems`                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `poems-by-currer-ellis-and-acton-bell`                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `poor-richards-almanack`                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `pride-and-prejudice`                                   | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `resurrection`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `roger-fry-a-biography`                                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `salome`                                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `selected-poems-of-emily-bronte`                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `sense-and-sensibility`                                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `shirley`                                               | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `strange-case-of-dr-jekyll-and-mr-hyde`                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `sylvie-and-bruno`                                      | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-abc-murders`                                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-adventures-of-sherlock-holmes`                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-adventures-of-tom-sawyer`                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-art-of-war`                                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-autobiography-of-benjamin-franklin`                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-black-arrow`                                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-black-tulip`                                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-brothers-karamazov`                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-children-of-captain-grant`                         | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-complete-poems-of-emily-bronte`                    | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-cossacks`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-count-of-monte-cristo`                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-death-of-ivan-ilyich`                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-first-men-in-the-moon`                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-happy-prince-and-other-tales`                      | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-hound-of-the-baskervilles`                         | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-hunting-of-the-snark`                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-idiot`                                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-importance-of-being-earnest`                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-invisible-man`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-island-of-doctor-moreau`                           | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-jewel-of-seven-stars`                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-lady-of-the-shroud`                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-lair-of-the-white-worm`                            | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-last-man`                                          | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-lost-world`                                        | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-murder-at-the-vicarage`                            | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-mysterious-island`                                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-picture-of-dorian-gray`                            | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-prince-and-the-pauper`                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-professor`                                         | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-road-to-wigan-pier`                                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-sign-of-the-four`                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-thoughts-of-marcus-aurelius`                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-three-musketeers`                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-time-machine`                                      | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-vicomte-of-bragelonne`                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-war-of-the-worlds`                                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-waves`                                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-way-to-wealth`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `the-wealth-of-nations`                                 | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `through-the-looking-glass`                             | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `to-the-lighthouse`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `tom-sawyer-abroad`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `tom-sawyer-detective`                                  | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `treasure-island`                                       | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `twenty-thousand-leagues-under-the-seas`                | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `twenty-years-after`                                    | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `valperga`                                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `villette`                                              | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `war-and-peace`                                         | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |
| `wuthering-heights`                                     | Unknown                  | None      | Not applicable  | Unverified legacy metadata; CSS placeholder |

## Cover Policy

No seeded record has a cover URL, and no cover asset is distributed. The UI uses its CSS
placeholder.
Before adding a seed cover, replace that record's empty cover fields with a metadata source, the
exact cover URL, a reviewed permitted-use basis, and limitations. Missing or unverified permission
means the URL stays absent. Public accessibility alone is not permission. Cover URLs remain
metadata:
no downloading, uploading, proxying, caching, or file storage is part of this workflow.
