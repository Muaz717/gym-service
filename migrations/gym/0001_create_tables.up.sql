-- Таблица клиентов
CREATE TABLE person (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    phone VARCHAR(20) NOT NULL,
    UNIQUE (full_name, phone) -- чтобы нельзя было добавить дубль "Иван Иванов" + один и тот же телефон
);

-- Таблица абонементов (используем номер карты вместо id)
CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,       -- Номер абонемента с карты
    title TEXT NOT NULL,            -- Название тарифа
    price NUMERIC(10, 2) NOT NULL,   -- Цена тарифа
    duration_days INT NOT NULL,      -- Срок действия в днях
    freeze_days INT DEFAULT 0        -- Количество допустимых дней заморозки
);

-- Таблица подписок клиента на абонементы
CREATE TABLE person_subscriptions (
    number varchar(32) PRIMARY KEY,
    person_id BIGINT NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' -- active / frozen / completed
);