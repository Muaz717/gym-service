ALTER TABLE person_subscriptions
    ADD COLUMN subscription_price NUMERIC NOT NULL DEFAULT 0, -- цена на момент оформления
    ADD COLUMN discount NUMERIC NOT NULL DEFAULT 0,           -- скидка в рублях
    ADD COLUMN final_price NUMERIC NOT NULL DEFAULT 0;        -- цена с учетом скидки