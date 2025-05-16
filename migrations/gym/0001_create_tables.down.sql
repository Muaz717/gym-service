-- Удаляем таблицу подписок клиента
DROP TABLE IF EXISTS client_subscriptions CASCADE;

-- Удаляем таблицу абонементов
DROP TABLE IF EXISTS subscriptions CASCADE;

-- Удаляем таблицу клиентов
DROP TABLE IF EXISTS person CASCADE;

DROP TABLE IF EXISTS person_subscriptions CASCADE