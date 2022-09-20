CREATE TABLE documents (
    id serial PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title VARCHAR(300) NOT NULL,
    body TEXT
);

CREATE TABLE users (
    id serial PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    username VARCHAR(300) NOT NULL,
    email VARCHAR(300) NOT NULL
);
