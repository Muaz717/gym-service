ALTER TABLE subscription_freeze DROP CONSTRAINT fk_subscription;
ALTER TABLE subscription_freeze
    ADD CONSTRAINT fk_subscription
        FOREIGN KEY (subscription_number)
            REFERENCES person_subscriptions(number)
            ON DELETE CASCADE;