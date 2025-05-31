CREATE TABLE IF NOT EXISTS subscription_freeze (
    id SERIAL PRIMARY KEY,
    subscription_number VARCHAR(32) NOT NULL,
    freeze_start TIMESTAMP NOT NULL,
    freeze_end TIMESTAMP,
    days_used INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_subscription
       FOREIGN KEY (subscription_number) REFERENCES person_subscriptions(number)
);