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
    task        TEXT
);

CREATE TABLE IF NOT EXISTS tasks_users(
    user_id INTEGER NOT NULL,
    task_id INTEGER NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE, -- ON DELETE CASCADE - Если задача (task) удаляется, то автоматически удалить ВСЕ записи в tasks_users, которые ссылаются на эту задачу
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE, 

    PRIMARY KEY (user_id, task_id)
);
