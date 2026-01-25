CREATE TABLE IF NOT EXISTS users
(
    id          SERIAL PRIMARY KEY,
    login       TEXT NOT NULL UNIQUE,
    key         TEXT NOT NULL,
    accesslevel INTEGER 
);

CREATE TABLE IF NOT EXISTS tasks
(
    id          SERIAL PRIMARY KEY,
    header      TEXT NOT NULL,
    task        TEXT,
    developers  INTEGER[]
);