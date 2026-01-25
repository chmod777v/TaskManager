INSERT INTO users (id, login, key, accesslevel)
VALUES (1, 'Kosty', 'XKsE-mtMnK9NJ5J+%oIrVpQ&1', 2)
ON CONFLICT DO NOTHING;

INSERT INTO tasks (id, header, task, developers)
VALUES (1, 'Сервис для авторизации', 'Разработать сервис для авторизации. Срок 3 недели. Использовать postgres...', ARRAY[1,2,3])
ON CONFLICT DO NOTHING;