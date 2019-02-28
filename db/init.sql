-- CREATE USER docker;
-- CREATE DATABASE docker;
-- GRANT ALL PRIVILEGES ON DATABASE docker TO docker;

\c docker;

CREATE TYPE currency AS ENUM ('USD', 'EUR');

CREATE TABLE account (
  user_id serial PRIMARY KEY,
  identifier VARCHAR(36) NOT NULL,
  currency currency NOT NULL,
  amount DECIMAL DEFAULT,
  created_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO account (identifier, currency) VALUES
('first', 'USD'),
('second', 'USD'),
('third', 'EUR');

GRANT ALL PRIVILEGES on TABLE account TO docker;
