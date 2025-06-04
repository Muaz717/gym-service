CREATE TABLE IF NOT EXISTS single_visits (
    id SERIAL PRIMARY KEY,
    visit_date DATE NOT NULL,
    final_price NUMERIC(10,2) NOT NULL
);