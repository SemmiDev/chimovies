CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    year integer NOT NULL,
    runtime integer NOT NULL,
    genres text[] NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE movies ADD CONSTRAINT movies_runtime_check CHECK (runtime>=0);
ALTER TABLE movies ADD CONSTRAINT movies_year_check CHECK (year BETWEEN 1888 and date_part('year',now()));
ALTER TABLE movies ADD CONSTRAINT genres_length_check CHECK (array_length(genres,1) BETWEEN 1 AND 5);

CREATE INDEX IF NOT EXISTS movie_title_idx ON movies USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS movie_genres_idx ON movies USING GIN (genres);

CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email text UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens (
	hash bytea PRIMARY KEY,
	user_id bigint NOT NULL REFERENCES users on DELETE CASCADE,
	expiry timestamp(0) with time zone NOT NULL,
	scope text NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code) VALUES('movies:read'),('movies:write');