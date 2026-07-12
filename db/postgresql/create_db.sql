-- create_db.sql
-- Run as a PostgreSQL superuser:
-- psql -v ON_ERROR_STOP=1 -f db/postgresql/create_db.sql

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_catalog.pg_roles
        WHERE rolname = 'book_social'
    ) THEN
        CREATE ROLE book_social WITH LOGIN PASSWORD 'pa55word';
    ELSE
        ALTER ROLE book_social WITH LOGIN PASSWORD 'pa55word';
    END IF;
END
$$;

SELECT 'CREATE DATABASE book_social OWNER book_social ENCODING ''UTF8'''
WHERE NOT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_database
    WHERE datname = 'book_social'
)\gexec

ALTER DATABASE book_social OWNER TO book_social;

\connect book_social

ALTER SCHEMA public OWNER TO book_social;
GRANT ALL ON SCHEMA public TO book_social;
