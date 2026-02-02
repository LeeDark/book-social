# Domain model v0.1

## User/Role

- User: ID, FirstName, SecondName, SurName, Login, Password*, Email, UserRoleID
- Role: ID, RoleName, IsAdmin

## Book/Author/Genre; Covers

- Book: ID, Title, Description, BookAuthorID*, BookGenreID*
- LATER: BookAuthor: BookID, AuthorID
- LATER: BookGenre: BookID, GenreID
- Author: ID, FirstName, SecondName, SurName, Description
- Genre: ID, Name, Description
- Cover: ?

## Library/Shelf/Tag

- Library: ID, LibraryShelfID, LibraryBookID, LibraryTagID*
- LATER: LibraryTag: LibraryID, TagID
- Shelf: ID, Name, Description
- Tag: ID, Name, Description
